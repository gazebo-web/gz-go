package scheduler

import (
	"gitlab.com/ignitionrobotics/web/ign-go/v6"
	sch "gitlab.com/ignitionrobotics/web/scheduler"
	"sync"
	"time"
)

// Scheduler represents a generic Scheduler
type Scheduler struct {
	*sch.Scheduler
}

// TaskScheduler defines the basic operations that every Scheduler should fulfill
type TaskScheduler interface {
	DoIn(task func(), seconds int) string
	DoEvery(task func(), seconds int) string
	DoAt(task func(), date time.Time) string
}

var once sync.Once
var scheduler *Scheduler

// initializeScheduler instantiate the singleton
func initializeScheduler() *ign.ErrMsg {
	s, err := sch.NewScheduler(1000)
	if err != nil {
		return ign.NewErrorMessage(ign.ErrorScheduler)
	}
	scheduler = &Scheduler{
		Scheduler: s,
	}
	return nil
}

// GetInstance returns the scheduler singleton.
func GetInstance() *Scheduler {
	once.Do(func() {
		initializeScheduler()
	})
	return scheduler
}

// DoIn runs a specific task after a number of seconds.
func (s *Scheduler) DoIn(task func(), seconds int) string {
	return scheduler.Delay().Second(seconds).Do(task)
}

// DoEvery repeatedly runs a task after a number of seconds.
func (s *Scheduler) DoEvery(task func(), seconds int) string {
	return scheduler.Every().Second(seconds).Do(task)
}

// DoAt runs a task on a specific date.
// If the date is in the past, DoAt will run instantly.
func (s *Scheduler) DoAt(task func(), date time.Time) string {
	now := time.Now()
	diff := date.Sub(now)

	seconds := int(diff.Seconds())

	if seconds < 0 {
		seconds = 0
	}

	return s.DoIn(task, seconds)
}
