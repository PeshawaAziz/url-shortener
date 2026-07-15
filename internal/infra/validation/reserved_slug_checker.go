package validation

import (
	"context"
	"sync"
)

var defaultReserved = []string{"api", "admin", "login", "auth", "www", "static", "assets", "healthz"}

type ConfigReservedChecker struct {
	mu        sync.RWMutex
	blocklist map[string]bool
}

func NewConfigReservedChecker(initialSlugs []string) *ConfigReservedChecker {
	bl := make(map[string]bool)
	for _, s := range initialSlugs {
		bl[s] = true
	}

	for _, s := range defaultReserved {
		bl[s] = true
	}

	return &ConfigReservedChecker{blocklist: bl}
}

func (c *ConfigReservedChecker) IsReserved(ctx context.Context, slug string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.blocklist[slug]
}

func (c *ConfigReservedChecker) HotReload(newSlugs []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	bl := make(map[string]bool)
	for _, s := range newSlugs {
		bl[s] = true
	}

	for _, s := range defaultReserved {
		bl[s] = true
	}
	c.blocklist = bl
}
