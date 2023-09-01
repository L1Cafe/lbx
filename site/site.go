package site

import (
	"errors"
	"fmt"
	"github.com/L1Cafe/lbx/config"
	"github.com/L1Cafe/lbx/log"
	"math/rand"
	"net/http"
	"net/url"
	"slices"
	"sync"
	"time"
)

// This package doesn't contain validation because the validation should be done at configuration parsing time

var sites map[string]*site

// healthyEndpoints is a thread-safe mutating structure that holds a list of the endpoints
type healthyEndpoints struct {
	mutex     *sync.RWMutex
	endpoints *[]url.URL
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
	heE := new([]url.URL)
	he.mutex = heM
	he.endpoints = heE
	s.healthyEndpoints = he
	return s
}

func Init(s *config.ParsedConfig) {
	for k, v := range s.Sites {
		// TODO assign as necessary
		sites[k] = newSite(k, v.Endpoints, v.RefreshPeriod, v.Domain, v.Path, v.Port)
		log.Wrapper(log.Info, fmt.Sprintf("Loaded configuration for site %s", k))
	}
	for k := range sites {
		go autoHealthCheck(k)
		log.Wrapper(log.Info, fmt.Sprintf("Dispatched health checks for site %s", k))
	}
}

// autoHealthCheck periodically checks the endpoints in the provided site name. It's ok for this function to not have
// any validaton, because it's unexported.
func autoHealthCheck(siteKey string) {
	s := sites[siteKey]
	for {
		currentHealthyEndpoints := new([]url.URL)
		log.Wrapper(log.Info, fmt.Sprintf("Checking healthy endpoints for site %s", siteKey))
		// TODO: check health of every endpoint and make list of healthy endpoints
		for _, endpoint := range s.endpoints {
			log.Wrapper(log.Info, fmt.Sprintf(
				"Checking health status of endpoint %s for site %s", endpoint.String(), siteKey))
			err := isUrlHealthy(endpoint)
			if err == nil {
				*currentHealthyEndpoints = append(*currentHealthyEndpoints, endpoint)
			}
		}
		s.healthyEndpoints.mutex.Lock()
		// swap list of previously healthy endpoints with the list of currently healthy ones
		s.healthyEndpoints.endpoints = currentHealthyEndpoints
		s.healthyEndpoints.mutex.Unlock()
		log.Wrapper(log.Info, fmt.Sprintf("Healthy endpoint list of site %s was updated", siteKey))
		time.Sleep(s.refreshPeriod)
	}
}

// healthCheck is a manual endpoint check that returns an error on any reading error as well as 500 error codes
func isUrlHealthy(u url.URL) error {
	// TODO check u
	res, err := http.Head(u.String())
	if err != nil {
		return err
	}
	if res.StatusCode >= 500 {
		return fmt.Errorf("received %d status from %s", res.StatusCode, u.String())
	}
	return nil
}

func isEndpointListedHealthy(s string, u url.URL) bool {
	// Skipping validation, this is supposed to be a safe environment
	site := sites[s]
	return slices.Contains(site.endpoints, u)
}

func asyncSiteHealthCheck(s string) {
	
}

func asyncEndpointHealthCheck(s string, u url.URL) {
	err := isUrlHealthy(u)
	if err != nil {
		// TODO remove endpoint

	} else {
		// TODO add endpoint if not already there
	}
}

func queueSiteHealthCheck(s string) error {
	// TODO queue an async check of the entire site, replace the list of healthy endpoints
	return nil
}

func queueEndpointHealthCheck(s string, u url.URL) error {
	// TODO queue an async check of a single endpoint from the list of healthy endpoints, remove from list of healthy endpoints if necessary
	return nil
}

func markUnhealthy(s string, u url.URL) {
	site := sites[s]
	site.healthyEndpoints.mutex.Lock()
	for _, su := range *site.healthyEndpoints.endpoints {
		currentHealthyEndpoints := new([]url.URL)
		if su != u {
			*currentHealthyEndpoints = append(*currentHealthyEndpoints, su)
		} else {
			log.Wrapper(log.Info, fmt.Sprintf("Endpoint %s evicted from healthy endpoints list for site %s", u.String(), s))
		}
	}
	site.healthyEndpoints.mutex.Unlock()
}

func GetRandomHealthyEndpoint(siteKey string) (url.URL, error) {
	s, ok := sites[siteKey]
	if !ok {
		return url.URL{}, errors.New(fmt.Sprintf("Site %s not found", siteKey))
	}
	// Choosing server at random, currently the only load balancing algorithm
	s.healthyEndpoints.mutex.RLock()
	defer s.healthyEndpoints.mutex.RUnlock()
	index := rand.Intn(len(*s.healthyEndpoints.endpoints))
	return (*s.healthyEndpoints.endpoints)[index], nil
}
