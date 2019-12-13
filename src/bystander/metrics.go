package bystander

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
)

// icantbelieveitsnotprometheus is a metrics registry
// WTF? Why wasn't the prometheus library used?
// Because: I couldn't register a guage with arbitrary labels
// e.g. it was not possible to create these two metrics:
// check_result{name="foo"} and check_result{name="bar", env="prod"}
// the registry code requires the number of labels to be consistent
type icantbelieveitsnotprometheus struct {
	m map[string]float64
	l *sync.Mutex
}

func newicantbelieveitsnotprometheus() *icantbelieveitsnotprometheus {
	return &icantbelieveitsnotprometheus{
		m: map[string]float64{},
		l: &sync.Mutex{},
	}
}

func getMetricName(name string, tags map[string]string) string {
	res := []string{}
	keys := []string{}
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		res = append(res, fmt.Sprintf("%s=\"%s\"", k, tags[k]))
	}
	return name + "{" + strings.Join(res, ",") + "}"
}

func (icbinp *icantbelieveitsnotprometheus) setCheckResult(tags map[string]string, ok bool) {
	s := getMetricName("check_status", tags)

	icbinp.l.Lock()
	defer icbinp.l.Unlock()

	if ok {
		icbinp.m[s] = 1.0
	} else {
		icbinp.m[s] = 0.0
	}
}

func (icbinp *icantbelieveitsnotprometheus) handler(w http.ResponseWriter, r *http.Request) {
	icbinp.l.Lock()
	defer icbinp.l.Unlock()
	keys := []string{}
	for k := range icbinp.m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(w, "%s %f\n", k, icbinp.m[k])
	}
}
