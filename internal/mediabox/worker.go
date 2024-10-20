package mediabox

import (
	"context"
	"log"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
)

var Pool *WorkerPool

func InitWorkPool() {
	ctx := context.TODO()
	Pool = NewPool(ctx, 1)
}

type Job interface {
	Run()
}

type WorkerPool struct {
	JobChan     chan Job
	Concurrency int

	Workers map[string]Worker
	mut     sync.RWMutex

	addWorkerChan    chan struct{}
	cancelWorkerChan chan struct{}

	ctx context.Context

	cancel   context.CancelFunc
	stopChan chan struct{}
	wg       sync.WaitGroup // 用于等待所有 worker 完成
}

func NewPool(ctx context.Context, concurrent int) *WorkerPool {
	ctx, cancel := context.WithCancel(ctx)

	p := &WorkerPool{
		JobChan:          make(chan Job),
		addWorkerChan:    make(chan struct{}),
		cancelWorkerChan: make(chan struct{}),
		Concurrency:      concurrent,
		Workers:          make(map[string]Worker),
		ctx:              ctx,
		cancel:           cancel,
		stopChan:         make(chan struct{}),
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-p.addWorkerChan:
				p.wg.Add(1) // 增加等待组计数器

				w := New(p.ctx, p.JobChan, &p.wg)
				go w.Run()
				log.Printf("new worker [%s]", w.ID)

				p.mut.Lock()
				p.Workers[w.ID] = *w
				log.Printf("add worker [%s] to pool", w.ID)
				p.mut.Unlock()

			case <-p.cancelWorkerChan:
				var worker Worker
				p.mut.RLock()
				for _, w := range p.Workers {
					worker = w
					break
				}
				p.mut.RUnlock()

				log.Printf("cancel worker [%s]", worker.ID)
				worker.Cancel()

				p.mut.Lock()
				delete(p.Workers, worker.ID)
				log.Printf("remove worker [%s] from pool", worker.ID)
				p.mut.Unlock()

			case <-ticker.C:
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

			case <-p.stopChan:
				log.Println("stopping worker pool")
				p.cancel()
				p.mut.Lock()
				for _, w := range p.Workers {
					w.Cancel()
				}
				p.mut.Unlock()
				close(p.JobChan)
				return
			}
		}
	}()

	return p
}

func (p *WorkerPool) Submit(j Job) {
	p.JobChan <- j
}

func (p *WorkerPool) SetPoolSize(c int) {
	p.Concurrency = c
}

func (p *WorkerPool) GetCurrentWorkers() (ws []*Worker) {
	p.mut.RLock()
	defer p.mut.RUnlock()

	for _, w := range p.Workers {
		ws = append(ws, &w)
	}

	return
}

type Worker struct {
	ID      string
	jobChan chan Job

	ctx    context.Context
	Cancel context.CancelFunc

	wg *sync.WaitGroup
}

func New(ctx context.Context, jobChan chan Job, wg *sync.WaitGroup) *Worker {
	c, cancel := context.WithCancel(ctx)

	return &Worker{
		ID:      uuid.NewV4().String(),
		jobChan: jobChan,

		ctx:    c,
		Cancel: cancel,
		wg:     wg,
	}
}

func (w *Worker) Run() {
	defer w.wg.Done()

	for {
		select {
		case job := <-w.jobChan:
			log.Printf("worker [%s] receive job", w.ID)
			w.run(job)
		case <-w.ctx.Done():
			log.Printf("worker [%s] exit", w.ID)
			return
		}
	}
}

func (w *Worker) run(j Job) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("job error, msg: %v", err)
		}
	}()

	j.Run()
}

func (w *Worker) Submit(j Job) {
	w.jobChan <- j
}

func (p *WorkerPool) Stop() {
	close(p.stopChan)
}

func (p *WorkerPool) Wait() {
	p.wg.Wait()
}
