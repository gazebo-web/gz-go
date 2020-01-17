package ign

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type QueueTestSuite struct {
	suite.Suite
	queue *Queue
}

const (
	testValue = "test value"
)

func (suite *QueueTestSuite) SetupTest() {
	suite.queue = NewQueue()
}

// ***************************************************************************************
// ** Queue initialization
// ***************************************************************************************

// no elements at initialization
func (suite *QueueTestSuite) TestNoElementsAtInitialization() {
	length := suite.queue.GetLen()
	suite.Equalf(0, length, "No elements expected at initialization, currently: %v", length)
}

// ***************************************************************************************
// ** Enqueue && GetLen
// ***************************************************************************************

// single enqueue (1 element, 1 goroutine)
func (suite *QueueTestSuite) TestEnqueueLenSingleGR() {
	suite.queue.Enqueue(testValue)
	length := suite.queue.GetLen()
	suite.Equalf(1, length, "Expected number of elements in queue: 1, currently: %v", length)

	suite.queue.Enqueue(5)
	length = suite.queue.GetLen()
	suite.Equalf(2, length, "Expected number of elements in queue: 2, currently: %v", length)
}

// single enqueue and wait for next element
func (suite *QueueTestSuite) TestEnqueueWaitForNextElementSingleGR() {
	waitForNextElement := make(chan interface{})
	// add the listener manually (ONLY for testings purposes)
	suite.queue.waitForNextElementChan <- waitForNextElement

	value := 100
	// enqueue from a different GR to avoid blocking the listener channel
	go suite.queue.Enqueue(value)
	// wait for the enqueued element
	result := <-waitForNextElement

	suite.Equal(value, result)
}

// TestEnqueueLenMultipleGR enqueues elements concurrently
//
// Detailed steps:
//	1 - Enqueue totalGRs concurrently (from totalGRs different GRs)
//	2 - Verifies the len, it should be equal to totalGRs
//	3 - Verifies that all elements from 0 to totalGRs were enqueued
func (suite *QueueTestSuite) TestEnqueueLenMultipleGR() {
	var (
		totalGRs = 500
		wg       sync.WaitGroup
	)

	// concurrent enqueueing
	// multiple GRs concurrently enqueueing consecutive integers from 0 to (totalGRs - 1)
	for i := 0; i < totalGRs; i++ {
		wg.Add(1)
		go func(value int) {
			defer wg.Done()
			suite.queue.Enqueue(value)
		}(i)
	}
	wg.Wait()

	// check that there are totalGRs elements enqueued
	totalElements := suite.queue.GetLen()
	suite.Equalf(totalGRs, totalElements, "Total enqueued elements should be %v, currently: %v", totalGRs, totalElements)

	// checking that the expected elements (1, 2, 3, ... totalGRs-1 ) were enqueued
	var (
		tmpVal                interface{}
		val                   int
		err                   *ErrMsg
		totalElementsVerified int
	)
	// slice to check every element
	sl2check := make([]bool, totalGRs)

	for i := 0; i < totalElements; i++ {
		tmpVal, err = suite.queue.GetElement(i)
		suite.Nil(err, "No error should be returned trying to get an existent element")

		val = tmpVal.(int)
		if !sl2check[val] {
			totalElementsVerified++
			sl2check[val] = true
		} else {
			suite.Failf("Duplicated element", "Unexpected duplicated value: %v", val)
		}
	}
	suite.True(totalElementsVerified == totalGRs, "Enqueued elements are missing")
}

// call GetLen concurrently
func (suite *QueueTestSuite) TestGetLenMultipleGRs() {
	var (
		totalGRs               = 100
		totalElementsToEnqueue = 10
		wg                     sync.WaitGroup
	)

	for i := 0; i < totalElementsToEnqueue; i++ {
		suite.queue.Enqueue(i)
	}

	for i := 0; i < totalGRs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			total := suite.queue.GetLen()
			suite.Equalf(totalElementsToEnqueue, total, "Expected len: %v", totalElementsToEnqueue)
		}()
	}
	wg.Wait()
}

