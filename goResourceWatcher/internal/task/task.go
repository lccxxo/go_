package task

import (
	"context"
	"fmt"
	"sync"

	"github.com/go_/lccxxo/goResourceWatcher/internal/logger"
	"github.com/go_/lccxxo/goResourceWatcher/internal/utils"
)

// Task 单个任务的结构体
type Task struct {
	ID       string                                   `json:"id"`
	TaskName string                                   `json:"taskName"`
	Func     func(context.Context, *sync.Cond, *bool) `json:"func"`
	ctx      context.Context                          `json:"ctx"`
	cancel   context.CancelFunc                       `json:"cancel"`
	pause    *sync.Cond                               `json:"pause"`
	isPaused bool                                     `json:"isPaused"`
}

func NewTask(id, taskName string, f func(context.Context, *sync.Cond, *bool)) *Task {
	ctx, cancel := context.WithCancel(context.Background())

	return &Task{
		ID:       id,
		TaskName: taskName,
		Func:     f,
		ctx:      ctx,
		cancel:   cancel,
		pause:    sync.NewCond(&sync.Mutex{}),
		isPaused: false,
	}
}

// TaskQueue 任务队列的结构体
type TaskQueue struct {
	tasks map[string]*Task
	mu    sync.Mutex
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		tasks: make(map[string]*Task),
	}
}

func (tq *TaskQueue) run(task *Task) {
	task.Func(task.ctx, task.pause, &task.isPaused)
}

// 添加任务 并自动执行
func (tq *TaskQueue) AddTask(taskName string, f func(context.Context, *sync.Cond, *bool)) string {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	taskId := fmt.Sprintf("task-%s", utils.GenerateUID())

	task := NewTask(taskId, taskName, f)

	tq.tasks[taskId] = task
	go tq.run(task)
	err := Create(ToTaskEntity(*task))
	if err != nil {
		logger.Logger.Infof("mysql data task-%s insert error:%v\n", taskId, err)
		return ""
	}
	return taskId
}

// 暂停任务
func (tq *TaskQueue) PauseTask(taskId string) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	task, ok := tq.tasks[taskId]
	if ok {
		task.pause.L.Lock()
		task.isPaused = true
		err := Pause(ToTaskEntity(*task))
		if err != nil {
			logger.Logger.Infof("mysql data task-%s update error:%v\n", taskId, err)
			return
		}
	}
}

// 恢复任务
func (tq *TaskQueue) ResumeTask(taskId string) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	task, ok := tq.tasks[taskId]
	if ok {
		task.isPaused = false
		task.pause.L.Unlock()
		task.pause.Signal()
		err := Resume(ToTaskEntity(*task))
		if err != nil {
			logger.Logger.Infof("mysql data task-%s update error:%v\n", taskId, err)
			return
		}
	}
}

// 删除任务
func (tq *TaskQueue) DeleteTask(taskId string) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	if task, exists := tq.tasks[taskId]; exists {
		task.cancel()
		delete(tq.tasks, taskId)
		err := Delete(ToTaskEntity(*task))
		if err != nil {
			logger.Logger.Infof("mysql data task-%s delete error:%v\n", taskId, err)
			return
		}
	}
}
