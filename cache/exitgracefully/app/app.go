package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type App struct {
	servers []*Server

	// 优雅退出整个超时时间，默认30秒
	shutdownTimeout time.Duration

	// 优雅退出时候等待处理已有请求时间，默认10秒钟
	waitTime time.Duration
	// 自定义回调超时时间，默认三秒钟
	cbTimeout time.Duration

	cbs []ShutdownCallback
}

type Option func(*App)

type ShutdownCallback func(ctx context.Context)

func WithShutdownCallbacks(cbs ...ShutdownCallback) Option {
	return func(app *App) {
		app.cbs = cbs
	}
}

// NewApp 创建 App 实例，注意设置默认值，同时使用这些选项
func NewApp(servers []*Server, opts ...Option) *App {
	app := &App{
		servers:         servers,
		shutdownTimeout: 30 * time.Second,
		waitTime:        10 * time.Second,
		cbTimeout:       3 * time.Second,
	}
	for _, opt := range opts {
		opt(app)
	}
	return app
}

// StartAndServe 你主要要实现这个方法
func (app *App) StartAndServe() {
	for _, s := range app.servers {
		srv := s
		go func() {
			if err := srv.Start(); err != nil {
				if err == http.ErrServerClosed {
					log.Printf("服务器%s已关闭", srv.name)
				} else {
					log.Printf("服务器%s异常退出", srv.name)
				}
			}
		}()
	}
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs,
		os.Interrupt, os.Kill, syscall.SIGKILL, syscall.SIGSTOP,
		syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP,
		syscall.SIGABRT, syscall.SIGSYS, syscall.SIGTERM,
	)
	<-sigs
	go func() {
		select {
		case <-sigs:
			fmt.Println("再次收到关闭信号，强制退出")
			os.Exit(1)
		case <-time.After(app.shutdownTimeout):
			fmt.Println("超时强制退出")
			os.Exit(1)
		}
	}()
	// 从这里开始优雅退出监听系统信号，强制退出以及超时强制退出。
	// 优雅退出的具体步骤在 shutdown 里面实现
	// 所以你需要在这里恰当的位置，调用 shutdown
	app.shutdown()
}

// shutdown 你要设计这里面的执行步骤。
func (app *App) shutdown() {
	log.Println("开始关闭应用，停止接收新请求")
	// 你需要在这里让所有的 server 拒绝新请求
	for _, server := range app.servers {
		server.rejectReq()
	}
	log.Println("等待正在执行请求完结")
	// 在这里等待一段时间
	time.Sleep(app.waitTime)
	log.Println("开始关闭服务器")
	wg := sync.WaitGroup{}
	wg.Add(len(app.servers))
	for _, s := range app.servers {
		go func(server *Server) {
			if err := server.stop(context.Background()); err != nil {
				fmt.Errorf("关闭服务%s错误，错误原因%w", server.name, err)
			}
			wg.Done()
		}(s)
	}
	wg.Wait()
	// 并发关闭服务器，同时要注意协调所有的 server 都关闭之后才能步入下一个阶段
	log.Println("开始执行自定义回调")
	// 并发执行回调，要注意协调所有的回调都执行完才会步入下一个阶段
	wg.Add(len(app.cbs))
	for _, cb := range app.cbs {
		go func(cb ShutdownCallback) {
			ctx, cancel := context.WithTimeout(context.Background(), app.cbTimeout)
			defer cancel()
			cb(ctx)
			wg.Done()
		}(cb)
	}
	wg.Wait()
	// 释放资源
	log.Println("开始释放资源")

	// 这一个步骤不需要你干什么，这是假装我们整个应用自己要释放一些资源
	app.close()
}

func (app *App) close() {
	// 在这里释放掉一些可能的资源
	time.Sleep(time.Second)
	log.Println("应用关闭")
}
