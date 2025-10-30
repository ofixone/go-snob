package workerpool

import (
	"context"
	"sync"
)

type Processor[P any] func(ctx context.Context, p P)

type Worker[T any] struct {
	wg     *sync.WaitGroup
	procWg *sync.WaitGroup
	ch     <-chan T
	proc   func(context.Context, T)
}

func (w *Worker[T]) Start(ctx context.Context) {
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case payload, ok := <-w.ch:
			if !ok {
				return
			}

			w.procWg.Add(1)
			go func(p T) {
				defer w.procWg.Done()
				w.proc(ctx, p)
			}(payload)
		}
	}
}

type Pool[P any] struct {
	payloadCh chan P
	wg        *sync.WaitGroup
	procWg    *sync.WaitGroup
	workers   []*Worker[P]
}

func NewPool[P any](proc Processor[P], workerCount, queueSize int) *Pool[P] {
	ch := make(chan P, queueSize)
	wg := &sync.WaitGroup{}
	procWg := &sync.WaitGroup{}
	wp := &Pool[P]{
		payloadCh: ch,
		wg:        wg,
		procWg:    procWg,
	}

	wp.workers = make([]*Worker[P], workerCount)
	for i := 0; i < workerCount; i++ {
		worker := &Worker[P]{
			wg:     wg,
			procWg: procWg,
			ch:     ch,
			proc:   proc,
		}
		wp.workers[i] = worker
	}

	return wp
}

func (wp *Pool[P]) WrChan() chan<- P {
	return wp.payloadCh
}

func (wp *Pool[P]) SetChan(ch chan P) {
	wp.payloadCh = ch
}

func (wp *Pool[P]) Start(ctx context.Context) {
	wp.wg.Add(len(wp.workers))
	for _, w := range wp.workers {
		go w.Start(ctx)
	}
}

func (wp *Pool[P]) Shutdown() {
	close(wp.payloadCh)
	wp.wg.Wait()
	wp.procWg.Wait()
}
