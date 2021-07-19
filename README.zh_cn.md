# Unlimited-buffer Channel For Go
无限容量管道, 基于Ring结构

支持的方法:
1. Put() / Get() / Peek()  &nbsp;&nbsp;                     # 不阻塞执行，返回值或者错误           列表特性
2. Send() / Receive()      &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;   # 阻塞直到成功, 返回值或者错误         管道特性

## 例子:
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

## 性能:
性能大约为原生channel的±5%,测试详情见channel_test.go

## [English Document](https://github.com/jamestack/channel/blob/main/README.md)
