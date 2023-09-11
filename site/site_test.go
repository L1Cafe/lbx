package site

import (
	"net/url"
	"testing"
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
