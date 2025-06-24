package main

type WaitGroup struct {
	counter int
	done    chan struct{}
	lock    chan struct{}
}

func New() *WaitGroup {
	wg := &WaitGroup{
		counter: 0,
		done:    make(chan struct{}),
		lock:    make(chan struct{}, 1),
	}

	wg.lock <- struct{}{}
	close(wg.done)
	return wg
}

func (wg *WaitGroup) Add(delta int) {
	<-wg.lock
	newCounter := wg.counter + delta

	if newCounter < 0 {
		wg.lock <- struct{}{}
		panic("negative WaitGroup counter")
	}

	if wg.counter == 0 && delta > 0 {
		wg.done = make(chan struct{})
	}

	wg.counter = newCounter

	if wg.counter == 0 {
		close(wg.done)
	}

	wg.lock <- struct{}{}
}

func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

func (wg *WaitGroup) Wait() {
	<-wg.lock
	ch := wg.done
	wg.lock <- struct{}{}
	<-ch
}
