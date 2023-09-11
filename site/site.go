package site

import (
	"context"
	"errors"
	"fmt"
	"github.com/L1Cafe/lbx/config"
	"github.com/L1Cafe/lbx/log"
	"github.com/go-chi/chi/v5"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type pathSiteMap map[string]*site

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

// Global variables
// sites contains the data structures required to keep track of thread-safe access and more
var sites map[string]*site

// portMap contains the following:
// - ports as keys
// - a map of path:site as values
// This determines on what port each site is serving data. Also, what path is each site responsible for.
// Each path must only be claimed by one site at a time. If more than one site claim a path on the same port, the application quits with an error message.
var portMap map[uint16]pathSiteMap

// gracefulShutdown receives a "true" when goroutines need to stop running. Goroutines must start cleanup immediately, and exit as soon as possible.
var gracefulShutdownChannel chan bool

// runningGoroutines is used to block the gracefulShutdown function until all goroutines have finished
var runningGoroutines sync.WaitGroup

// runningHttpServers holds pointers to servers to allow graceful termination
var runningHttpServers []*http.Server

// signalChannel receives OS signals
var signalChannel chan os.Signal

// Making sure things are only initialised once
var once sync.Once

// running is initialised by Init to true, and then set to false by the gracefulShutdown function
var running atomic.Bool

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

// gracefulShutdownSignalHandler simply waits to receive a message
func signalHandler() {
	s, ok := <-signalChannel
	if !ok { // Channel was closed
		return
	}
	log.Wrapper(log.Info, fmt.Sprintf("%v received", s))
	Stop()
	return
}

// Stop  blocks until the shutdown is complete
func Stop() {
	log.Wrapper(log.Info, "Performing graceful shutdown...")
	close(gracefulShutdownChannel)
	close(signalChannel)
	// Stop each running server
	for _, e := range runningHttpServers {
		e.Shutdown(context.Background())
	}
	runningGoroutines.Wait()
	log.Wrapper(log.Info, "All healthchecks stopped")
	log.Wrapper(log.Info, "All HTTP servers stopped")
	sites = nil
	portMap = nil
	gracefulShutdownChannel = nil
	runningGoroutines = sync.WaitGroup{}
	runningHttpServers = nil
	signalChannel = nil
	running.Store(false)
	log.Wrapper(log.Info, "Application stopped")
	return
}

func Init(conf *config.ParsedConfig) {
	if running.Load() {
		log.Wrapper(log.Info, "Application was already running, stopping...")
		Stop()
	} else {
		running.Store(true)
		log.Wrapper(log.Info, "Starting application...")
	}
	gracefulShutdownChannel = make(chan bool)
	signalChannel = make(chan os.Signal)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	sites = map[string]*site{}
	portMap = map[uint16]pathSiteMap{}
	go signalHandler() // FIXME is this really the way to do this?
	// Step 1: Read the config
	for siteName, siteValue := range conf.Sites {
		ns := newSite(siteName, siteValue.Endpoints, siteValue.RefreshPeriod, siteValue.Domain, siteValue.Path, siteValue.Port)
		// Step 2: Add the sites to the sites map
		sites[siteName] = ns
		// Step 3: Add the port -> path:site to the portMap map
		// Step 3.1: Check if port already exists
		v, portExists := portMap[siteValue.Port]
		// Step 3.2: If it does, check if path already exists
		if portExists {
			s, pathExists := v[siteValue.Path]
			if pathExists {
				// A path cannot be served by two sites under the same port
				log.Wrapper(log.Fatal, fmt.Sprintf("Path %s defined twice or more for port %d.\nConflicting sites:\n\t%s\n\t%s", siteValue.Path, siteValue.Port, s.name, siteName))
			} else {
				// Step 3.2.2: If path doesn't exist, add it
				portMap[siteValue.Port][siteValue.Path] = ns
			}
		} else {
			// Step 3.3: If port doesn't exist, add it
			psm := map[string]*site{}
			psm[siteValue.Path] = ns
			portMap[siteValue.Port] = psm
		}
		// Step 4: Dispatch healthchecks
		log.Wrapper(log.Info, fmt.Sprintf("Dispatching health checks for site %s", siteName))
		go ns.autoHealthCheck()
	}
	for port := range portMap {
		log.Wrapper(log.Info, fmt.Sprintf("Starting server for port %d", port))
		// Step 5: Start servers
		go startServer(port)
	}
	log.Wrapper(log.Info, "The application is ready.")
}

func siteHandler(site *site) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		endpoint, eErr := site.getRandomHealthyEndpoint()
		if eErr != nil {
			log.Wrapper(log.Warn, fmt.Sprintf("%s", eErr.Error()))
			http.Error(w, "Error encountered when attempting to connect to upstream server, see server logs for details", http.StatusServiceUnavailable)
			return
		}
		rPath := r.URL.Path
		endpointR, eRErr := http.Get(endpoint.String() + rPath)
		if eRErr != nil {
			log.Wrapper(log.Warn, fmt.Sprintf("%s", eRErr.Error()))
			http.Error(w, "Error encountered when attempting to connect to upstream server, see server logs for details", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(endpointR.StatusCode)
		for key, values := range endpointR.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		defer endpointR.Body.Close()
		if _, err := io.Copy(w, endpointR.Body); err != nil {
			log.Wrapper(log.Warn, fmt.Sprintf("Failed to write response body: %s", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		log.Wrapper(log.Info, fmt.Sprintf("Request for site %s, path %s, client %v, served via %v", site.name, r.URL, r.RemoteAddr, endpoint.Host))
	}
}

func startServer(port uint16) {
	runningGoroutines.Add(1)
	pathSite := portMap[port]
	chiRouter := chi.NewRouter()
	for p, s := range pathSite {
		chiRouter.Get(p, siteHandler(s))
	}
	portString := strconv.Itoa(int(port))
	srv := http.Server{
		Addr:    ":" + portString,
		Handler: chiRouter,
	}
	runningHttpServers = append(runningHttpServers, &srv)
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Wrapper(log.Fatal, fmt.Sprintf("Error starting server on port %d: %s", port, err))
	} else if errors.Is(err, http.ErrServerClosed) {
		log.Wrapper(log.Info, fmt.Sprintf("Graceful shutdown requested, terminating server on port %d...", port))
		defer runningGoroutines.Done()
		return
	}
}

// autoHealthCheck periodically checks the endpoints in the provided site name. Intended to be run as goroutine.
func (s *site) autoHealthCheck() {
	runningGoroutines.Add(1)
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
		select {

		case <-gracefulShutdownChannel:
			// Channel was closed
			log.Wrapper(log.Info, fmt.Sprintf("Graceful shutdown requested, terminating %v healthchecks...", s.name))
			defer runningGoroutines.Done()
			return
		case <-time.After(s.refreshPeriod):
			// Do nothing, just wait		return
		}
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

func (s *site) getRandomHealthyEndpoint() (url.URL, error) {
	// Choosing server at random, currently the only load balancing algorithm
	s.healthyEndpoints.mutex.RLock()
	defer s.healthyEndpoints.mutex.RUnlock()
	hEL := *s.healthyEndpoints.endpoints
	if len(hEL) < 1 {
		return url.URL{}, errors.New(fmt.Sprintf("No healthy endpoints available for site %s", s.name))
	}
	index := rand.Intn(len(*s.healthyEndpoints.endpoints))
	return (*s.healthyEndpoints.endpoints)[index], nil
}
