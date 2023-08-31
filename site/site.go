package site

import (
	"github.com/L1Cafe/lbx/config"
	"net/url"
	"sync"
	"time"
)

// This package doesn't contain validation because the validation should be done at configuration parsing time

var sites map[string]*site

// healthyEndpoints is a thread-safe mutating structure that holds a list of the endpoints
type healthyEndpoints struct {
	mutex                *sync.RWMutex
	currentEndpointIndex *uint32
	endpoints            *[]url.URL
}

// site is a read-only structure that comes from the parameters defined in the configuration file
type site struct {
	name             string
	endpoints        []url.URL
	healthyEndpoints *healthyEndpoints
	refreshPeriod    time.Duration
	domain           string
	path             string
	port             uint16
}

func newSite(name string, endpoints []url.URL, refreshPeriod time.Duration, domain string, path string, port uint16) *site {
	s := new(site)
	s.name = name
	s.endpoints = endpoints
	s.refreshPeriod = refreshPeriod
	s.domain = domain
	s.path = path
	s.port = port
	he := new(healthyEndpoints)
	heM := new(sync.RWMutex)
	heCei := new(uint32)
	heE := new([]url.URL)
	he.mutex = heM
	he.currentEndpointIndex = heCei
	he.endpoints = heE
	s.healthyEndpoints = he
	return s
}

func Init(s *config.ParsedConfig) {
	for k, v := range s.Sites {
		// TODO assign as necessary
		sites[k] = newSite(k, v.Endpoints, v.RefreshPeriod, v.Domain, v.Path, v.Port)
	}
	for k, _ := range sites {
		go autoHealthCheck(k)
	}
}

func autoHealthCheck(siteKey string) {
	s := sites[siteKey]
	for {
		s.healthyEndpoints.mutex.RLock()
		currentHealthyEndpoints := new([]url.URL)
		// check health of every endpoint and make list of healthy endpoints
		s.healthyEndpoints.mutex.RUnlock()
		s.healthyEndpoints.mutex.Lock()
		// swap list of healthy endpoints with the current one
		s.healthyEndpoints.endpoints = currentHealthyEndpoints
		s.healthyEndpoints.mutex.Unlock()
		time.Sleep(s.refreshPeriod)
	}
}
