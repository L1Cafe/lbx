package site

import (
	"errors"
	"fmt"
	"github.com/L1Cafe/lbx/config"
	"github.com/L1Cafe/lbx/log"
	"math/rand"
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
			// TODO actually check endpoint
		}
		s.healthyEndpoints.mutex.Lock()
		// swap list of previously healthy endpoints with the list of currently healthy ones
		s.healthyEndpoints.endpoints = currentHealthyEndpoints
		s.healthyEndpoints.mutex.Unlock()
		log.Wrapper(log.Info, fmt.Sprintf("Healthy endpoint list of site %s updated", siteKey))
		time.Sleep(s.refreshPeriod)
	}
}

// healthCheck is a manual endpoint check that is triggered when an endpoint misbehaves, and will evict the endpoint if
// it's unhealthy
func healthCheck(siteKey string, u url.URL) {
	s := sites[siteKey]
	endpointList := s.endpoints
	for _, endpointUrl := range endpointList {
		if endpointUrl == u {
			// TODO check u
			// If u healthy, make sure it's in the healthy list
			// If u unhealthy, make a new list, lock the old healthy one for read, if it's there take it out, and then swap
		}
	}
}

func GetRandomEndpoint(siteKey string) (url.URL, error) {
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
