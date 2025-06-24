package main

import (
	"sort"
	"sync"
)

type KeyLock struct {
	mu     sync.Mutex
	locked map[string]struct{}
	update chan struct{}
}

func New() *KeyLock {
	return &KeyLock{
		locked: make(map[string]struct{}),
		update: make(chan struct{}),
	}
}

func (l *KeyLock) LockKeys(keys []string, cancel <-chan struct{}) (canceled bool, unlock func()) {
	keysCopy := make([]string, len(keys))
	copy(keysCopy, keys)
	sort.Strings(keysCopy)

	l.mu.Lock()
	for {
		conflict := false
		for _, key := range keysCopy {
			if _, inUse := l.locked[key]; inUse {
				conflict = true
				break
			}
		}
		if !conflict {
			for _, key := range keysCopy {
				l.locked[key] = struct{}{}
			}
			l.mu.Unlock()

			return false, func() {
				l.mu.Lock()
				for _, key := range keysCopy {
					delete(l.locked, key)
				}
				close(l.update)
				l.update = make(chan struct{})
				l.mu.Unlock()
			}
		}
		currUpdate := l.update
		l.mu.Unlock()

		select {
		case <-cancel:
			return true, func() {}
		case <-currUpdate:
		}

		l.mu.Lock()
	}
}
