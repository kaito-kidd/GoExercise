**基于 errgroup 实现一个 http server 的启动和关闭，以及 linux signal 信号的注册和处理，要保证能够 一个退出，全部注销退出。**

## 代码说明：

- 以下代码模拟了同时开启 HTTP Server 和 Background Worker 的场景
- 其中 HTTP Server 提供 Web 服务，Background Worker 在后台处理任务
- 如果 HTTP Server 或 Background Worker 有其一启动失败，服务自动退出
- 如果全部启动成功，服务可在收到 ctrl + c 命令 或 kill 信号时，优雅退出
- 如果退出超时，也会在规定时间内强制退出整个进程

### HTTP Server 模块

```go
type handler struct{}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

// Server HTTP Server
type Server struct {
	Addr string
	http *http.Server
}

// NewServer 创建一个 HTTP Server
func NewServer(addr string) *Server {
	mux := http.NewServeMux()
	mux.Handle("/", &handler{})
	http := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return &Server{
		Addr: addr,
		http: http,
	}
}

// Start 开始服务
func (s *Server) Start() error {
	log.Print("http server start")
	if err := s.http.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	log.Print("http server exit")
	return nil
}

// Shutdown 停止服务
func (s *Server) Shutdown(ctx context.Context) error {
	log.Print("http server shutdown")
	if err := s.http.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
```

### Backgound Worker 模块

```go
// BackgroundWorker 后台任务Worker
type BackgroundWorker struct {
	stop chan struct{}

	ctx    context.Context
	cancel context.CancelFunc
}

// NewBackgroundWorker 创建后台任务Worker
func NewBackgroundWorker() *BackgroundWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &BackgroundWorker{
		stop: make(chan struct{}),

		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 后台Worker开始工作
func (w *BackgroundWorker) Start() error {
	// 退出时通知调用方
	defer close(w.stop)

	log.Print("background worker start")

	// 定时器
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// 模拟后台任务
	for {
		select {
		case <-ticker.C:
			time.Sleep(500 * time.Millisecond)
			log.Printf("background do job")
		case <-w.ctx.Done():
			return nil
		}
	}
}

// Stop 停止后台任务
func (w *BackgroundWorker) Stop(ctx context.Context) error {
	// 让后台任务停下来
	w.cancel()

	select {
	case <-w.stop: // 等待成功退出
	case <-ctx.Done(): // 外部也可以强制退
    	log.Print("background worker exit with context done")
		return ctx.Err()
	}

	log.Print("background worker exit")
	return nil
}
```

### `main` 主流程

```go
// Stop 停止 HTTP Server 和 后台任务
func Stop(s *Server, w *BackgroundWorker) error {
	g, _ := errgroup.WithContext(context.Background())

	// 设置5秒强制退出
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 停止后台任务
	g.Go(func() error {
		return w.Stop(ctx)
	})

	// 关闭 HTTP Server
	g.Go(func() error {
		return s.Shutdown(ctx)
	})

	// 等待全部完成
	return g.Wait()
}

func main() {
	g, ctx := errgroup.WithContext(context.Background())

	// 启动 HTTP Server
	s := NewServer(":8080")
	g.Go(func() error {
		return s.Start()
	})

	// 启动后台任务
	w := NewBackgroundWorker()
	g.Go(func() error {
		return w.Start()
	})

	// 监听退出信号
	g.Go(func() error {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-sigs:
			// 收到关闭信号
			log.Printf("signal caught: %s, ready to quit...", sig.String())
		case <-ctx.Done():
			// HTTP Server 或 后台任务启动失败 停止监听信号
			log.Println("monitor signal exit")
		}

		// 停止 HTTP Server 和 后台任务
		return Stop(s, w)
	})

	// 等待所有 goroutine 退出
	if err := g.Wait(); err != nil {
		log.Printf("server exit: %+v", err)
		return
	}

	// 服务优雅关闭
	log.Println("server graceful exit")
}
```

### 编译运行

```shell
# 编译
go build -o main
# 运行
./main
```

### 启动服务，发送 `kill $pid` 命令

服务优雅退出，输出日志如下：

```shell
2020/12/09 02:07:51 background worker start
2020/12/09 02:07:51 http server start
2020/12/09 02:07:53 background do job
2020/12/09 02:07:55 background do job
2020/12/09 02:07:57 background do job
2020/12/09 02:07:58 signal caught: terminated, ready to quit...
2020/12/09 02:07:58 http server shutdown
2020/12/09 02:07:58 background worker exit
2020/12/09 02:07:58 http server exit
2020/12/09 02:07:58 server graceful exit
```
