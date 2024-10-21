package mediabox

import (
	"context"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Job interface {
	Run() error
}

type WorkerPool struct {
	JobChan     chan Job
	Concurrency int

	Workers map[string]worker
	mut     sync.RWMutex

	addWorkerChan    chan struct{}
	cancelWorkerChan chan struct{}

	ctx    context.Context
	cancel context.CancelFunc
}

func NewWorkerPool(ctx context.Context, concurrent int) *WorkerPool {
	ctx, cancel := context.WithCancel(ctx)

	return &WorkerPool{
		JobChan:          make(chan Job),
		addWorkerChan:    make(chan struct{}),
		cancelWorkerChan: make(chan struct{}),
		Concurrency:      concurrent,
		Workers:          make(map[string]worker),
		ctx:              ctx,
		cancel:           cancel,
	}
}

func (p *WorkerPool) Run() {
	for i := 0; i < p.Concurrency; i++ {
		p.addWorker()
	}

	for {
		select {
		case <-p.addWorkerChan:
			p.addWorker()

		case <-p.cancelWorkerChan:
			p.delOnceWorker()

		case <-time.NewTicker(time.Second * 1).C:
			if len(p.Workers) < p.Concurrency {
				go func() {
					p.addWorkerChan <- struct{}{}
				}()
			}

			if len(p.Workers) > p.Concurrency {
				go func() {
					p.cancelWorkerChan <- struct{}{}
				}()
			}
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *WorkerPool) addWorker() {
	w := newWorker(p.ctx, p.JobChan)
	go w.Run()
	logger.Infof("new worker [%s]", w.ID)

	p.mut.Lock()
	p.Workers[w.ID] = *w
	p.mut.Unlock()

	logger.Infof("add worker [%s] to pool", w.ID)
}

func (p *WorkerPool) delOnceWorker() {
	var worker worker

	p.mut.RLock()
	for _, w := range p.Workers {
		worker = w
		break
	}
	p.mut.RUnlock()

	logger.Infof("cancel worker [%s]", worker.ID)
	worker.Cancel()

	p.mut.Lock()
	delete(p.Workers, worker.ID)
	p.mut.Unlock()

	logger.Infof("delete worker [%s] from pool", worker.ID)
}

func (p *WorkerPool) Submit(j Job) {
	p.JobChan <- j
}

func (p *WorkerPool) Scale(poolSize int) {
	p.Concurrency = c
}

type worker struct {
	ID      string
	jobChan chan Job

	ctx    context.Context
	Cancel context.CancelFunc
}

func newWorker(ctx context.Context, jobChan chan Job) *worker {
	c, cancel := context.WithCancel(ctx)

	return &worker{
		ID:      uuid.NewV4().String(),
		jobChan: jobChan,

		ctx:    c,
		Cancel: cancel,
	}
}

func (w *worker) Run() {
	for {
		select {
		case job := <-w.jobChan:
			logger.Infof("worker [%s] receive job", w.ID)
			w.run(job)
		case <-w.ctx.Done():
			logger.Infof("worker [%s] exit", w.ID)
			return
		}
	}
}

func (w *worker) run(j Job) {
	var err error

	defer func() {
		if err != nil {
			logger.Errorf("job error, msg: %v", err)
		}
	}()

	err = j.Run()
}

func (w *worker) Submit(j Job) {
	w.jobChan <- j
}
