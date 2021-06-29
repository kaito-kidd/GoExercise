**用 Go 实现一个 tcp server，用两个 goroutine 读写 conn，两个 goroutine 通过 chan 可以传递 message，能够正确退出。**

## 代码

```go
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	addr    = ":9090"
	quitMsg = "quit"
)

func main() {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	log.Printf("server start with %s", addr)

	wg := &sync.WaitGroup{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 捕捉退出信号
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigs
		log.Printf("signal caught: %s, ready to quit...", sig.String())

		cancel()
		listener.Close()
	}()

	for {
		// 新连接
		conn, err := listener.Accept()
		if err != nil {
			break
		}

		// 处理连接
		wg.Add(1)
		go handleConn(conn, wg, ctx)
	}

	// 等待所有 client 处理完成后退出
	wg.Wait()

	log.Print("server exit")
}

func handleConn(conn net.Conn, w *sync.WaitGroup, ctx context.Context) {
	defer w.Done()

	connAddr := conn.RemoteAddr()

	log.Printf("%s conn accepted", connAddr)
	defer log.Printf("%s conn closed", connAddr)

	wg := &sync.WaitGroup{}

	ch := make(chan []byte, 1024)

	wg.Add(2)

	// 处理读写
	go handleRead(conn, ch, wg, ctx)
	go handleWrite(conn, ch, wg, ctx)

	// 等待连接关闭
	wg.Wait()
}

func handleRead(conn net.Conn, ch chan<- []byte, wg *sync.WaitGroup, ctx context.Context) {
	connAddr := conn.RemoteAddr()

	defer log.Printf("%s read exit", connAddr)

	defer wg.Done()

	// 通知 write 协程退出
	defer close(ch)

	rd := bufio.NewReader(conn)

	for {
		msg, _, err := rd.ReadLine()
		if err != nil {
			return
		}

		// 收到退出指令
		if string(msg) == quitMsg {
			log.Printf("%s receive quit cmd", connAddr)
			return
		}

		select {
		case ch <- msg:
		case <-ctx.Done():
			return
		}
	}
}

func handleWrite(conn net.Conn, ch <-chan []byte, wg *sync.WaitGroup, ctx context.Context) {
	connAddr := conn.RemoteAddr()

	defer log.Printf("%s write exit", connAddr)

	defer wg.Done()

	// write 协程退出关闭 client 连接
	defer conn.Close()

	wr := bufio.NewWriter(conn)

	for {
		select {
		case line, ok := <-ch:
			if !ok {
				return
			}

			if len(line) <= 0 {
				continue
			}

			// echo msg
			msg := fmt.Sprintf("echo %s\n", string(line))
			wr.WriteString(msg)
			wr.Flush()
		case <-ctx.Done():
			return
		}
	}
}
```

## 测试

- 启动 TCP Server：

```shell
go run main.go

2021/01/28 15:24:32 server start with :9090
```

使用 `telnet` 连接 TCP Server 测试，其中发送 `quit` 命令可正常关闭连接：

- 客户端1

```shell
➜  ~ telnet 127.0.0.1 9090
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
hello
echo hello
Go
echo Go
quit
Connection closed by foreign host.
```

- 客户端2

```shell
➜  ~ telnet 127.0.0.1 9090
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
zhangsan
echo zhangsan
lisi
echo lisi
quit
Connection closed by foreign host.
```

- Server 输出日志

```shell
2021/01/28 15:24:32 server start with :9090
2021/01/28 15:25:43 127.0.0.1:63473 conn accepted
2021/01/28 15:26:16 127.0.0.1:63855 conn accepted
2021/01/28 15:27:38 127.0.0.1:63473 receive quit cmd
2021/01/28 15:27:38 127.0.0.1:63473 read exit
2021/01/28 15:27:38 127.0.0.1:63473 write exit
2021/01/28 15:27:38 127.0.0.1:63473 conn closed
2021/01/28 15:27:40 127.0.0.1:63855 receive quit cmd
2021/01/28 15:27:40 127.0.0.1:63855 read exit
2021/01/28 15:27:40 127.0.0.1:63855 write exit
2021/01/28 15:27:40 127.0.0.1:63855 conn closed
```

- Server 执行 Ctrl + C，正常停机：

```shell
^C2021/01/28 15:28:25 signal caught: interrupt, ready to quit...
2021/01/28 15:28:25 server exit
```
