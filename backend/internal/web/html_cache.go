//go:build embed

package web

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

// cacheEntry represents a single cached HTML page for a specific domain
type cacheEntry struct {
	content []byte
	etag    string
}

// HTMLCache manages the cached index.html with injected settings.
// Supports multi-tenant caching keyed by domain name.
// The default (main) site uses "" as the key.
type HTMLCache struct {
	mu              sync.RWMutex
	entries         map[string]*cacheEntry // domain -> cache entry ("" = default)
	baseHTMLHash    string                 // Hash of the original index.html (immutable after build)
	settingsVersion uint64                 // Incremented when settings change
}

// CachedHTML represents the cache state
type CachedHTML struct {
	Content []byte
	ETag    string
}

// NewHTMLCache creates a new HTML cache instance
func NewHTMLCache() *HTMLCache {
	return &HTMLCache{
		entries: make(map[string]*cacheEntry),
	}
}

// SetBaseHTML initializes the cache with the base HTML template
func (c *HTMLCache) SetBaseHTML(baseHTML []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hash := sha256.Sum256(baseHTML)
	c.baseHTMLHash = hex.EncodeToString(hash[:8]) // First 8 bytes for brevity
}

// Invalidate marks all cache entries as stale (e.g. when global settings change)
func (c *HTMLCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.settingsVersion++
	c.entries = make(map[string]*cacheEntry)
}

// InvalidateDomain invalidates a specific domain's cache entry
func (c *HTMLCache) InvalidateDomain(domain string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, domain)
}

// Get returns the cached HTML for a domain, or nil if cache is stale.
// Use "" for the default (main) site.
func (c *HTMLCache) Get(domain string) *CachedHTML {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry := c.entries[domain]
	if entry == nil {
		return nil
	}
	return &CachedHTML{
		Content: entry.content,
		ETag:    entry.etag,
	}
}

// Set updates the cache with new rendered HTML for a specific domain.
// Use "" for the default (main) site.
func (c *HTMLCache) Set(domain string, html []byte, settingsJSON []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[domain] = &cacheEntry{
		content: html,
		etag:    c.generateETag(settingsJSON),
	}
}

// generateETag creates an ETag from base HTML hash + settings hash
func (c *HTMLCache) generateETag(settingsJSON []byte) string {
	settingsHash := sha256.Sum256(settingsJSON)
	return `"` + c.baseHTMLHash + "-" + hex.EncodeToString(settingsHash[:8]) + `"`
}
