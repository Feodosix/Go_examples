package main

const maxReaders = 1 << 4

type RWMutex struct {
	writer  chan struct{}
	readers chan struct{}
}

func New() *RWMutex {
	rw := &RWMutex{
		writer:  make(chan struct{}, 1),
		readers: make(chan struct{}, maxReaders),
	}

	rw.writer <- struct{}{}
	for i := 0; i < maxReaders; i++ {
		rw.readers <- struct{}{}
	}
	return rw
}

func (rw *RWMutex) RLock() {
	<-rw.readers
}

func (rw *RWMutex) RUnlock() {
	rw.readers <- struct{}{}
}

func (rw *RWMutex) Lock() {
	<-rw.writer

	for i := 0; i < cap(rw.readers); i++ {
		<-rw.readers
	}
}

func (rw *RWMutex) Unlock() {
	for i := 0; i < cap(rw.readers); i++ {
		rw.readers <- struct{}{}
	}
	rw.writer <- struct{}{}
}
