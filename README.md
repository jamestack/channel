# Unlimited-buffer Channel For Go
Unlimited-buffer Channel for Go, based on ring buffer

Support Method:
1. Put() / Get() / Peek()  &nbsp;&nbsp;                     # not block, return res or error
2. Send() / Receive()      &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;   # block until success or close, return res or closed

## Example:
```go
func TestChannel(t *testing.T) {
    var res int64
	ch := NewChannel()
	done := make(chan bool)
	go func() {
		for {
			_,ok := ch.Receive()
			if !ok {
				done <- true
				return
			}

			res += 1
		}
	}()

	for i:=int64(1);i <= 1000000;i++ {
		ch.Send(i)
	}
	ch.Close()

	<-done
	return
}
```

## Performance:
performance about go native channel ±5%, test code in channel_test.go

## [中文文档](https://github.com/jamestack/channel/blob/main/README.zh_cn.md)