// ***************************************************************************************
// ** GetCap
// ***************************************************************************************

// single GR getCapacity
func (suite *QueueTestSuite) TestGetCapSingleGR() {
	// initial capacity
	suite.Equal(cap(suite.queue.slice), suite.queue.GetCap(), "unexpected capacity")

	// checking after adding 2 items
	suite.queue.Enqueue(1)
	suite.queue.Enqueue(2)
	suite.Equal(cap(suite.queue.slice), suite.queue.GetCap(), "unexpected capacity")
}

// ***************************************************************************************
// ** Get
// ***************************************************************************************

// get a valid element
func (suite *QueueTestSuite) TestGetSingleGR() {
	suite.queue.Enqueue(testValue)
	val, err := suite.queue.GetElement(0)

	// verify error (should be nil)
	suite.Nil(err, "No error should be enqueueing an element")

	// verify element's value
	suite.Equalf(testValue, val, "Different element returned: %v", val)
}

// get a invalid element
func (suite *QueueTestSuite) TestGetInvalidElementSingleGR() {
	suite.queue.Enqueue(testValue)
	val, err := suite.queue.GetElement(1)

	// verify error
	suite.Equal(ErrorQueueIndexOutOfBounds, err.ErrCode, "An error should be returned after ask for a no existent element")

	// verify element's value
	suite.Equalf(val, nil, "Nil should be returned, currently returned: %v", val)
}

// call Get concurrently
func (suite *QueueTestSuite) TestGetMultipleGRs() {
	var (
		totalGRs               = 100
		totalElementsToEnqueue = 10
		wg                     sync.WaitGroup
	)

	for i := 0; i < totalElementsToEnqueue; i++ {
		suite.queue.Enqueue(i)
	}

	for i := 0; i < totalGRs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, err := suite.queue.GetElement(5)

			suite.Nil(err, "No error should be returned trying to get an existent element")
			suite.Equal(5, val.(int), "Expected element's value: 5")
		}()
	}
	wg.Wait()

	total := suite.queue.GetLen()
	suite.Equalf(totalElementsToEnqueue, total, "Expected len: %v", totalElementsToEnqueue)
}

// ***************************************************************************************
// ** Remove
// ***************************************************************************************

// remove elements
func (suite *QueueTestSuite) TestRemoveSingleGR() {
	suite.queue.Enqueue(testValue)
	suite.queue.Enqueue(5)

	// removing first element
	err := suite.queue.Remove(testValue)
	suite.Nil(err, "Unexpected error")

	// get element at index 0
	val, err2 := suite.queue.GetElement(0)
	suite.Nil(err2, "Unexpected error")
	suite.Equal(5, val, "Queue returned the wrong element")
}

func (suite *QueueTestSuite) TestRemoveNotFound() {
	suite.queue.Enqueue(testValue)

	err := suite.queue.Remove("anotherValue")

	suite.Equal(ErrorIDNotFound, err.ErrCode)
}

// TestRemoveMultipleGRs removes elements concurrently.
//
// Detailed steps:
//	1 - Enqueues totalElementsToEnqueue consecutive elements (0, 1, 2, 3, ... totalElementsToEnqueue - 1)
//	2 - Hits queue.Remove(1) concurrently from totalElementsToRemove different GRs
//	3 - Verifies the final len == totalElementsToEnqueue - totalElementsToRemove
//	4 - Verifies that final 2nd element == (1 + totalElementsToRemove)
func (suite *QueueTestSuite) TestRemoveMultipleGRs() {
	var (
		wg                     sync.WaitGroup
		totalElementsToEnqueue = 100
		totalElementsToRemove  = 90
	)

	for i := 0; i < totalElementsToEnqueue; i++ {
		suite.queue.Enqueue(testValue)
	}

	for i := 0; i < totalElementsToRemove; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := suite.queue.Remove(testValue)
			suite.Nil(err, "Unexpected error during concurrent Remove(n)")
		}()
	}
	wg.Wait()

	// check len, should be == totalElementsToEnqueue - totalElementsToRemove
	totalElementsAfterRemove := suite.queue.GetLen()
	suite.Equal(totalElementsToEnqueue-totalElementsToRemove, totalElementsAfterRemove, "Total elements on list does not match with expected number")

	// check current 2nd element (index 1) on the queue
	_, err := suite.queue.GetElement(1)
	suite.Nil(err, "No error should be returned when getting an existent element")
}

