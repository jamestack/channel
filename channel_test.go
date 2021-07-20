package channel

import (
	"fmt"
	"testing"
	"time"
)

func TestChannel(t *testing.T) {
	ch := NewRingBuffer(1)
	//fmt.Println(ch.Close())

	fmt.Println(ch.Get())
	fmt.Println(ch.Put(1))
	fmt.Println(ch.Put(2))
	fmt.Println(ch.Put(3))
	fmt.Println(ch.Get())
	fmt.Println(ch.Get())
	fmt.Println(ch.Get())
	fmt.Println(ch.Put(3))
	fmt.Println(ch.Put(4))
	fmt.Println(ch.Put(5))
	fmt.Println(ch.Put(6))
	fmt.Println(ch.Close())
	fmt.Println(ch.Get())
	fmt.Println(ch.Get())
	fmt.Println(ch.Get())
	fmt.Println(ch.Get())
	fmt.Println(ch.Get())

}

func benchSysChanel(num int64, size int) (res int64) {
	ch := make(chan int64, size)
	done := make(chan bool)
	go func() {
		for range ch {
			res += 1
		}
		done <- true
	}()
	for i := int64(1); i <= num; i++ {
		ch <- i
	}
	close(ch)

	<-done
	return
}

func benchSnowChannel(num int64, size int) (res int64) {
	ch := NewRingBuffer(size)
	done := make(chan bool)
	go func() {
		for {
			// <-time.After(100 * time.Millisecond)
			_, err := ch.Receive()
			if err != nil {
				done <- true
				return
			}

			res += 1
		}
	}()

	for i := int64(1); i <= num; i++ {
		// <-time.After(10 * time.Millisecond)
		ch.Send(i)
	}
	ch.Close()

	<-done
	return
}

func benchTwoSizedChannel(num int64, size int) {
	fmt.Println("benchSysChanel start")
	start := time.Now()
	fmt.Println(benchSysChanel(num, size))
	fmt.Println("benchSysChanel end", time.Since(start))

	fmt.Println("benchSnowChannel start")
	start = time.Now()
	fmt.Println(benchSnowChannel(num, size))
	fmt.Println("benchSnowChannel end", time.Since(start))
}

func benchTwoNoSizedChannel(num int64) {
	fmt.Println("benchSysChanel start 1024")
	start := time.Now()
	fmt.Println(benchSysChanel(num, 1024))
	fmt.Println("benchSysChanel end", time.Since(start))

	fmt.Println("benchSnowChannel start")
	start = time.Now()
	fmt.Println(benchSnowChannel(num, 0))
	fmt.Println("benchSnowChannel end", time.Since(start))
}

// sized channel test
func TestBenchSizedChannel(t *testing.T) {
	fmt.Println("-----------bench 10000 1w 1---------")
	benchTwoSizedChannel(10000, 1)
	fmt.Println("-----------bench 100000 10w 10---------")
	benchTwoSizedChannel(100000, 10)
	fmt.Println("-----------bench 10000000 1000w 100---------")
	benchTwoSizedChannel(10000000, 100)
	fmt.Println("-----------bench 100000000 1e 1k---------")
	benchTwoSizedChannel(100000000, 1000)
	// fmt.Println("-----------bench 1000000000 10e 10k---------")
	// benchTwoSizedChannel(1000000000, 10000)
}

// no limited channel test
func TestBenchNoSizedChannel(t *testing.T) {
	fmt.Println("-----------bench 1000 1w---------")
	benchTwoNoSizedChannel(10000)
	fmt.Println("-----------bench 100000 10w---------")
	benchTwoNoSizedChannel(100000)
	fmt.Println("-----------bench 10000000 1000w---------")
	benchTwoNoSizedChannel(10000000)
	fmt.Println("-----------bench 100000000 1e---------")
	benchTwoNoSizedChannel(100000000)
	// fmt.Println("-----------bench 1000000000 10e---------")
	// benchTwoNoSizedChannel(1000000000)
}
