package cron

import (
    "context"
    "github.com/camry/dove/log"
    "sort"
    "sync"
    "time"
)

// Cron 跟踪任意数量的条目，调用调度指定的关联函数。
// Cron 可以启动、停止，并且可以在运行时检查条目。
type Cron struct {
    entries   []*Entry
    chain     Chain
    stop      chan struct{}
    add       chan *Entry
    remove    chan EntryID
    snapshot  chan chan []Entry
    running   bool
    logger    *log.Helper
    runningMu sync.Mutex
    location  *time.Location
    parser    ScheduleParser
    nextID    EntryID
    jobWaiter sync.WaitGroup
}

// ScheduleParser 返回 Schedule 的调度规范解析器的接口。
type ScheduleParser interface {
    Parse(spec string) (Schedule, error)
}

// Job 提交 cron 任务的接口。
type Job interface {
    Run()
}

// Schedule 描述一个任务的工作周期。
type Schedule interface {
    // Next 返回下一个激活时间，晚于给定时间。
    // Next 最初调用，然后每次运行作业时调用。
    Next(time.Time) time.Time
}

// EntryID 标识 Cron 实例中的条目。
type EntryID int

// Entry 由一个时间表和在该时间表上执行的函数组成。
type Entry struct {
    // ID 此条目的 cron 分配的 ID，可用于查找快照或删除它。
    ID EntryID

    // Schedule 运行此任务的计划。
    Schedule Schedule

    // Next 任务将运行的时间，或者如果 Cron 尚未启动或此条目的计划无法满足，则为零时间。
    Next time.Time

    // Prev 上次运行此作业的时间，如果从未运行，则为零时间。
    Prev time.Time

    // WrappedJob 激活调度时要运行的任务。
    WrappedJob Job

    // Job 提交给 cron 的任务。
    Job Job
}

// Valid 如果这不是零条目，则返回 true。
func (e Entry) Valid() bool { return e.ID != 0 }

// byTime 是按时间对条目数组进行排序的包装器（最后时间为零）。
type byTime []*Entry

func (s byTime) Len() int      { return len(s) }
func (s byTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byTime) Less(i, j int) bool {
    // 两个零次应该返回 false。
    // 否则，零比任何其他时间都“更大”。
    // （将其排序在列表的末尾。）
    if s[i].Next.IsZero() {
        return false
    }
    if s[j].Next.IsZero() {
        return true
    }
    return s[i].Next.Before(s[j].Next)
}

// New 返回由给定选项修改的新 Cron 任务运行器。
//
// 可用设置
//
//   TimeZone
//     描述: 解释时间表的时区
//     默认: time.Local
//
//   Parser
//     描述: Parser 将 cron 规范字符串转换为 cron.Schedules。
//     默认: 接受此规范：https://en.wikipedia.org/wiki/Cron
//
//   Chain
//     描述: 以自定义行为包装提交任务。
//     默认: 恢复 panic 并将其记录到标准错误 stderr。
//
// 请参阅“cron.With*”以修改默认行为。
func New(opts ...Option) *Cron {
    c := &Cron{
        entries:   nil,
        chain:     NewChain(),
        add:       make(chan *Entry),
        stop:      make(chan struct{}),
        snapshot:  make(chan chan []Entry),
        remove:    make(chan EntryID),
        running:   false,
        runningMu: sync.Mutex{},
        logger:    log.NewHelper(log.GetLogger()),
        location:  time.Local,
        parser:    standardParser,
    }
    for _, opt := range opts {
        opt(c)
    }
    return c
}

// FuncJob 是将 func() 转换为 cron.Job 的包装器
type FuncJob func()

func (f FuncJob) Run() { f() }

// AddFunc 向 Cron 添加一个函数，按给定的时间表运行。
// 使用此 Cron 实例的时区作为默认值解析规范。
// 返回一个不透明的 ID，以后可以使用它来移除它。
func (c *Cron) AddFunc(spec string, cmd func()) (EntryID, error) {
    return c.AddJob(spec, FuncJob(cmd))
}

// AddJob adds a Job to the Cron to be run on the given schedule.
// The spec is parsed using the time zone of this Cron instance as the default.
// An opaque ID is returned that can be used to later remove it.
func (c *Cron) AddJob(spec string, cmd Job) (EntryID, error) {
    schedule, err := c.parser.Parse(spec)
    if err != nil {
        return 0, err
    }
    return c.Schedule(schedule, cmd), nil
}

// Schedule 将任务添加至 Cron 按指定的时间表运行。
// 该任务使用配置的 Chain 进行包装。
func (c *Cron) Schedule(schedule Schedule, cmd Job) EntryID {
    c.runningMu.Lock()
    defer c.runningMu.Unlock()
    c.nextID++
    entry := &Entry{
        ID:         c.nextID,
        Schedule:   schedule,
        WrappedJob: c.chain.Then(cmd),
        Job:        cmd,
    }
    if !c.running {
        c.entries = append(c.entries, entry)
    } else {
        c.add <- entry
    }
    return entry.ID
}