// ***************************************************************************************
// ** Dequeue
// ***************************************************************************************

// dequeue an empty queue
func (suite *QueueTestSuite) TestDequeueEmptyQueueSingleGR() {
	val, err := suite.queue.Dequeue()

	// error expected
	suite.Equal(ErrorQueueEmpty, err.ErrCode, "Can't dequeue an empty queue")

	// no value expected
	suite.Equal(nil, val, "Can't get a value different than nil from an empty queue")
}

// dequeue all elements
func (suite *QueueTestSuite) TestDequeueSingleGR() {
	suite.queue.Enqueue(testValue)
	suite.queue.Enqueue(5)

	// get the first element
	val, err := suite.queue.Dequeue()
	suite.Nil(err, "Unexpected error")
	suite.Equal(testValue, val, "Wrong element's value")
	length := suite.queue.GetLen()
	suite.Equal(1, length, "Incorrect number of queue elements")

	// get the second element
	val, err = suite.queue.Dequeue()
	suite.Nil(err, "Unexpected error")
	suite.Equal(5, val, "Wrong element's value")
	length = suite.queue.GetLen()
	suite.Equal(0, length, "Incorrect number of queue elements")

}

// TestDequeueMultipleGRs dequeues elements concurrently
//
// Detailed steps:
//	1 - Enqueues totalElementsToEnqueue consecutive integers
//	2 - Dequeues totalElementsToDequeue concurrently from totalElementsToDequeue GRs
//	3 - Verifies the final len, should be equal to totalElementsToEnqueue - totalElementsToDequeue
//	4 - Verifies that the next dequeued element's value is equal to totalElementsToDequeue
func (suite *QueueTestSuite) TestDequeueMultipleGRs() {
	var (
		wg                     sync.WaitGroup
		totalElementsToEnqueue = 100
		totalElementsToDequeue = 90
	)

	for i := 0; i < totalElementsToEnqueue; i++ {
		suite.queue.Enqueue(i)
	}

	for i := 0; i < totalElementsToDequeue; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := suite.queue.Dequeue()
			suite.Nil(err, "Unexpected error during concurrent Dequeue()")
		}()
	}
	wg.Wait()

	// check len, should be == totalElementsToEnqueue - totalElementsToDequeue
	totalElementsAfterDequeue := suite.queue.GetLen()
	suite.Equal(totalElementsToEnqueue-totalElementsToDequeue, totalElementsAfterDequeue, "Total elements on queue (after Dequeue) does not match with expected number")

	// check current first element
	val, err := suite.queue.Dequeue()
	suite.Nil(err, "No error should be returned when dequeuing an existent element")
	suite.Equalf(totalElementsToDequeue, val, "The expected last element's value should be: %v", totalElementsToEnqueue-totalElementsToDequeue)
}

// ***************************************************************************************
// ** DequeueOrWaitForNextElement
// ***************************************************************************************

// single GR DequeueOrWaitForNextElement with a previous enqueued element
func (suite *QueueTestSuite) TestDequeueOrWaitForNextElementWithEnqueuedElementSingleGR() {
	value := 100
	length := suite.queue.GetLen()
	suite.queue.Enqueue(value)

	result, err := suite.queue.DequeueOrWaitForNextElement()

	suite.Nil(err)
	suite.Equal(value, result)
	// length must be exactly the same as it was before
	suite.Equal(length, suite.queue.GetLen())
}

