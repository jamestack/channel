package channel

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCond(t *testing.T) {
	cond := sync.NewCond(&sync.Mutex{})

	mail := 0

	go func() {
		cond.L.Lock()
		for mail != 3 {
			cond.Wait()
		}
		cond.L.Unlock()
		fmt.Println("mail == 3")

	}()

	<-time.After(1 * time.Second)

	for i := 0; i < 10; i++ {
		cond.L.Lock()
		mail = i
		<-time.After(1 * time.Second)
		cond.Broadcast()
		cond.L.Unlock()
	}

	<-time.After(3 * time.Second)

}

type Queue struct {
	cond *sync.Cond
	data []interface{}
	capc int
}

func NewQueue(capacity int) *Queue {
	return &Queue{cond: &sync.Cond{L: &sync.Mutex{}}, data: make([]interface{}, 0), capc: capacity}
}

func (q *Queue) Enqueue(d interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for len(q.data) == q.capc {
		q.cond.Wait()
	}
	// FIFO入队
	q.data = append(q.data, d)
	// 通知其他waiter进行Dequeue或Enqueue操作
	q.cond.Broadcast()

}

func (q *Queue) Dequeue() (d interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for len(q.data) == 0 {
		q.cond.Wait()
	}
	// FIFO出队
	d = q.data[0]
	q.data = q.data[1:]
	// 通知其他waiter进行Dequeue或Enqueue操作
	q.cond.Broadcast()
	return
}

func (q *Queue) Len() int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return len(q.data)
}