// Entries 返回 cron 条目的快照。
func (c *Cron) Entries() []Entry {
    c.runningMu.Lock()
    defer c.runningMu.Unlock()
    if c.running {
        replyChan := make(chan []Entry, 1)
        c.snapshot <- replyChan
        return <-replyChan
    }
    return c.entrySnapshot()
}

// Location 获取时区。
func (c *Cron) Location() *time.Location {
    return c.location
}

// Entry 返回给定条目的快照，如果找不到，则返回 nil。
func (c *Cron) Entry(id EntryID) Entry {
    for _, entry := range c.Entries() {
        if id == entry.ID {
            return entry
        }
    }
    return Entry{}
}

// Remove 移除将要运行的条目。
func (c *Cron) Remove(id EntryID) {
    c.runningMu.Lock()
    defer c.runningMu.Unlock()
    if c.running {
        c.remove <- id
    } else {
        c.removeEntry(id)
    }
}

// Start 启动 cron 调度程序，如果已经启动则无需操作。
func (c *Cron) Start() {
    c.runningMu.Lock()
    defer c.runningMu.Unlock()
    if c.running {
        return
    }
    c.running = true
    go c.run()
}

// Run 运行 cron 调度程序, 如果已经运行则无需操作。
func (c *Cron) Run() {
    c.runningMu.Lock()
    if c.running {
        c.runningMu.Unlock()
        return
    }
    c.running = true
    c.runningMu.Unlock()
    c.run()
}

// run 调度程序.. 这是私有的，只是因为需要同步对“运行”状态变量的访问。
func (c *Cron) run() {
    // c.logger.Info("start")

    // 计算出每个条目的下一个激活时间。
    now := c.now()
    for _, entry := range c.entries {
        entry.Next = entry.Schedule.Next(now)
        // c.logger.Infow(log.DefaultMessageKey, "Cron", "action", "schedule", "now", now, "entry", entry.ID, "next", entry.Next)
    }

    for {
        // 确定要运行的下一个条目。
        sort.Sort(byTime(c.entries))

        var timer *time.Timer
        if len(c.entries) == 0 || c.entries[0].Next.IsZero() {
            // 如果还没有条目，只需休眠 - 它仍会处理新条目并停止请求。
            timer = time.NewTimer(100000 * time.Hour)
        } else {
            timer = time.NewTimer(c.entries[0].Next.Sub(now))
        }

        for {
            select {
            case now = <-timer.C:
                now = now.In(c.location)
                // c.logger.Infow(log.DefaultMessageKey, "Cron", "action", "wake", "now", now)

                // 运行下一次小于现在的每个条目
                for _, e := range c.entries {
                    if e.Next.After(now) || e.Next.IsZero() {
                        break
                    }
                    c.startJob(e.WrappedJob)
                    e.Prev = e.Next
                    e.Next = e.Schedule.Next(now)
                    // c.logger.Infow(log.DefaultMessageKey, "Cron", "action", "run", "now", now, "entry", e.ID, "next", e.Next)
                }

            case newEntry := <-c.add:
                timer.Stop()
                now = c.now()
                newEntry.Next = newEntry.Schedule.Next(now)
                c.entries = append(c.entries, newEntry)
                // c.logger.Infow(log.DefaultMessageKey, "Cron", "action", "added", "now", now, "entry", newEntry.ID, "next", newEntry.Next)

            case replyChan := <-c.snapshot:
                replyChan <- c.entrySnapshot()
                continue

            case <-c.stop:
                timer.Stop()
                // c.logger.Info("stop")
                return

            case id := <-c.remove:
                timer.Stop()
                now = c.now()
                c.removeEntry(id)
                // c.logger.Infow(log.DefaultMessageKey, "Cron", "action", "removed", "entry", id)
            }

            break
        }
    }
}

// startJob 在新的 goroutine 中运行指定的任务。
func (c *Cron) startJob(j Job) {
    c.jobWaiter.Add(1)
    go func() {
        defer c.jobWaiter.Done()
        j.Run()
    }()
}

// now 从 c location 获取当前时间。
func (c *Cron) now() time.Time {
    return time.Now().In(c.location)
}

// Stop 如果 cron 调度程序正在运行，则停止它； 否则它什么也不做。
// 传入一个上下文，以便调用者可以等待正在运行的作业完成。
func (c *Cron) Stop(ctx context.Context) context.Context {
    c.runningMu.Lock()
    defer c.runningMu.Unlock()
    if c.running {
        c.stop <- struct{}{}
        c.running = false
    }
    ctx, cancel := context.WithCancel(ctx)
    go func() {
        c.jobWaiter.Wait()
        cancel()
    }()

    return ctx
}

// entrySnapshot 返回当前 cron 条目列表的副本。
func (c *Cron) entrySnapshot() []Entry {
    var entries = make([]Entry, len(c.entries))
    for i, e := range c.entries {
        entries[i] = *e
    }
    return entries
}

// removeEntry 移除当前 cron 指定的条目。
func (c *Cron) removeEntry(id EntryID) {
    var entries []*Entry
    for _, e := range c.entries {
        if e.ID != id {
            entries = append(entries, e)
        }
    }
    c.entries = entries
}
