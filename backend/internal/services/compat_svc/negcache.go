package compat_svc

import (
	"errors"
	"sync"
	"time"

	"github.com/leocov-dev/packwiz-nxt/core"
)

const (
	// fetchFailureTTL is how long a failed download (dead URL, HTTP error,
	// timeout, oversized jar) is not tried again.
	fetchFailureTTL = time.Hour
	// hashMismatchTTL is how long a URL that served other bytes than the pack
	// recorded is not tried again. Longer: such a URL rarely fixes itself.
	hashMismatchTTL = 24 * time.Hour
	// negativeCacheMax caps the entries kept; a pack with more dead URLs than
	// this only loses the oldest entries early.
	negativeCacheMax = 4096
)

// failureTTL is how long a download error is remembered. Context errors say
// nothing about the URL and are never remembered (ok false).
func failureTTL(err error) (ttl time.Duration, ok bool) {
	switch {
	case err == nil, isContextErr(err):
		return 0, false
	case errors.Is(err, ErrHashMismatch):
		return hashMismatchTTL, true
	default:
		return fetchFailureTTL, true
	}
}

// failureKey identifies one download attempt: the same jar hash can sit behind
// several URLs, and only the URL that failed may be skipped.
type failureKey struct {
	Mode, URL, HashFormat, Hash string
}

func failureKeyOf(m *core.Mod) failureKey {
	d := m.Download
	return failureKey{Mode: d.Mode, URL: d.URL, HashFormat: d.HashFormat, Hash: d.Hash}
}

type failureEntry struct {
	err   error
	until time.Time
}

// negativeCache remembers recent download failures so a pack with many dead
// URLs does not spend a full timeout on each of them in every scan. It is in
// memory, keyed by (mode, URL, hash), not stored in mod_jar_meta: that table
// is keyed by the jar hash alone, so a failure recorded there for the hash a
// pack expects would also block other packs that reach the same jar through
// a working URL. The cost is that a restart forgets the failures, which only
// repeats the downloads once.
type negativeCache struct {
	mu      sync.Mutex
	entries map[failureKey]failureEntry
}

func newNegativeCache() *negativeCache {
	return &negativeCache{entries: make(map[failureKey]failureEntry)}
}

// get returns the remembered error of key if it has not expired.
func (c *negativeCache) get(key failureKey, now time.Time) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		return false, nil
	}
	if !now.Before(e.until) {
		delete(c.entries, key)
		return false, nil
	}
	return true, e.err
}

// put remembers err for the TTL failureTTL gives it, if any.
func (c *negativeCache) put(key failureKey, err error, now time.Time) {
	ttl, ok := failureTTL(err)
	if !ok {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= negativeCacheMax {
		for k, e := range c.entries {
			if !now.Before(e.until) {
				delete(c.entries, k)
			}
		}
		if len(c.entries) >= negativeCacheMax {
			// all live: start over rather than grow without bound
			c.entries = make(map[failureKey]failureEntry)
		}
	}
	c.entries[key] = failureEntry{err: err, until: now.Add(ttl)}
}

// forget drops key, after a successful download.
func (c *negativeCache) forget(key failureKey) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}
