package gz

import (
	"sync"
	"time"
)

const (
	// WaitForNextElementChanCapacity is being used to set the maximum capacity of listeners
	WaitForNextElementChanCapacity = 1000
	// dequeueOrWaitForNextElementInvokeGapTime is being used to set the time between dequeue attempts.
	dequeueOrWaitForNextElementInvokeGapTime = 10
)

// Queue is a thread-safe data type that uses an underlying slice as a queue.
// It provides a set of methods to modify the elements of the queue.
// It was created to replace a go channel since channels don't let you modify their elements order.
// You can push elements inside the queue using Enqueue method. They will be pushed to the back of the queue.
// You can pop elements from the queue using both Dequeue() and DequeueOrWaitForNextElement().
// DequeueOrWaitForNextElement() waits until the queue has any element inside to pop it out.
type Queue struct {
	// slice represents the actual queue.
	slice []interface{}
	// rwmutex represents the read-write lock for the queue.
	rwmutex sync.RWMutex
	// queue for watchers that will wait for next elements (if queue is empty at DequeueOrWaitForNextElement execution)
	waitForNextElementChan chan chan interface{}
}

// NewQueue returns a new Queue instance.
func NewQueue() (queue *Queue) {
	queue = &Queue{}
	queue.initialize()
	return
}

// initialize initializes the queue properties.
func (st *Queue) initialize() {
	st.slice = make([]interface{}, 0)
	st.waitForNextElementChan = make(chan chan interface{}, WaitForNextElementChanCapacity)
}

// Enqueue enqueues an element.
func (st *Queue) Enqueue(value interface{}) {
	// check if there is a listener waiting for the next element (this element)
	select {
	case listener := <-st.waitForNextElementChan:
		// send the element through the listener's channel instead of enqueue it
		select {
		case listener <- value:
		default:
			st.enqueue(value, true)
		}

	default:
		st.enqueue(value, true)
	}
}

// enqueue appends a value to the queue
func (st *Queue) enqueue(value interface{}, lock bool) {
	if lock {
		st.rwmutex.Lock()
		defer st.rwmutex.Unlock()
	}

	st.slice = append(st.slice, value)
}

// Dequeue dequeues an element. Returns error if queue is locked or empty.
func (st *Queue) Dequeue() (interface{}, *ErrMsg) {
	st.rwmutex.Lock()
	defer st.rwmutex.Unlock()

	len := len(st.slice)
	if len == 0 {
		return nil, NewErrorMessage(ErrorQueueEmpty)
	}

	element := st.slice[0]
	st.slice = st.slice[1:]

	return element, nil
}

// DequeueOrWaitForNextElement dequeues an element (if exist) or waits until the next element gets enqueued and returns it.
// Multiple calls to DequeueOrWaitForNextElement() would enqueue multiple "listeners" for future enqueued elements.
func (st *Queue) DequeueOrWaitForNextElement() (interface{}, *ErrMsg) {
	for {
		// get the slice's len
		st.rwmutex.Lock()
		length := len(st.slice)
		st.rwmutex.Unlock()

		if length == 0 {
			// channel to wait for next enqueued element
			waitChan := make(chan interface{})

			select {
			// enqueue a watcher into the watchForNextElementChannel to wait for the next element
			case st.waitForNextElementChan <- waitChan:

				// re-checks every i milliseconds (top: 10 times) ... the following verifies if an item was enqueued
				// around the same time DequeueOrWaitForNextElement was invoked, meaning the waitChan wasn't yet sent over
				// st.waitForNextElementChan
				for i := 0; i < dequeueOrWaitForNextElementInvokeGapTime; i++ {
					select {
					case dequeuedItem := <-waitChan:
						return dequeuedItem, nil
					case <-time.After(time.Millisecond * time.Duration(i)):
						if dequeuedItem, err := st.Dequeue(); err == nil {
							return dequeuedItem, nil
						}
					}
				}

				// return the next enqueued element, if any
				return <-waitChan, nil
			default:
				// too many watchers (waitForNextElementChanCapacity) enqueued waiting for next elements
				return nil, NewErrorMessage(ErrorQueueTooManyListeners)
			}
		}

		st.rwmutex.Lock()

		// verify that at least 1 item resides on the queue
		if len(st.slice) == 0 {
			st.rwmutex.Unlock()
			continue
		}
		elementToReturn := st.slice[0]
		st.slice = st.slice[1:]

		st.rwmutex.Unlock()
		return elementToReturn, nil
	}
}

// GetElement returns an element's value and keeps the element at the queue
func (st *Queue) GetElement(index int) (interface{}, *ErrMsg) {
	st.rwmutex.RLock()
	defer st.rwmutex.RUnlock()

	if len(st.slice) <= index {
		return nil, NewErrorMessage(ErrorQueueIndexOutOfBounds)
	}

	return st.slice[index], nil
}

// GetElements returns the entire list of elements from the queue
func (st *Queue) GetElements() ([]interface{}, *ErrMsg) {

	st.rwmutex.RLock()
	defer st.rwmutex.RUnlock()

	return st.slice, nil
}

