package site

import (
	"github.com/L1Cafe/lbx/config"
	"net/url"
	"sync"
	"time"
)

//var sites map[string]config.SiteParsedConfig

var sites map[string]*site

type site struct {
	mutex         *sync.Mutex
	name          string
	endpoints     []url.URL
	refreshPeriod time.Duration
	domain        string
	path          string
	port          uint16
}

func Init(s *config.ParsedConfig) {
	for k, v := range s.Sites {
		newSite := new(site)
		v.RefreshPeriod = newSite.refreshPeriod
		v.Domain = newSite.domain
		v.Path = newSite.path
		v.Port = newSite.port
		v.Endpoints = newSite.endpoints
		// TODO assign as necessary
		sites[k] = newSite
	}
}

func autoHealthCheck(siteKey string) {
	//servers := sites[siteKey].Servers
	for {
		// TODO periodically check

	}
}
