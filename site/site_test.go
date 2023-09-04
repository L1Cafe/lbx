package site

import (
	"fmt"
	"net/url"
	"testing"
	"time"
)

func TestIsUrlHealthy(t *testing.T) {
	u, _ := url.Parse("https://google.com")
	err := isUrlHealthy(*u)
	if err != nil {
		t.Error("Error querying Google, are you connected to the Internet?")
	}
	err = nil
	u, _ = url.Parse("https://localhost:9952/doesnot/exist")
	err = isUrlHealthy(*u)
	if err == nil {
		t.Error("Testing for a non-existant site found a working site")
	}
}

func TestInit(t *testing.T) {
	u, _ := url.Parse("https://localhost:16311")
	var e []url.URL
	e = append(e, *u)
	s := newSite("testing_site", e, 1*time.Second, "", "/", 4000)
	time.Sleep(11 * time.Second)
	fmt.Printf("%v", s)

}