// GetFilteredElements returns a subset list from the queue
func (st *Queue) GetFilteredElements(offset, limit int) ([]interface{}, *ErrMsg) {
	st.rwmutex.Lock()
	defer st.rwmutex.Unlock()

	length := len(st.slice)

	if length == 0 {
		return st.slice, nil
	}

	if offset >= length || offset < 0 || limit <= 0 {
		return nil, NewErrorMessage(ErrorQueueIndexOutOfBounds)
	}

	if (offset + limit) >= length {
		limit = len(st.slice) - offset
	}
	low := offset
	high := offset + limit
	subset := st.slice[low:high]

	return subset, nil
}

// Find returns a list of ids of the elements that match the given criteria.
// Returns an empty slice if there are not elements that match.
func (st *Queue) Find(criteria func(element interface{}) bool) []int {
	return st.find(criteria, true)
}

// find returns a list of ids of the elements that match the given criteria.
func (st *Queue) find(criteria func(element interface{}) bool, lock bool) (result []int) {
	if lock {
		st.rwmutex.Lock()
		defer st.rwmutex.Unlock()
	}

	for id, item := range st.slice {
		if criteria(item) {
			result = append(result, id)
		}
	}

	return
}

// FindOne returns the id from a given element.
// Returns -1 if the element does not exist in the queue.
func (st *Queue) FindOne(target interface{}) int {
	return st.findOne(target, true)
}

// findOne returns the id from a given element.
func (st *Queue) findOne(target interface{}, lock bool) int {
	if lock {
		st.rwmutex.Lock()
		defer st.rwmutex.Unlock()
	}

	for id, item := range st.slice {
		if item == target {
			return id
		}
	}

	return -1
}

// FindByIDs returns a list of elements of the given ids.
// Returns an empty slice if there are no elements in the queue.
func (st *Queue) FindByIDs(ids []int) []interface{} {
	return st.findByIDs(ids, true)
}

// findByIDs returns a list of elements of the given ids.
func (st *Queue) findByIDs(ids []int, lock bool) (result []interface{}) {
	if lock {
		st.rwmutex.Lock()
		defer st.rwmutex.Unlock()
	}

	length := len(ids)
	count := 0

	for id, item := range st.slice {
		for _, i := range ids {
			if id == i {
				result = append(result, item)
				count++
				break
			}
		}

		if count == length {
			break
		}
	}

	return
}

// Remove removes an element from the queue
func (st *Queue) Remove(target interface{}) *ErrMsg {
	st.rwmutex.Lock()
	defer st.rwmutex.Unlock()

	index := st.findOne(target, false)

	if index == -1 {
		return NewErrorMessage(ErrorIDNotFound)
	}

	// remove the element
	st.slice = append(st.slice[:index], st.slice[index+1:]...)

	return nil
}

// GetLen returns the number of enqueued elements
func (st *Queue) GetLen() int {
	st.rwmutex.RLock()
	defer st.rwmutex.RUnlock()

	return len(st.slice)
}

// GetCap returns the queue's capacity
func (st *Queue) GetCap() int {
	st.rwmutex.RLock()
	defer st.rwmutex.RUnlock()

	return cap(st.slice)
}

// Swap swaps values A and B.
func (st *Queue) Swap(a interface{}, b interface{}) *ErrMsg {
	st.rwmutex.Lock()
	defer st.rwmutex.Unlock()

	length := len(st.slice)
	if length == 0 {
		return NewErrorMessage(ErrorQueueEmpty)
	}

	aIndex := st.findOne(a, false)
	bIndex := st.findOne(b, false)

	if aIndex == -1 || bIndex == -1 {
		return NewErrorMessage(ErrorIDNotFound)
	}

	if aIndex == bIndex {
		return NewErrorMessage(ErrorQueueSwapIndexesMatch)
	}

	st.slice[aIndex], st.slice[bIndex] = st.slice[bIndex], st.slice[aIndex]

	return nil
}

// MoveToFront moves an element to the front of the queue
func (st *Queue) MoveToFront(target interface{}) *ErrMsg {
	st.rwmutex.Lock()
	defer st.rwmutex.Unlock()

	length := len(st.slice)
	if length == 0 {
		return NewErrorMessage(ErrorQueueEmpty)
	}

	index := st.findOne(target, false)
	if index == -1 {
		return NewErrorMessage(ErrorIDNotFound)
	}

	if index == 0 {
		return NewErrorMessage(ErrorQueueMoveIndexFrontPosition)
	}

	// Moves the element all the way to the back of the queue.
	// The element is moved one position at a time using bubble sort algorithm.
	for i := index; i >= 1; i-- {
		st.slice[i], st.slice[i-1] = st.slice[i-1], st.slice[i]
	}
	return nil
}

// MoveToBack moves an element to the back of the queue
func (st *Queue) MoveToBack(target interface{}) *ErrMsg {
	st.rwmutex.Lock()
	defer st.rwmutex.Unlock()

	length := len(st.slice)
	if length == 0 {
		return NewErrorMessage(ErrorQueueEmpty)
	}

	index := st.findOne(target, false)

	if index == -1 {
		return NewErrorMessage(ErrorIDNotFound)
	}

	if index == length-1 {
		return NewErrorMessage(ErrorQueueMoveIndexBackPosition)
	}

	// Moves the element all the way to the back of the queue.
	// The element is moved one position at a time using bubble sort algorithm.
	for i := index; i < length-1; i++ {
		st.slice[i], st.slice[i+1] = st.slice[i+1], st.slice[i]
	}

	return nil
}
