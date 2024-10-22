package mediabox

import (
	"context"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
)

var Pool *WorkerPool

type Job interface {
	Run() error
}

type WorkerPool struct {
	JobChan chan Job
	Size    int

	Workers map[string]worker
	mut     sync.RWMutex

	addWorkerChan    chan struct{}
	cancelWorkerChan chan struct{}

	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup
}

func NewWorkerPool(ctx context.Context, size int) *WorkerPool {
	ctx, cancel := context.WithCancel(ctx)

	return &WorkerPool{
		JobChan:          make(chan Job, 100),
		addWorkerChan:    make(chan struct{}),
		cancelWorkerChan: make(chan struct{}),
		Size:             size,
		Workers:          make(map[string]worker),
		ctx:              ctx,
		cancel:           cancel,
	}
}

func (p *WorkerPool) Run() {
	for i := 0; i < p.Size; i++ {
		p.addWorker()
	}

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-p.addWorkerChan:
			p.addWorker()

		case <-p.cancelWorkerChan:
			p.delOnceWorker()

		case <-ticker.C:
			if len(p.Workers) < p.Size {
				p.addWorkerChan <- struct{}{}
			}

			if len(p.Workers) > p.Size {
				p.cancelWorkerChan <- struct{}{}
			}
		case <-p.ctx.Done():
			logger.Println("stopping worker pool")
			p.mut.Lock()
			for _, w := range p.Workers {
				w.cancel()
			}
			p.mut.Unlock()
			return
		}
	}
}

func (p *WorkerPool) addWorker() {
	p.wg.Add(1)

	w := newWorker(p.ctx, p.JobChan, &p.wg)
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
	worker.cancel()

	p.mut.Lock()
	delete(p.Workers, worker.ID)
	p.mut.Unlock()

	logger.Infof("delete worker [%s] from pool", worker.ID)
}

func (p *WorkerPool) Submit(j Job) {
	if j == nil {
		logger.Errorf("submit nil job")
		return
	}

	logger.Infof("submit job")
	p.JobChan <- j
}

func (p *WorkerPool) Scale(poolSize int) {
	p.Size = poolSize
}

func (p *WorkerPool) Stop() {
	close(p.JobChan)

	p.wg.Wait()
	p.cancel()
}

type worker struct {
	ID      string
	jobChan chan Job

	ctx    context.Context
	cancel context.CancelFunc

	wg *sync.WaitGroup
}

func newWorker(ctx context.Context, jobChan chan Job, wg *sync.WaitGroup) *worker {
	c, cancel := context.WithCancel(ctx)

	return &worker{
		ID:      uuid.NewV4().String(),
		jobChan: jobChan,

		ctx:    c,
		cancel: cancel,
		wg:     wg,
	}
}

func (w *worker) Run() {
	defer w.wg.Done()

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
