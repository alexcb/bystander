package bystander

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/ianschenck/envflag"
	yaml "gopkg.in/yaml.v2"
)

// Config holds global configuration for this app
type Config struct {
	ListenAddr   string
	WebAddress   string
	SlackWebHook string
	Checks       []Check
	DatabasePath string
	MaxHistory   int
}

var errMissingDB = fmt.Errorf("missing db path")

func getConfig() (*Config, error) {
	var ok bool
	var err error

	configPath := envflag.String("BYSTANDER_CONFIG", "/etc/bystander.conf", "path to config")
	envflag.Parse()

	fp, err := os.Open(*configPath)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	data, err := ioutil.ReadAll(fp)
	if err != nil {
		return nil, err
	}

	// A generic interface was used due to difficulties getting the tags into a map[string]string
	root := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(data), &root)
	if err != nil {
		return nil, err
	}

	config := Config{
		Checks: []Check{},
	}

	listenAddr, ok := root["listen_address"]
	if !ok {
		listenAddr = "localhost:8088"
	}
	config.ListenAddr = listenAddr.(string)

	if slackHook, ok := root["slack_webhook"]; ok {
		config.SlackWebHook = slackHook.(string)
	}

	if db, ok := root["db"]; ok {
		config.DatabasePath = db.(string)
	} else {
		return nil, errMissingDB
	}

	if x, ok := root["max_history"]; ok {
		config.MaxHistory, ok = x.(int)
		if !ok {
			panic("failed to parse max_history")
		}
	} else {
		config.MaxHistory = 10
	}

	if webAddr, ok := root["web_address"]; ok {
		config.WebAddress = webAddr.(string)
	} else {
		config.WebAddress = config.ListenAddr
	}

	seenChecks := map[string]bool{}

	checks, ok := root["checks"]
	if ok {
		for _, c := range checks.([]interface{}) {
			checkConfig := c.(map[interface{}]interface{})

			checkType, ok := checkConfig["type"]
			if !ok {
				panic("check config without type")
			}

			if _, ok := checkConfig["tags"]; !ok {
				panic("check config without tags")
			}

			numFailuresBeforeAlerting := 1
			if x, ok := checkConfig["num_failures_before_alerting"]; ok {
				numFailuresBeforeAlerting, ok = x.(int)
				if !ok {
					panic("unable to parse numFailuresBeforeAlerting -- value is not an integer")
				}
				if numFailuresBeforeAlerting > config.MaxHistory {
					panic("num_failures_before_alerting must be less than max_history")
				}
			}

			numSuccessBeforeRecovery := numFailuresBeforeAlerting
			if x, ok := checkConfig["num_success_before_recovery"]; ok {
				numSuccessBeforeRecovery, ok = x.(int)
				if !ok {
					panic("unable to parse numFailuresBeforeAlerting -- value is not an integer")
				}
				if numSuccessBeforeRecovery > config.MaxHistory {
					panic("num_failures_before_alerting must be less than max_history")
				}
			}

			tags := map[string]string{}
			for _, x := range checkConfig["tags"].([]interface{}) {
				xx := x.(map[interface{}]interface{})
				for k, v := range xx {
					tags[k.(string)] = v.(string)
				}
			}

			checkName := getNameFromTags(tags)
			if _, ok := seenChecks[checkName]; ok {
				panic(fmt.Sprintf("duplicate check tags: %s", checkName))
			}
			seenChecks[checkName] = true

			var check Check
			switch checkType {
			case "url":
				check = parseURLCheck(checkConfig)
			case "docker":
				check = parseDockerCheck(checkConfig)
			}
			check.CommonConfig().tags = tags
			check.CommonConfig().numFailuresBeforeAlerting = numFailuresBeforeAlerting
			check.CommonConfig().numSuccessBeforeRecovery = numSuccessBeforeRecovery
			config.Checks = append(config.Checks, check)
		}
	}

	return &config, nil
}

// Server implements the server that displays the status of the canary
type Server struct {
	config       *Config
	checkManager *checkManager
}

func (s *Server) checksJSON(w http.ResponseWriter, r *http.Request) {
	status := s.checkManager.getChecksJSON()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(status))
}

func (s *Server) silencersJSON(w http.ResponseWriter, r *http.Request) {
	status := s.checkManager.getSilencersJSON()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(status))
}

func (s *Server) addSilencers(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	silencer, err := silencerFromJSON(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.checkManager.updateSilencer(silencer)

	//w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) deleteSilencers(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	silencer, err := silencerFromJSON(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.checkManager.removeSilencer(silencer.key())

	//w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) serveInBackground(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	mux.HandleFunc("/checks.json", s.checksJSON)
	mux.HandleFunc("/silencers.json", s.silencersJSON)
	mux.HandleFunc("/add-silencer", s.addSilencers)
	mux.HandleFunc("/delete-silencer", s.deleteSilencers)
	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			//log.Err(err).Panic("error listening on HTTP status server")
			panic(err)
		}
	}()
}

// Run runs the canary
func Run() {
	config, err := getConfig()
	if err != nil {
		panic(err)
	}

	boltdb, err := bolt.Open(config.DatabasePath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		panic(err)
	}

	checkManager := newCheckManager(config.Checks, config.MaxHistory, boltdb)

	server := Server{
		checkManager: checkManager,
	}

	server.serveInBackground(config.ListenAddr)

	alerter := newSlackAlerter(config.SlackWebHook, config.WebAddress)

	for {
		checkManager.run(alerter.alert)
		time.Sleep(time.Minute)
	}
}
