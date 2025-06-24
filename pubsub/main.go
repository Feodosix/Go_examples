package main

import (
	"context"
	"errors"
	"sync"
)

type MySubscription struct {
	pubsub    *MyPubSub
	topic     string
	handler   MsgHandler
	mu        sync.Mutex
	queue     []interface{}
	cond      *sync.Cond
	closed    bool
	done      chan struct{}
	unsubOnce sync.Once
}

func (s *MySubscription) run() {
	defer close(s.done)
	for {
		s.mu.Lock()
		for len(s.queue) == 0 && !s.closed {
			s.cond.Wait()
		}
		if len(s.queue) == 0 && s.closed {
			s.mu.Unlock()
			return
		}
		msg := s.queue[0]
		s.queue = s.queue[1:]
		s.mu.Unlock()

		s.handler(msg)
	}
}

func (s *MySubscription) addqueue(msg interface{}) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.queue = append(s.queue, msg)
	s.cond.Signal()
	s.mu.Unlock()
}

func (s *MySubscription) Unsubscribe() {
	s.unsubOnce.Do(func() {
		s.pubsub.removeSubscription(s.topic, s)
		s.mu.Lock()
		s.closed = true
		s.cond.Broadcast()
		s.mu.Unlock()
	})
}

type MyPubSub struct {
	mu     sync.Mutex
	subs   map[string]map[*MySubscription]struct{}
	closed bool
}

func NewPubSub() PubSub {
	return &MyPubSub{
		subs: make(map[string]map[*MySubscription]struct{}),
	}
}

func (p *MyPubSub) Subscribe(subj string, cb MsgHandler) (Subscription, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, errors.New("pubsub closed")
	}

	sub := &MySubscription{
		pubsub:  p,
		topic:   subj,
		handler: cb,
		queue:   make([]interface{}, 0),
		done:    make(chan struct{}),
	}
	sub.cond = sync.NewCond(&sub.mu)
	if _, ok := p.subs[subj]; !ok {
		p.subs[subj] = make(map[*MySubscription]struct{})
	}
	p.subs[subj][sub] = struct{}{}
	p.mu.Unlock()

	go sub.run()

	return sub, nil
}

func (p *MyPubSub) Publish(subj string, msg interface{}) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return errors.New("pubsub closed")
	}

	subs, ok := p.subs[subj]
	if !ok {
		p.mu.Unlock()
		return nil
	}

	subsCopy := make([]*MySubscription, 0, len(subs))
	for s := range subs {
		subsCopy = append(subsCopy, s)
	}
	p.mu.Unlock()

	for _, s := range subsCopy {
		s.addqueue(msg)
	}

	return nil
}

func (p *MyPubSub) removeSubscription(topic string, sub *MySubscription) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return
	}
	if subs, ok := p.subs[topic]; ok {
		delete(subs, sub)
		if len(subs) == 0 {
			delete(p.subs, topic)
		}
	}
}

func (p *MyPubSub) Close(ctx context.Context) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	var allSubs []*MySubscription
	for _, subs := range p.subs {
		for s := range subs {
			allSubs = append(allSubs, s)
		}
	}
	p.subs = nil
	p.mu.Unlock()

	for _, s := range allSubs {
		s.mu.Lock()
		s.closed = true
		s.cond.Broadcast()
		s.mu.Unlock()
	}

	done := make(chan struct{})
	go func() {
		for _, s := range allSubs {
			<-s.done
		}
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
