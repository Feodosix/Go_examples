package main

type Locker interface {
	Lock()
	Unlock()
}

type Cond struct {
	L       Locker
	waiters []chan struct{}
	lock    chan struct{}
}

func New(l Locker) *Cond {
	c := &Cond{
		L:       l,
		waiters: make([]chan struct{}, 0),
		lock:    make(chan struct{}, 1),
	}

	c.lock <- struct{}{}
	return c
}

func (c *Cond) Wait() {
	ch := make(chan struct{})

	<-c.lock
	c.waiters = append(c.waiters, ch)
	c.lock <- struct{}{}

	c.L.Unlock()
	<-ch
	c.L.Lock()
}

func (c *Cond) Signal() {
	<-c.lock
	if len(c.waiters) > 0 {
		waiter := c.waiters[0]
		c.waiters = c.waiters[1:]
		close(waiter)
	}
	c.lock <- struct{}{}
}

func (c *Cond) Broadcast() {
	<-c.lock
	for _, waiter := range c.waiters {
		close(waiter)
	}
	c.waiters = c.waiters[:0]
	c.lock <- struct{}{}
}
