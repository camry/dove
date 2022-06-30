package cron

import (
    "fmt"
    "runtime"
    "sync"
    "time"

    "github.com/camry/dove/log"
)

// JobWrapper 用一些行为装饰指定的 Job。
type JobWrapper func(Job) Job

// Chain 是一个 JobWrappers 序列，用于装饰提交的任务。
// cross-cutting behaviors like logging or synchronization.
type Chain struct {
    wrappers []JobWrapper
}

// NewChain 横切行为，如日志记录或同步。
func NewChain(c ...JobWrapper) Chain {
    return Chain{c}
}

// Then 用链中的所有 JobWrapper 装饰指定的任务。
//
// 这个:
//     NewChain(m1, m2, m3).Then(job)
// 相当于:
//     m1(m2(m3(job)))
func (c Chain) Then(j Job) Job {
    for i := range c.wrappers {
        j = c.wrappers[len(c.wrappers)-i-1](j)
    }
    return j
}

// Recover 使用日志记录器，记录包装任务中的 panic。
func Recover(logger *log.Helper) JobWrapper {
    return func(j Job) Job {
        return FuncJob(func() {
            defer func() {
                if r := recover(); r != nil {
                    const size = 64 << 10
                    buf := make([]byte, size)
                    buf = buf[:runtime.Stack(buf, false)]
                    err, ok := r.(error)
                    if !ok {
                        err = fmt.Errorf("%v", r)
                    }
                    logger.Error(err, "panic", "stack", "...\n"+string(buf))
                }
            }()
            j.Run()
        })
    }
}

// DelayIfStillRunning 序列化作业，延迟后续运行，直到前一个完成。延迟超过一分钟后运行的作业会在信息中记录延迟。
func DelayIfStillRunning(logger *log.Helper) JobWrapper {
    return func(j Job) Job {
        var mu sync.Mutex
        return FuncJob(func() {
            start := time.Now()
            mu.Lock()
            defer mu.Unlock()
            if dur := time.Since(start); dur > time.Minute {
                logger.Infow(log.DefaultMessageKey, "Cron", "action", "delay", "duration", dur)
            }
            j.Run()
        })
    }
}

// SkipIfStillRunning 如果先前的调用仍在运行，则跳过对 Job 的调用。它记录跳转到信息级别的给定记录器。
func SkipIfStillRunning(logger *log.Helper) JobWrapper {
    return func(j Job) Job {
        var ch = make(chan struct{}, 1)
        ch <- struct{}{}
        return FuncJob(func() {
            select {
            case v := <-ch:
                defer func() { ch <- v }()
                j.Run()
            default:
                logger.Info("skip")
            }
        })
    }
}
