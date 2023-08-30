package site

import "github.com/L1Cafe/lbx/config"

var sites map[string]config.SiteParsedConfig

func Init(s *config.ParsedConfig) {
	sites = s.Sites
}

func autoHealthCheck(siteKey string) {
	//servers := sites[siteKey].Servers
	for {
		// TODO periodically check

	}
}
