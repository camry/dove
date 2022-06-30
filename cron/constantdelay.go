package cron

import "time"

// ConstantDelaySchedule 表示一个简单的循环占空比，例如 “每 5 分钟”。
// 它不支持比每秒一次更频繁的任务。
type ConstantDelaySchedule struct {
    Delay time.Duration
}

// Every 返回每个持续时间激活一次的 crontab 计划。
// 不支持小于一秒的延迟（将四舍五入到 1 秒）。
// 任何小于秒的字段都会被截断。
func Every(duration time.Duration) ConstantDelaySchedule {
    if duration < time.Second {
        duration = time.Second
    }
    return ConstantDelaySchedule{
        Delay: duration - time.Duration(duration.Nanoseconds())%time.Second,
    }
}

// Next 返回下次应该的运行时间。
func (schedule ConstantDelaySchedule) Next(t time.Time) time.Time {
    return t.Add(schedule.Delay - time.Duration(t.Nanosecond())*time.Nanosecond)
}
