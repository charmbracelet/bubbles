package memoization

import (
	"container/list"
	"crypto/sha256"
	"fmt"
	"sync"
)

type Hasher interface {
	Hash() string
}

// entry is used to hold a value in the evictionList
type entry[T any] struct {
	key   string
	value T
}

// MemoCache represents a cache with a set capacity that uses an LRU eviction policy.
type MemoCache[H Hasher, T any] struct {
	capacity      int
	mutex         sync.Mutex
	cache         map[string]*list.Element // The cache holding the results
	evictionList  *list.List               // A list to keep track of the order for LRU
	hashableItems map[string]T             // This map keeps track of the original hashable items (optional)
}

// NewMemoCache creates a new MemoCache given a certain capacity.
func NewMemoCache[H Hasher, T any](capacity int) *MemoCache[H, T] {
	return &MemoCache[H, T]{
		capacity:      capacity,
		cache:         make(map[string]*list.Element),
		evictionList:  list.New(),
		hashableItems: make(map[string]T),
	}
}

func (m *MemoCache[H, T]) Capacity() int {
	return m.capacity
}

func (m *MemoCache[H, T]) Size() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.evictionList.Len()
}

// Get returns the value associated with the given hashable item.
// If there is no corresponding value, the method returns nil.
func (m *MemoCache[H, T]) Get(h H) (T, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	hashedKey := h.Hash()
	if element, found := m.cache[hashedKey]; found {
		m.evictionList.MoveToFront(element)
		return element.Value.(*entry[T]).value, true
	}
	var result T
	return result, false
}

// Set sets the value for the hashable item.
func (m *MemoCache[H, T]) Set(h H, value T) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	hashedKey := h.Hash()
	if element, found := m.cache[hashedKey]; found {
		m.evictionList.MoveToFront(element)
		element.Value.(*entry[T]).value = value
		return
	}

	// Check if the cache is at capacity
	if m.evictionList.Len() >= m.capacity {
		// Evict the least recently used item from the cache
		toEvict := m.evictionList.Back()
		if toEvict != nil {
			evictedEntry := m.evictionList.Remove(toEvict).(*entry[T])
			delete(m.cache, evictedEntry.key)
			delete(m.hashableItems, evictedEntry.key) // if you're keeping track of original items
		}
	}

	// Add the value to the cache and the evictionList
	newEntry := &entry[T]{
		key:   hashedKey,
		value: value,
	}
	element := m.evictionList.PushFront(newEntry)
	m.cache[hashedKey] = element
	m.hashableItems[hashedKey] = value // if you're keeping track of original items
}

type HString string

func (h HString) Hash() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(h)))
}

type HInt int

func (h HInt) Hash() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d", h))))
}