// single GR DequeueOrWaitForNextElement 1 element
func (suite *QueueTestSuite) TestDequeueOrWaitForNextElementWithEmptyQueue() {
	var (
		value  = 100
		result interface{}
		err    *ErrMsg
		done   = make(chan struct{})
	)

	// waiting for next enqueued element
	go func() {
		result, err = suite.queue.DequeueOrWaitForNextElement()
		done <- struct{}{}
	}()

	// enqueue an element
	go func() {
		suite.queue.Enqueue(value)
	}()

	select {
	// wait for the dequeued element
	case <-done:
		suite.Nil(err)
		suite.Equal(value, result)

	// the following comes first if more time than expected happened while waiting for the dequeued element
	case <-time.After(2 * time.Second):
		suite.Fail("too much time waiting for the enqueued element")

	}
}

// single GR calling DequeueOrWaitForNextElement (WaitForNextElementChanCapacity + 1) times, last one should return error
func (suite *QueueTestSuite) TestDequeueOrWaitForNextElementWithFullWaitingChannel() {
	// enqueue WaitForNextElementChanCapacity listeners to future enqueued elements
	for i := 0; i < WaitForNextElementChanCapacity; i++ {
		suite.queue.waitForNextElementChan <- make(chan interface{})
	}

	result, err := suite.queue.DequeueOrWaitForNextElement()
	suite.Nil(result)
	suite.Equal(ErrorQueueTooManyListeners, err.ErrCode)
}

// multiple GRs, calling DequeueOrWaitForNextElement from different GRs and enqueuing the expected values later
func (suite *QueueTestSuite) TestDequeueOrWaitForNextElementMultiGR() {
	var (
		wg sync.WaitGroup
		// channel to enqueue dequeued values
		dequeuedValues = make(chan int, WaitForNextElementChanCapacity)
		// map[dequeued_value] = times dequeued
		mp = make(map[int]int)
	)

	for i := 0; i < WaitForNextElementChanCapacity; i++ {
		go func() {
			// wait for the next enqueued element
			result, err := suite.queue.DequeueOrWaitForNextElement()
			// no error && no nil result
			suite.Nil(err)
			suite.NotNil(result)

			// send each dequeued element into the dequeuedValues channel
			resultInt, _ := result.(int)
			dequeuedValues <- resultInt

			// let the wg.Wait() know that this GR is done
			wg.Done()
		}()
	}

	// enqueue all needed elements
	for i := 0; i < WaitForNextElementChanCapacity; i++ {
		wg.Add(1)
		suite.queue.Enqueue(i)
		// save the enqueued value as index
		mp[i] = 0
	}

	// wait until all GRs dequeue the elements
	wg.Wait()
	// close dequeuedValues channel in order to only read the previous enqueued values (from the channel)
	close(dequeuedValues)

	// verify that all enqueued values were dequeued
	for v := range dequeuedValues {
		val, ok := mp[v]
		suite.Truef(ok, "element dequeued but never enqueued: %v", val)
		// increment the m[p] value meaning the value p was dequeued
		mp[v] = val + 1
	}
	// verify that there are no duplicates
	for k, v := range mp {
		suite.Equalf(1, v, "%v was dequeued %v times", k, v)
	}
}

