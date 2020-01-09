package scheduler

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

// TestSingleton tests the correct type of the singleton
func TestSingleton(t *testing.T) {
	s1 := GetInstance()
	s2 := GetInstance()
	assert.IsType(t, Scheduler{}, *s1)
	assert.IsType(t, Scheduler{}, *s2)
	assert.Equal(t, s1, s2)
}

// TestDoIn tests the changes on data when using DoIn method.
func TestDoIn(t *testing.T) {
	s := GetInstance()
	b := false
	var wg sync.WaitGroup

	assert.Equal(t, false, b)

	wg.Add(1)
	s.DoIn(func() {
		b = true
		wg.Done()
	}, 1)

	wg.Wait()

	assert.Equal(t, true, b)
}

// TestDoEvery tests the changes on data when using DoEvery method.
func TestDoEvery(t *testing.T) {
	s := GetInstance()
	count := 1
	var wg sync.WaitGroup

	assert.Equal(t, 1, count)

	wg.Add(2)
	s.DoEvery(func() {
		count = count + 1
		wg.Done()
	}, 1)

	wg.Wait()

	assert.Equal(t, 3, count)
}

// TestDoAt tests the changes on data when using DoAt method.
func TestDoAt(t *testing.T) {
	s := GetInstance()
	b := false
	var wg sync.WaitGroup

	assert.Equal(t, false, b)

	wg.Add(1)
	s.DoAt(func() {
		b = true
		wg.Done()
	}, time.Now().Add(time.Second*3))

	wg.Wait()

	assert.Equal(t, true, b)
}
