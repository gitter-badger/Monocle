// Code generated by github.com/vektah/dataloaden, DO NOT EDIT.

package generated

import (
	"sync"
	"time"

	"github.com/ddouglas/monocle"
)

// CorporationAllianceHistoryLoaderConfig captures the config to create a new CorporationAllianceHistoryLoader
type CorporationAllianceHistoryLoaderConfig struct {
	// Fetch is a method that provides the data for the loader
	Fetch func(keys []uint) ([][]*monocle.CorporationAllianceHistory, []error)

	// Wait is how long wait before sending a batch
	Wait time.Duration

	// MaxBatch will limit the maximum number of keys to send in one batch, 0 = not limit
	MaxBatch int
}

// NewCorporationAllianceHistoryLoader creates a new CorporationAllianceHistoryLoader given a fetch, wait, and maxBatch
func NewCorporationAllianceHistoryLoader(config CorporationAllianceHistoryLoaderConfig) *CorporationAllianceHistoryLoader {
	return &CorporationAllianceHistoryLoader{
		fetch:    config.Fetch,
		wait:     config.Wait,
		maxBatch: config.MaxBatch,
	}
}

// CorporationAllianceHistoryLoader batches and caches requests
type CorporationAllianceHistoryLoader struct {
	// this method provides the data for the loader
	fetch func(keys []uint) ([][]*monocle.CorporationAllianceHistory, []error)

	// how long to done before sending a batch
	wait time.Duration

	// this will limit the maximum number of keys to send in one batch, 0 = no limit
	maxBatch int

	// INTERNAL

	// lazily created cache
	cache map[uint][]*monocle.CorporationAllianceHistory

	// the current batch. keys will continue to be collected until timeout is hit,
	// then everything will be sent to the fetch method and out to the listeners
	batch *corporationAllianceHistoryLoaderBatch

	// mutex to prevent races
	mu sync.Mutex
}

type corporationAllianceHistoryLoaderBatch struct {
	keys    []uint
	data    [][]*monocle.CorporationAllianceHistory
	error   []error
	closing bool
	done    chan struct{}
}

// Load a CorporationAllianceHistory by key, batching and caching will be applied automatically
func (l *CorporationAllianceHistoryLoader) Load(key uint) ([]*monocle.CorporationAllianceHistory, error) {
	return l.LoadThunk(key)()
}

// LoadThunk returns a function that when called will block waiting for a CorporationAllianceHistory.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *CorporationAllianceHistoryLoader) LoadThunk(key uint) func() ([]*monocle.CorporationAllianceHistory, error) {
	l.mu.Lock()
	if it, ok := l.cache[key]; ok {
		l.mu.Unlock()
		return func() ([]*monocle.CorporationAllianceHistory, error) {
			return it, nil
		}
	}
	if l.batch == nil {
		l.batch = &corporationAllianceHistoryLoaderBatch{done: make(chan struct{})}
	}
	batch := l.batch
	pos := batch.keyIndex(l, key)
	l.mu.Unlock()

	return func() ([]*monocle.CorporationAllianceHistory, error) {
		<-batch.done

		var data []*monocle.CorporationAllianceHistory
		if pos < len(batch.data) {
			data = batch.data[pos]
		}

		var err error
		// its convenient to be able to return a single error for everything
		if len(batch.error) == 1 {
			err = batch.error[0]
		} else if batch.error != nil {
			err = batch.error[pos]
		}

		if err == nil {
			l.mu.Lock()
			l.unsafeSet(key, data)
			l.mu.Unlock()
		}

		return data, err
	}
}

// LoadAll fetches many keys at once. It will be broken into appropriate sized
// sub batches depending on how the loader is configured
func (l *CorporationAllianceHistoryLoader) LoadAll(keys []uint) ([][]*monocle.CorporationAllianceHistory, []error) {
	results := make([]func() ([]*monocle.CorporationAllianceHistory, error), len(keys))

	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}

	corporationAllianceHistorys := make([][]*monocle.CorporationAllianceHistory, len(keys))
	errors := make([]error, len(keys))
	for i, thunk := range results {
		corporationAllianceHistorys[i], errors[i] = thunk()
	}
	return corporationAllianceHistorys, errors
}

// LoadAllThunk returns a function that when called will block waiting for a CorporationAllianceHistorys.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *CorporationAllianceHistoryLoader) LoadAllThunk(keys []uint) func() ([][]*monocle.CorporationAllianceHistory, []error) {
	results := make([]func() ([]*monocle.CorporationAllianceHistory, error), len(keys))
	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}
	return func() ([][]*monocle.CorporationAllianceHistory, []error) {
		corporationAllianceHistorys := make([][]*monocle.CorporationAllianceHistory, len(keys))
		errors := make([]error, len(keys))
		for i, thunk := range results {
			corporationAllianceHistorys[i], errors[i] = thunk()
		}
		return corporationAllianceHistorys, errors
	}
}

// Prime the cache with the provided key and value. If the key already exists, no change is made
// and false is returned.
// (To forcefully prime the cache, clear the key first with loader.clear(key).prime(key, value).)
func (l *CorporationAllianceHistoryLoader) Prime(key uint, value []*monocle.CorporationAllianceHistory) bool {
	l.mu.Lock()
	var found bool
	if _, found = l.cache[key]; !found {
		// make a copy when writing to the cache, its easy to pass a pointer in from a loop var
		// and end up with the whole cache pointing to the same value.
		cpy := make([]*monocle.CorporationAllianceHistory, len(value))
		copy(cpy, value)
		l.unsafeSet(key, cpy)
	}
	l.mu.Unlock()
	return !found
}

// Clear the value at key from the cache, if it exists
func (l *CorporationAllianceHistoryLoader) Clear(key uint) {
	l.mu.Lock()
	delete(l.cache, key)
	l.mu.Unlock()
}

func (l *CorporationAllianceHistoryLoader) unsafeSet(key uint, value []*monocle.CorporationAllianceHistory) {
	if l.cache == nil {
		l.cache = map[uint][]*monocle.CorporationAllianceHistory{}
	}
	l.cache[key] = value
}

// keyIndex will return the location of the key in the batch, if its not found
// it will add the key to the batch
func (b *corporationAllianceHistoryLoaderBatch) keyIndex(l *CorporationAllianceHistoryLoader, key uint) int {
	for i, existingKey := range b.keys {
		if key == existingKey {
			return i
		}
	}

	pos := len(b.keys)
	b.keys = append(b.keys, key)
	if pos == 0 {
		go b.startTimer(l)
	}

	if l.maxBatch != 0 && pos >= l.maxBatch-1 {
		if !b.closing {
			b.closing = true
			l.batch = nil
			go b.end(l)
		}
	}

	return pos
}

func (b *corporationAllianceHistoryLoaderBatch) startTimer(l *CorporationAllianceHistoryLoader) {
	time.Sleep(l.wait)
	l.mu.Lock()

	// we must have hit a batch limit and are already finalizing this batch
	if b.closing {
		l.mu.Unlock()
		return
	}

	l.batch = nil
	l.mu.Unlock()

	b.end(l)
}

func (b *corporationAllianceHistoryLoaderBatch) end(l *CorporationAllianceHistoryLoader) {
	b.data, b.error = l.fetch(b.keys)
	close(b.done)
}
