package site

import (
	"errors"
	"fmt"
	"github.com/L1Cafe/lbx/config"
	"github.com/L1Cafe/lbx/log"
	"github.com/go-chi/chi/v5"
	"math/rand"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"sync"
	"time"
)

// This package doesn't contain validation because the validation should be done at configuration parsing time

var sites map[string]*site

// portMap contains the following:
// - ports as keys
// - a map of path:site
// This determines on what port each site is serving data. Also, what path is each site responsible for
// Each path should only be claimed by one site at a time
var portMap map[uint16]mapPathSite

type mapPathSite map[string]*site

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

func Init(conf *config.ParsedConfig) {
	// FIXME NEED TO RETHINK THIS
	// Step 1: Read the config
	// Step 2: Add the sites to the sites map
	// Step 3: Add the port -> path:site to the portMap map
	// Step 4: Dispatch healthchecks
	// Step 5: Start servers
}

func startServer(port uint16) {
	pathSite := portMap[port]
	// TODO finish actually serving the website
	chiRouter := chi.NewRouter()
	for p, _ := range pathSite {
		chiRouter.Get(p, func(w http.ResponseWriter, r *http.Request) {
			// TODO finish this
			w.Write([]byte("Hello, world!"))
		})
	}
	portString := strconv.Itoa(int(port))
	err := http.ListenAndServe(":"+portString, chiRouter)
	if err != nil {
		log.Wrapper(log.Fatal, fmt.Sprintf("Error trying to bind to port %d: %s", port, err))
	}
}

// autoHealthCheck periodically checks the endpoints in the provided site name.
func (s *site) autoHealthCheck() {
	for {
		currentHealthyEndpoints := new([]url.URL)
		log.Wrapper(log.Info, fmt.Sprintf("Checking healthy endpoints for site %s", s.name))
		for _, endpoint := range s.endpoints {
			log.Wrapper(log.Info, fmt.Sprintf(
				"Checking health status of endpoint %s for site %s", endpoint.String(), s.name))
			err := isUrlHealthy(endpoint)
			if err == nil {
				*currentHealthyEndpoints = append(*currentHealthyEndpoints, endpoint)
			}
		}
		s.healthyEndpoints.mutex.Lock()
		// swap list of previously healthy endpoints with the list of currently healthy ones
		s.healthyEndpoints.endpoints = currentHealthyEndpoints
		s.healthyEndpoints.mutex.Unlock()
		log.Wrapper(log.Info, fmt.Sprintf("Healthy endpoint list of site %s was updated", s.name))
		time.Sleep(s.refreshPeriod)
	}
}

// healthCheck is an endpoint check that returns an error on any reading error as well as 500 error codes
func isUrlHealthy(u url.URL) error {
	httpClient := http.Client{Timeout: 5 * time.Second}
	res, err := httpClient.Head(u.String())
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
	//TODO
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
