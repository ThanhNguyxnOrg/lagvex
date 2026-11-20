package client

import (
	"strings"
	"sync/atomic"
)

// buildVersion is the version reported by the dashboard footer. The client
// binary sets it at startup from its ldflags value; library users and tests
// see "dev" instead of a fabricated release number.
var buildVersion atomic.Pointer[string]

func init() {
	v := "dev"
	buildVersion.Store(&v)
}

// SetBuildVersion records the running client's version. Empty values are
// ignored so a build without ldflags keeps the honest "dev" default.
func SetBuildVersion(v string) {
	v = strings.TrimSpace(v)
	if v == "" {
		return
	}
	buildVersion.Store(&v)
}

// BuildVersion returns the version the dashboard should display.
func BuildVersion() string {
	if ptr := buildVersion.Load(); ptr != nil {
		return *ptr
	}
	return "dev"
}
