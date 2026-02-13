// Package geoip provides IP to country lookup using embedded IP2Location data.
package geoip

import (
	"log"
	"net/netip"
	"sync"

	"github.com/phuslu/iploc"
)

var (
	instance *Service
	once     sync.Once
)

// Service wraps the embedded IP geolocation lookup.
type Service struct{}

// Init initializes the global GeoIP service.
// The dbPath parameter is kept for backward compatibility but ignored
// since we now use embedded IP2Location data via phuslu/iploc.
func Init(dbPath string) {
	once.Do(func() {
		instance = &Service{}
		log.Printf("[GeoIP] Initialized with embedded IP2Location data")
	})
}

// Get returns the global GeoIP service instance.
func Get() *Service {
	return instance
}

// LookupCountry returns the ISO 3166-1 alpha-2 country code for the given IP.
// Returns empty string if the IP is invalid or lookup fails.
func (s *Service) LookupCountry(ipStr string) string {
	if s == nil {
		return ""
	}
	addr, err := netip.ParseAddr(ipStr)
	if err != nil {
		return ""
	}
	code := iploc.IPCountry(addr)
	if code == "" || code == "ZZ" {
		return ""
	}
	return code
}

// IsAvailable returns true if the GeoIP service is initialized.
// With embedded data, this is always true once Init is called.
func (s *Service) IsAvailable() bool {
	return s != nil
}