// multiple GRs, calling DequeueOrWaitForNextElement from different GRs and enqueuing the expected values later
func (suite *QueueTestSuite) TestDequeueOrWaitForNextElementMultiGR2() {
	var (
		done       = make(chan int, 10)
		total      = 2000
		results    = make(map[int]struct{})
		totalOk    = 0
		totalError = 0
	)

	go func(queue *Queue, done chan int, total int) {
		for i := 0; i < total; i++ {
			go func(queue *Queue, done chan int) {
				rawValue, err := queue.DequeueOrWaitForNextElement()
				if err != nil {
					fmt.Println(err)
					// error
					done <- -1
				} else {
					val, _ := rawValue.(int)
					done <- val
				}
			}(queue, done)

			go func(queue *Queue, value int) {
				queue.Enqueue(value)
			}(queue, i)
		}
	}(suite.queue, done, total)

	i := 0
	for {
		v := <-done
		if v != -1 {
			totalOk++
			_, ok := results[v]
			suite.Falsef(ok, "duplicated value %v", v)

			results[v] = struct{}{}
		} else {
			totalError++
		}

		i++
		if i == total {
			break
		}
	}

	suite.Equal(total, totalError+totalOk)
}

// call DequeueOrWaitForNextElement(), wait some time and enqueue an item
func (suite *QueueTestSuite) TestDequeueOrWaitForNextElementGapSingleGR() {
	var (
		expectedValue = 50
		done          = make(chan struct{}, 3)
	)

	// DequeueOrWaitForNextElement()
	go func(queue *Queue, done chan struct{}) {
		val, err := queue.DequeueOrWaitForNextElement()
		suite.Nil(err)
		suite.Equal(expectedValue, val)
		done <- struct{}{}
	}(suite.queue, done)

	// wait and Enqueue function
	go func(queue *Queue, done chan struct{}) {
		time.Sleep(time.Millisecond * dequeueOrWaitForNextElementInvokeGapTime * dequeueOrWaitForNextElementInvokeGapTime)
		queue.Enqueue(expectedValue)
		done <- struct{}{}
	}(suite.queue, done)

	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-time.After(2 * time.Millisecond * dequeueOrWaitForNextElementInvokeGapTime * dequeueOrWaitForNextElementInvokeGapTime):
			suite.FailNow("Too much time waiting for the value")
		}
	}
}

// ***************************************************************************************
// ** Swap
// ***************************************************************************************
func (suite *QueueTestSuite) TestSwapEmptyQueue() {
	const (
		a = 4
		b = 4
	)

	err := suite.queue.Swap(a, b)
	suite.Equal(ErrorQueueEmpty, err.ErrCode)
}

func (suite *QueueTestSuite) TestSwapIndexesNotFound() {
	const (
		size = 2
		a    = 4
		b    = 3
	)

	for i := 0; i < size; i++ {
		suite.queue.Enqueue(i)
	}

	err := suite.queue.Swap(a, b)

	suite.Equal(ErrorIDNotFound, err.ErrCode)
}

func (suite *QueueTestSuite) TestSwapSameIndex() {
	const (
		size = 10
		a    = 4
		b    = 4
	)

	for i := 0; i < size; i++ {
		suite.queue.Enqueue(i)
	}

	err := suite.queue.Swap(a, b)

	suite.Equal(ErrorQueueSwapIndexesMatch, err.ErrCode)
}

func (suite *QueueTestSuite) TestSwapElements() {

	const (
		size = 10
		a    = 3
		b    = 4
	)

	for i := 0; i < size; i++ {
		suite.queue.Enqueue(i)
	}

	oldValueFromA, _ := suite.queue.GetElement(a)
	oldValueFromB, _ := suite.queue.GetElement(b)

	suite.queue.Swap(a, b)

	newValueFromA, _ := suite.queue.GetElement(a)
	newValueFromB, _ := suite.queue.GetElement(b)

	suite.Equal(oldValueFromA, newValueFromB)
	suite.Equal(oldValueFromB, newValueFromA)
}

// ***************************************************************************************
// ** Move elements inside an empty queue
// ***************************************************************************************
func (suite *QueueTestSuite) TestMoveToFrontEmptyQueue() {
	const (
		front = 10
	)

	suite.Equal(ErrorQueueEmpty, suite.queue.MoveToFront(front).ErrCode)
}

