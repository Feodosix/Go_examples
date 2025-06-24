package main

import (
	"context"
	"sync"
)

type callGroup struct {
	res          interface{}
	err          error
	done         chan struct{}
	waiting      int
	cancelWorker context.CancelFunc
}

type Call struct {
	mu      sync.Mutex
	current *callGroup
}

func (c *Call) Do(
	ctx context.Context,
	cb func(context.Context) (interface{}, error),
) (interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	c.mu.Lock()

	if c.current != nil {
		select {
		case <-c.current.done:
			c.current = nil
		default:
		}
	}
	if c.current != nil {
		grp := c.current
		grp.waiting++
		c.mu.Unlock()

		select {
		case <-ctx.Done():
			c.mu.Lock()
			grp.waiting--
			if grp.waiting == 0 {
				grp.cancelWorker()
			}
			c.mu.Unlock()
			return nil, ctx.Err()
		case <-grp.done:
			res, err := grp.res, grp.err
			c.mu.Lock()
			grp.waiting--
			c.mu.Unlock()
			return res, err
		}
	}

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	grp := &callGroup{
		done:         make(chan struct{}),
		waiting:      1,
		cancelWorker: cancelWorker,
	}
	c.current = grp
	c.mu.Unlock()

	go func() {
		res, err := cb(workerCtx)
		c.mu.Lock()
		grp.res, grp.err = res, err
		close(grp.done)
		c.current = nil
		c.mu.Unlock()
	}()

	select {
	case <-ctx.Done():
		c.mu.Lock()
		grp.waiting--
		if grp.waiting == 0 {
			grp.cancelWorker()
		}
		c.mu.Unlock()
		return nil, ctx.Err()
	case <-grp.done:
		c.mu.Lock()
		res, err := grp.res, grp.err
		grp.waiting--
		c.mu.Unlock()
		return res, err
	}
}
