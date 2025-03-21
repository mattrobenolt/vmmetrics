package metrics

import (
	"errors"
	"fmt"
	"hash/maphash"
	"runtime"
	"strings"
	"sync"
	"weak"
)

var identCache sync.Map // map[uint64]weak.Pointer[string]

// makeIdentHandle is a specialized version of unique.Make[string]
// that also caches the fact that a given Ident is a valid identifier.
// Caching uniqueness at this point makes it significantly faster to
// repeatedly look up the same identifier.
func makeIdentHandle(value string) Ident {
	if len(value) == 0 {
		panic(errors.New("empty identifier"))
	}

	key := maphash.String(globalSeed, value)

	// Keep around any values we allocate for insertion. There
	// are a few different ways we can race with other threads
	// and create values that we might discard. By keeping
	// the first one we make around, we can avoid generating
	// more than one per racing thread.
	var (
		toInsert     *string // Keep this around to keep it alive.
		toInsertWeak weak.Pointer[string]
	)
	var ptr *string
	for {
		// Check the map.
		wp, ok := identCache.Load(key)
		if !ok {
			// Try to insert a new value into the map.
			if toInsert == nil {
				if !validateIdent(value) {
					panic(fmt.Errorf("invalid identifier: %q", value))
				}
				toInsert = new(string)
				*toInsert = strings.Clone(value)
				toInsertWeak = weak.Make(toInsert)
			}
			wp, _ = identCache.LoadOrStore(key, toInsertWeak)
		}
		// Now that we're sure there's a value in the map, let's
		// try to get the pointer we need out of it.
		ptr = wp.(weak.Pointer[string]).Value()
		if ptr != nil {
			break
		}
		// The weak pointer is nil, so the old value is truly dead.
		// Try to remove it and start over.
		identCache.CompareAndDelete(key, wp)
	}
	runtime.KeepAlive(toInsert)
	return Ident{ptr}
}