func (suite *QueueTestSuite) TestMoveToBackEmptyQueue() {
	const (
		back = 1
	)

	suite.Equal(ErrorQueueEmpty, suite.queue.MoveToBack(back).ErrCode)
}

// ***************************************************************************************
// ** Move elements to the front
// ***************************************************************************************
func (suite *QueueTestSuite) TestMoveToFront() {
	const (
		size  = 10
		front = 5
	)

	for i := 0; i < size; i++ {
		suite.queue.Enqueue(i + 1)
	}

	result := []interface{}{5, 1, 2, 3, 4, 6, 7, 8, 9, 10}

	suite.Nil(suite.queue.MoveToFront(front))
	suite.Equal(result, suite.queue.slice)
}

func (suite *QueueTestSuite) TestMoveToFrontAlreadyInFront() {
	const (
		size  = 10
		front = "sim-1"
	)

	for i := 1; i < size; i++ {
		item := fmt.Sprintf("sim-%d", i)
		suite.queue.Enqueue(item)
	}

	suite.Equal(ErrorQueueMoveIndexFrontPosition, suite.queue.MoveToFront(front).ErrCode)
}

func (suite *QueueTestSuite) TestMoveToFrontNotFound() {
	suite.queue.Enqueue(testValue)

	err := suite.queue.MoveToFront("anotherValue")

	suite.Equal(ErrorIDNotFound, err.ErrCode)
}

// ***************************************************************************************
// ** Move elements to the back
// ***************************************************************************************
func (suite *QueueTestSuite) TestMoveToBack() {
	const (
		size = 10
		back = 5
	)

	for i := 0; i < size; i++ {
		suite.queue.Enqueue(i + 1)
	}

	result := []interface{}{1, 2, 3, 4, 6, 7, 8, 9, 10, 5}

	suite.Nil(suite.queue.MoveToBack(back))
	suite.Equal(result, suite.queue.slice)
}

func (suite *QueueTestSuite) TestMoveToBackAlreadyInBack() {
	const (
		size = 10
		back = "sim-9"
	)

	for i := 1; i < size; i++ {
		item := fmt.Sprintf("sim-%d", i)
		suite.queue.Enqueue(item)
	}

	suite.Equal(ErrorQueueMoveIndexBackPosition, suite.queue.MoveToBack(back).ErrCode)
}

func (suite *QueueTestSuite) TestMoveToBackNotFound() {
	suite.queue.Enqueue(testValue)

	err := suite.queue.MoveToBack("anotherValue")

	suite.Equal(ErrorIDNotFound, err.ErrCode)
}

// ***************************************************************************************
// ** Get all elements
// ***************************************************************************************
func (suite *QueueTestSuite) TestGetAll() {
	const (
		size = 10
	)
	var slice []interface{}

	for i := 0; i < size; i++ {
		suite.queue.Enqueue(i)
		slice = append(slice, i)
	}

	result, err := suite.queue.GetElements()
	suite.Nil(err)
	suite.Equal(slice, result)
}

func (suite *QueueTestSuite) TestGetAllFiltered() {
	const (
		size = 10
	)
	var (
		offset = 4
		limit  = 2
	)
	var slice []interface{}
	for i := 1; i < size; i++ {
		suite.queue.Enqueue(i)
		slice = append(slice, i)
	}

	expected := []interface{}{5, 6}

	result, err := suite.queue.GetFilteredElements(offset, limit)
	suite.Nil(err)
	suite.Equal(slice, suite.queue.slice)
	suite.Equal(expected, result)
}

func (suite *QueueTestSuite) TestGetAllFilteredOutOfBounds() {
	const (
		size = 10
	)
	var (
		limit  = 2
		offset = 10
	)
	var slice []interface{}
	for i := 0; i < size; i++ {
		suite.queue.Enqueue(i)
		slice = append(slice, i)
	}

	_, err := suite.queue.GetFilteredElements(offset, limit)
	suite.Equal(ErrorQueueIndexOutOfBounds, err.ErrCode)
	suite.Equal(slice, suite.queue.slice)
}

