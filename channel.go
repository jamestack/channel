package channel

import (
	"errors"
	"math"
	"sync"
)

// 无限缓存Channel
type Channel struct {
	lock            sync.Mutex
	nw              *sRing
	nr              *sRing
	receiveCh       chan interface{}
	sendCh          chan struct{}
	close           bool
	cap             int
	addCap          int
	maxCap          int
	needSendSign    int32
	needReceiveSign int32
	chanLock        sync.Mutex
}

// 模拟无限缓存的Channel，size为最大容量，如果size为0则为无限容量
func NewChannel(size ...int) *Channel {
	ch := &Channel{}
	if len(size) > 0 {
		ch.maxCap = size[0]
	}
	return ch
}

func (q *Channel) getReceiveChan() chan interface{} {
	if q.receiveCh == nil {
		q.chanLock.Lock()
		if q.receiveCh == nil {
			q.receiveCh = make(chan interface{})
		}
		q.chanLock.Unlock()
	}
	return q.receiveCh
}

func (q *Channel) getSendChan() chan struct{} {
	if q.sendCh == nil {
		q.chanLock.Lock()
		if q.sendCh == nil {
			q.sendCh = make(chan struct{})
		}
		q.chanLock.Unlock()
	}
	return q.sendCh
}

type sRing struct {
	value interface{}
	next  *sRing
}

type Err error

var (
	ErrEmpty          Err = errors.New("[Channel: Empty]")
	ErrNotEnoughSpace Err = errors.New("[Channel: NotEnoughSpace]")
	ErrClosed         Err = errors.New("[Channel: Closed]")
)

func (q *Channel) put(v interface{}, block bool) (err Err) {
	if q.close {
		return ErrClosed
	}

	if q.nw == nil {
		q.lock.Lock()
		if q.nw == nil {
			head := &sRing{}
			head.next = head
			q.nw = head
			q.nr = head
			q.cap = 0
			q.addCap = 0
			if q.maxCap <= 0 {
				q.maxCap = math.MaxInt32
			}
		}
		q.lock.Unlock()
	}

	q.lock.Lock()
	if q.nw.next == q.nr {
		if q.cap < q.maxCap {
			if q.cap < 1024 {
				if q.cap == 0 {
					q.addCap = 1
				} else {
					q.addCap = q.cap
				}
			} else {
				q.addCap = q.cap / 4
			}
			if q.cap+q.addCap > q.maxCap {
				q.addCap = q.maxCap - q.cap
			}
			q.addCaps(q.addCap)
		} else {
			if !block {
				q.lock.Unlock()
				return ErrNotEnoughSpace
			}
			q.needSendSign += 1
			q.lock.Unlock()

			q.getSendChan() <- struct{}{}

			q.lock.Lock()
		}
	}

	if q.needReceiveSign > 0 {
		q.needReceiveSign -= 1

		q.lock.Unlock()

		q.getReceiveChan() <- v
	} else {
		q.nw.value = v
		q.nw = q.nw.next

		q.lock.Unlock()
	}

	return nil
}

func (q *Channel) Put(v interface{}) (err Err) {
	return q.put(v, false)
}

// 向Channel发送数据，如果达到最大容量则阻塞
func (q *Channel) Send(v interface{}) (err Err) {
	return q.put(v, true)
}

func (q *Channel) addCaps(addCap int) {
	list := make([]sRing, addCap)
	for i := 0; i < addCap-1; i++ {
		list[i].next = &list[i+1]
	}

	list[addCap-1].next = q.nw.next
	q.nw.next = &list[0]

	q.cap += addCap
}

func (q *Channel) Cap() int {
	return q.cap
}

// 阻塞直到取到值或者该队列关闭，模拟channel取值符，如果ok为false则表明此channel已关闭
func (q *Channel) Receive() (v interface{}, err Err) {
	return q.get(true)
}

// 关闭队列并回收队列
func (q *Channel) Close() (ok bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.close {
		return false
	}

	// 更新关闭状态
	q.close = true

	return true
}

func (q *Channel) get(block bool) (v interface{}, err Err) {
	q.lock.Lock()

	if q.nr == q.nw {
		if q.close {
			q.lock.Unlock()
			err = ErrClosed
		} else {
			if block {
				q.needReceiveSign += 1
				q.lock.Unlock()
				v = <-q.getReceiveChan()
			} else {
				q.lock.Unlock()
				err = ErrEmpty
			}
		}
		return
	}

	v = q.nr.value
	q.nr.value = nil
	q.nr = q.nr.next

	if q.needSendSign > 0 {
		q.needSendSign -= 1

		q.lock.Unlock()
		<-q.getSendChan()
	} else {
		q.lock.Unlock()
	}

	// 完全关闭后，自动回收内存
	if q.close && q.nr == q.nw {
		q.gc()
	}
	return
}

// 从Chanel中取数据，不阻塞，如果队列为空或者已关闭则返回错误
func (q *Channel) Get() (v interface{}, err Err) {
	return q.get(false)
}

func (q *Channel) Peek() (v interface{}, err Err) {
	q.lock.Lock()

	if q.nr == q.nw {
		if q.close {
			q.lock.Unlock()
			err = ErrClosed
		} else {
			q.lock.Unlock()
			err = ErrEmpty
		}
		return
	}

	v = q.nr.value
	q.lock.Unlock()
	return
}

// 内存回收，避免已close的queue占用大量内存无法被Runtime使用
func (q *Channel) gc() {
	// fmt.Println("gc cap", q.cap)
	if q.nw != nil && q.nr != nil {
		var next *sRing
		for n := q.nw; n.next != q.nw && n.next != q.nr; n = n.next {
			next = n.next
			q.cap -= 1
		}
		if next != nil {
			q.nw.next = next
		}
	}

	if q.receiveCh != nil {
		close(q.receiveCh)
	}
	if q.sendCh != nil {
		close(q.sendCh)
	}
}
