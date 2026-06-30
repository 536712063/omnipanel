package dtd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TaskType string

const (
	TaskRestart           TaskType = "restart"
	TaskBroadcast         TaskType = "broadcast"
	TaskBloodMoonCountdown TaskType = "blood_moon_countdown"
	TaskCustomCommand     TaskType = "custom_command"
)

type ScheduledTask struct {
	ID        string
	Name      string
	Type      TaskType
	Schedule  string
	Interval  time.Duration
	Enabled   bool
	LastRunAt *time.Time
	NextRunAt *time.Time
	Payload   string
	CreatedAt time.Time
}

type Scheduler struct {
	tasks  map[string]*ScheduledTask
	mu     sync.RWMutex
	runner TaskRunner
	stopCh chan struct{}
	wg     sync.WaitGroup
	ticker *time.Ticker
}

type TaskRunner interface {
	RunTask(ctx context.Context, task ScheduledTask) error
}

func NewScheduler(runner TaskRunner) *Scheduler {
	return &Scheduler{
		tasks:  make(map[string]*ScheduledTask),
		runner: runner,
		stopCh: make(chan struct{}),
		ticker: time.NewTicker(30 * time.Second),
	}
}

func (s *Scheduler) AddTask(task ScheduledTask) (*ScheduledTask, error) {
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	if task.Interval == 0 && task.Schedule == "" {
		return nil, fmt.Errorf("task must have interval or schedule")
	}
	if task.Type == "" {
		return nil, fmt.Errorf("task type is required")
	}
	task.CreatedAt = time.Now()
	now := time.Now()
	task.NextRunAt = &now

	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = &task
	return &task, nil
}

func (s *Scheduler) RemoveTask(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, id)
}

func (s *Scheduler) GetTasks() []ScheduledTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]ScheduledTask, 0, len(s.tasks))
	for _, t := range s.tasks {
		result = append(result, *t)
	}
	return result
}

func (s *Scheduler) UpdateTask(task ScheduledTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[task.ID]; !ok {
		return fmt.Errorf("task not found")
	}
	s.tasks[task.ID] = &task
	return nil
}

func (s *Scheduler) Start() {
	s.wg.Add(1)
	go s.loop()
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	s.ticker.Stop()
}

func (s *Scheduler) loop() {
	defer s.wg.Done()
	for {
		select {
		case <-s.stopCh:
			return
		case <-s.ticker.C:
			s.checkAndRun()
		}
	}
}

func (s *Scheduler) checkAndRun() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, task := range s.tasks {
		if !task.Enabled {
			continue
		}
		if task.NextRunAt == nil || task.NextRunAt.After(now) {
			continue
		}
		t := *task
		t.LastRunAt = &now
		next := now.Add(task.Interval)
		t.NextRunAt = &next
		s.tasks[task.ID] = &t

		go func(task ScheduledTask) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			if s.runner != nil {
				_ = s.runner.RunTask(ctx, task)
			}
		}(t)
	}
}

func CalculateBloodMoon(currentDay int, bloodMoonDay int) int {
	if currentDay <= bloodMoonDay {
		return bloodMoonDay
	}
	diff := currentDay - bloodMoonDay
	cycles := diff / 7
	if diff%7 > 0 {
		cycles++
	}
	return bloodMoonDay + cycles*7
}