func (suite *QueueTestSuite) TestGetAllFilteredWrongRange() {
	const (
		size = 10
	)
	var (
		limit  = -5
		offset = -3
	)
	var slice []interface{}
	for i := 0; i < size; i++ {
		suite.queue.Enqueue(i)
		slice = append(slice, i)
	}

	_, err := suite.queue.GetFilteredElements(offset, limit)
	suite.Equal(err.ErrCode, ErrorQueueIndexOutOfBounds)
	suite.Equal(slice, suite.queue.slice)
}

func (suite *QueueTestSuite) TestGetAllFilteredLimitOutOfBounds() {
	const (
		size = 10
	)
	var (
		limit  = 50
		offset = 5
	)

	for i := 1; i < size; i++ {
		suite.queue.Enqueue(i)
	}

	expected := []interface{}{6, 7, 8, 9}

	result, err := suite.queue.GetFilteredElements(offset, limit)
	suite.Nil(err)
	suite.Equal(expected, result)
}

// ***************************************************************************************
// ** Find by criteria
// ***************************************************************************************
func (suite *QueueTestSuite) TestQueueFindNotFound() {
	const (
		size = 10
	)

	for i := 1; i < size; i++ {
		item := fmt.Sprintf("sim-%d", i)
		suite.queue.Enqueue(item)
	}

	results := suite.queue.Find(func(element interface{}) bool {
		return element == "sim-11"
	})

	suite.Empty(results)
}

func (suite *QueueTestSuite) TestQueueFind() {
	const (
		size = 10
	)

	for i := 1; i < size; i++ {
		item := fmt.Sprintf("sim-%d", i)
		suite.queue.Enqueue(item)
	}

	expected := []int{7, 8}
	results := suite.queue.Find(func(element interface{}) bool {
		return element == "sim-8" || element == "sim-9"
	})

	suite.NotEmpty(results)
	suite.Equal(expected, results)
}

// ***************************************************************************************
// ** Find one by target
// ***************************************************************************************
func (suite *QueueTestSuite) TestQueueFindOneNotFound() {
	const (
		size = 10
	)

	for i := 1; i < size; i++ {
		item := fmt.Sprintf("sim-%d", i)
		suite.queue.Enqueue(item)
	}

	result := suite.queue.FindOne("sim-11")
	suite.Equal(-1, result)
}

func (suite *QueueTestSuite) TestQueueFindOne() {
	const (
		size = 10
	)

	for i := 1; i < size; i++ {
		item := fmt.Sprintf("sim-%d", i)
		suite.queue.Enqueue(item)
	}

	result := suite.queue.FindOne("sim-9")
	suite.Equal(8, result)
}

// ***************************************************************************************
// ** Find by IDS
// ***************************************************************************************
func (suite *QueueTestSuite) TestQueueFindByIDsNotFound() {
	const (
		size = 10
	)

	for i := 1; i < size; i++ {
		item := fmt.Sprintf("sim-%d", i)
		suite.queue.Enqueue(item)
	}

	ids := []int{10, 11}
	results := suite.queue.FindByIDs(ids)

	suite.Empty(results)
}

func (suite *QueueTestSuite) TestQueueFindByIDs() {
	const (
		size = 10
	)

	for i := 1; i < size; i++ {
		item := fmt.Sprintf("sim-%d", i)
		suite.queue.Enqueue(item)
	}

	ids := []int{7, 8}
	expected := []interface{}{"sim-8", "sim-9"}

	results := suite.queue.FindByIDs(ids)

	suite.NotEmpty(results)
	suite.Equal(expected, results)
}

func TestQueueTestSuite(t *testing.T) {
	suite.Run(t, new(QueueTestSuite))
}
