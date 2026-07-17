package url

import (
	"crypto/sha1"
	"encoding/binary"
)

type RulesEngine struct {
	abHasher ABTestHasher
}

func NewRulesEngine(abHasher ABTestHasher) *RulesEngine {
	return &RulesEngine{abHasher: abHasher}
}

func (e *RulesEngine) ResolveDestination(u *URL, reqCtx RequestContext) string {
	for _, rule := range u.RoutingConfig.DeviceRules {
		if rule.DeviceType == reqCtx.DeviceType {
			return rule.Destination
		}
	}

	for _, rule := range u.RoutingConfig.GeoRules {
		if rule.CountryCode == reqCtx.CountryCode {
			return rule.Destination
		}
	}

	if len(u.RoutingConfig.Variants) > 0 {
		totalWeight := 0
		for _, v := range u.RoutingConfig.Variants {
			totalWeight += v.Weight
		}

		if totalWeight > 0 {
			bucketKey := reqCtx.IPAddress + string(u.Slug)
			bucket := e.abHasher.Bucket(bucketKey, totalWeight)

			runningWeight := 0
			for _, v := range u.RoutingConfig.Variants {
				runningWeight += v.Weight
				if bucket < runningWeight {
					return v.Destination
				}
			}
		}
	}

	return string(u.OriginalURL)
}

type SHA1ABTestHasher struct{}

func (h *SHA1ABTestHasher) Bucket(key string, buckets int) int {
	hash := sha1.Sum([]byte(key))
	val := binary.BigEndian.Uint32(hash[:4])
	return int(val % uint32(buckets))
}
