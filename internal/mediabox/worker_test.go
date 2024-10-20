package mediabox

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockJob struct {
	ID        int
	Completed bool
	mu        sync.Mutex
}

func (m *MockJob) Run() {
	time.Sleep(100 * time.Millisecond) // 模拟任务执行
	m.mu.Lock()
	m.Completed = true
	m.mu.Unlock()
}

func TestWorkerPool_InitAndSubmitJob(t *testing.T) {
	// 初始化WorkerPool
	InitWorkPool()

	// 创建mock job
	job := &MockJob{ID: 1}

	// 提交job到池中
	Pool.Submit(job)

	// 等待工作完成
	time.Sleep(500 * time.Millisecond)

	// 检查job是否完成
	assert.True(t, job.Completed, "Job should be completed")
}

func TestWorkerPool_SetPoolSize(t *testing.T) {
	// 初始化WorkerPool
	ctx := context.TODO()
	pool := NewPool(ctx, 2)

	// 设置新的并发数量
	pool.SetPoolSize(3)

	// 验证并发数
	assert.Equal(t, 3, pool.Concurrency, "Pool size should be updated to 3")

	// 停止池
	pool.Stop()
}

func TestWorkerPool_AddAndCancelWorker(t *testing.T) {
	// 初始化WorkerPool
	ctx := context.TODO()
	pool := NewPool(ctx, 1)

	// 增加Worker
	pool.addWorkerChan <- struct{}{}
	time.Sleep(100 * time.Millisecond)

	// 检查当前是否有1个worker
	workers := pool.GetCurrentWorkers()
	assert.Equal(t, 1, len(workers), "There should be 1 worker in the pool")

	// 取消一个worker
	pool.cancelWorkerChan <- struct{}{}
	time.Sleep(100 * time.Millisecond)

	// 检查worker是否被移除
	workers = pool.GetCurrentWorkers()
	assert.Equal(t, 0, len(workers), "Worker should be removed from the pool")

	// 停止池
	pool.Stop()
}

func TestWorker_RunJob(t *testing.T) {
	// 创建Worker
	ctx := context.TODO()
	jobChan := make(chan Job)
	wg := sync.WaitGroup{}
	worker := New(ctx, jobChan, &wg)

	// 创建Mock Job
	job := &MockJob{ID: 1}
	wg.Add(1)

	// 启动worker
	go worker.Run()

	// 提交job到worker
	worker.Submit(job)

	// 等待工作完成
	wg.Wait()

	// 检查job是否完成
	assert.True(t, job.Completed, "Job should be completed")

	// 取消worker
	worker.Cancel()
}

func TestWorkerPool_StopAndWait(t *testing.T) {
	// 初始化WorkerPool
	ctx := context.TODO()
	pool := NewPool(ctx, 1)

	// 提交一些job
	for i := 0; i < 5; i++ {
		job := &MockJob{ID: i + 1}
		pool.Submit(job)
	}

	// 停止worker pool
	pool.Stop()

	// 确保所有worker都已完成
	pool.Wait()

	// 验证所有job都已经完成
	workers := pool.GetCurrentWorkers()
	for _, worker := range workers {
		assert.Empty(t, worker.jobChan, "All job channels should be closed")
	}
}
