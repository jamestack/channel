package channel

import (
	"fmt"
	"math"
	"sync"
)

type Any interface{}

type RingBuffer struct {
	lock   sync.Mutex
	close  bool
	List   []Any
	RIndex int
	WIndex int
	Cap    int
	maxCap int

	receiveCh       chan Any
	sendCh          chan struct{}
	needSendSign    int32
	needReceiveSign int32
}

func NewRingBuffer(size ...int) *RingBuffer {
	// lock := &sync.Mutex{}
	rb := &RingBuffer{
		List:      make([]Any, 0),
		receiveCh: make(chan Any),
		sendCh:    make(chan struct{}),
	}

	if len(size) > 0 && size[0] > 0 {
		rb.maxCap = size[0]
	} else {
		rb.maxCap = math.MaxInt32
	}

	return rb
}

func (r *RingBuffer) get(block bool) (v Any, err Err) {
	r.lock.Lock()
	if r.RIndex < r.WIndex {
		v = r.List[r.RIndex%r.Cap]
		r.List[r.RIndex%r.Cap] = nil
		r.RIndex += 1

		if r.needSendSign > 0 {
			r.needSendSign -= 1

			r.lock.Unlock()
			r.sendCh <- struct{}{}
			return
		}

		r.lock.Unlock()
		return
	}

	if r.close {

		r.lock.Unlock()
		return nil, ErrClosed
	}

	if !block {

		r.lock.Unlock()
		return nil, ErrEmpty
	}

	// 等待数据

	r.needReceiveSign += 1

	r.lock.Unlock()
	v = <-r.receiveCh

	return
}

func (r *RingBuffer) put(v Any, block bool) (err Err) {
	r.lock.Lock()
	if r.close {
		r.lock.Unlock()
		return ErrClosed
	}

	if r.WIndex-r.RIndex >= r.Cap && r.Cap < r.maxCap {
		if r.Cap == 0 {
			// fmt.Println("addCap", 1)
			r.Cap = 1
			r.List = make([]Any, 1)
		} else {
			var addCap int

			if r.Cap < 1024 {
				addCap = r.Cap + 1
			} else {
				addCap = r.Cap/4 + 1
			}

			if addCap+r.Cap > r.maxCap {
				addCap = addCap + r.Cap - r.maxCap
			}

			// fmt.Println("addCap", addCap)

			r.List = append(r.List[r.RIndex%r.Cap:], r.List[:r.RIndex%r.Cap]...)
			r.List = append(r.List, make([]Any, addCap)...)
			r.RIndex = 0
			r.WIndex = r.Cap
			r.Cap += addCap
		}
	}

	if r.WIndex-r.RIndex < r.Cap {
		if r.needReceiveSign > 0 {
			r.needReceiveSign -= 1

			r.lock.Unlock()

			r.receiveCh <- v
			return nil
		} else {
			r.List[r.WIndex%r.Cap] = v
			r.WIndex += 1
		}

		r.lock.Unlock()
		return nil
	}

	if !block {
		r.lock.Unlock()
		return ErrNotEnoughSpace
	}

	r.needSendSign += 1
	r.lock.Unlock()
	// 等待信号
	<-r.sendCh

	return r.put(v, block)
}

func (r *RingBuffer) Put(v Any) (err Err) {
	return r.put(v, false)
}

func (r *RingBuffer) Get() (v Any, err Err) {
	return r.get(false)
}

func (r *RingBuffer) Send(v Any) (err Err) {
	return r.put(v, true)
}

func (r *RingBuffer) Receive() (v Any, err Err) {
	v, err = r.get(true)
	fmt.Println("R", v)
	return
}

func (r *RingBuffer) Close() (ok bool) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.close {
		return false
	}

	// 更新关闭状态
	r.close = true

	return true
}
