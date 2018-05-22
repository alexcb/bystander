package bystander

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

type alertFunc func(id, checkName string, ok bool, details map[string]string)

// CheckStatus defines the status of a check run
type CheckStatus struct {
	ok       bool
	time     time.Time
	duration time.Duration
	details  map[string]string
}

type check struct {
	config   Check
	silenced bool
	status   []*CheckStatus

	lastAlerted       time.Time
	lastAlertedStatus bool
}

func (s *check) shouldAlert() bool {
	if s.silenced {
		return false
	}

	ok, numConsecutive := getConsecutiveStatus(s.status)

	if ok {
		if numConsecutive < s.config.CommonConfig().numSuccessBeforeRecovery {
			return false
		}
	} else {
		if numConsecutive < s.config.CommonConfig().numFailuresBeforeAlerting {
			return false
		}
	}

	if s.lastAlertedStatus == ok {
		if ok {
			return false
		}
		if time.Since(s.lastAlerted) < time.Hour {
			// only re-alert after an hour
			return false
		}
	}

	s.lastAlerted = time.Now().UTC()
	s.lastAlertedStatus = ok

	return true
}

func getNameFromTags(tags map[string]string) string {
	res := []string{}
	keys := []string{}
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		res = append(res, fmt.Sprintf("%s=%s", k, tags[k]))
	}
	name := tags["name"]
	return name + "{" + strings.Join(res, ", ") + "}"
}

func (s *check) name() string {
	return getNameFromTags(s.config.CommonConfig().tags)
}

func (s *check) text() string {
	if len(s.status) == 0 {
		return "never run"
	}
	return fmt.Sprintf("%v", s.status[0].details)
}

func (s *check) id() string {
	data := sha1.Sum([]byte(s.name()))
	return base64.StdEncoding.EncodeToString(data[:])
}

func truncateString(s string, n int) string {
	if len(s) > n {
		if n > 3 {
			n -= 3
		}
		s = s[0:n] + "..."
	}
	return s
}

func (s *check) run() *CheckStatus {
	now := time.Now().UTC()

	ok, details := s.config.Run()

	duration := time.Since(now)

	return &CheckStatus{
		ok:       ok,
		details:  details,
		duration: duration,
		time:     now,
	}
}

func (s *check) json() string {
	details := map[string]string{}
	var lastCheck time.Time
	var duration time.Duration

	ok, numConsecutive := getConsecutiveStatus(s.status)

	if len(s.status) > 0 {
		details = s.status[0].details
		duration = s.status[0].duration
		lastCheck = s.status[0].time
	}

	data, err := json.Marshal(&CheckResult{
		ID:             s.id(),
		Tags:           s.config.CommonConfig().tags,
		Details:        details,
		OK:             ok,
		NumConsecutive: numConsecutive,
		LastRun:        lastCheck,
		Duration:       duration.Seconds(),
		Silenced:       s.silenced,
	})
	if err != nil {
		panic(err)
	}
	return string(data)
}

type checkManager struct {
	checks     []*check
	lock       *sync.Mutex
	silencers  map[string]*silencer
	db         *bolt.DB
	maxHistory int
}

func newCheckManager(checkConfigs []Check, maxHistory int, db *bolt.DB) *checkManager {
	silencers, err := loadSilencers(db)
	if err != nil {
		panic(err)
	}
	checks := []*check{}
	for _, cc := range checkConfigs {
		checks = append(checks, &check{
			config:            cc,
			lastAlertedStatus: true,
		})
	}
	manager := &checkManager{
		checks:     checks,
		lock:       &sync.Mutex{},
		silencers:  silencers,
		db:         db,
		maxHistory: maxHistory,
	}
	manager.applySilencersUnsafe()
	return manager
}

func getConsecutiveStatus(statuses []*CheckStatus) (bool, int) {
	if len(statuses) == 0 {
		return true, 0
	}
	ok := statuses[0].ok
	num := 0
	for _, status := range statuses {
		if ok != status.ok {
			break
		}
		num++
	}
	return ok, num
}

func (s *checkManager) run(fn alertFunc) {
	for _, check := range s.checks {
		status := check.run()
		s.lock.Lock()

		if len(check.status) >= s.maxHistory {
			check.status = check.status[:(s.maxHistory - 1)]
		}

		check.status = append([]*CheckStatus{status}, check.status...)

		alertNeeded := check.shouldAlert()

		s.lock.Unlock()

		if alertNeeded {
			fn(check.id(), check.name(), status.ok, status.details)
		}
	}
}

func isSilenced(silencer, check map[string]string) bool {
	for k, v := range silencer {
		vv, ok := check[k]
		if !ok || vv != v {
			return false
		}
	}
	return true
}

func (s *checkManager) applySilencersUnsafe() {
	for _, check := range s.checks {
		check.silenced = false
		for _, silencer := range s.silencers {
			if isSilenced(silencer.Filters, check.config.CommonConfig().tags) {
				check.silenced = true
			}
		}
	}
}

// CheckResult is the results of a check
type CheckResult struct {
	ID             string            `json:"id"`
	Tags           map[string]string `json:"tags"`
	OK             bool              `json:"ok"`
	NumConsecutive int               `json:"num_consecutive"`
	Details        map[string]string `json:"details"`
	LastRun        time.Time         `json:"last_run"`
	Duration       float64           `json:"duration"`
	Silenced       bool              `json:"silenced"`
}

func (s *checkManager) getChecksJSON() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	res := []string{}
	for _, check := range s.checks {
		res = append(res, check.json())
	}
	return "[" + strings.Join(res, ", ") + "]"
}

func (s *checkManager) getSilencersJSON() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	res := []string{}
	for _, silencer := range s.silencers {
		res = append(res, silencer.json())
	}
	return "[" + strings.Join(res, ", ") + "]"
}

func (s *checkManager) updateSilencer(ss *silencer) {
	key, val := serializeSilencer(ss)

	s.lock.Lock()
	defer s.lock.Unlock()

	s.silencers[string(key)] = ss

	err := s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("silencers"))
		if err != nil {
			return err
		}
		return b.Put(key, val)
	})
	if err != nil {
		panic(err)
	}

	s.applySilencersUnsafe()
}

func (s *checkManager) removeSilencer(k string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.silencers[k]; !ok {
		return
	}

	delete(s.silencers, k)

	err := s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("silencers"))
		if err != nil {
			return err
		}
		return b.Delete([]byte(k))
	})
	if err != nil {
		panic(err)
	}

	s.applySilencersUnsafe()
}
