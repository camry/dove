package cron

import "time"

// SpecSchedule 根据传统的 crontab 规范指定占空比（到第二个粒度）。
// 它最初被计算并存储为位集。
type SpecSchedule struct {
    Second, Minute, Hour, Dom, Month, Dow uint64

    // 覆盖此计划的时区。
    Location *time.Location
}

// bounds 提供一系列可接受的值（加上名称到值的映射）。
type bounds struct {
    min, max uint
    names    map[string]uint
}

// 每个字段的边界。
var (
    seconds = bounds{0, 59, nil}
    minutes = bounds{0, 59, nil}
    hours   = bounds{0, 23, nil}
    dom     = bounds{1, 31, nil}
    months  = bounds{1, 12, map[string]uint{
        "jan": 1,
        "feb": 2,
        "mar": 3,
        "apr": 4,
        "may": 5,
        "jun": 6,
        "jul": 7,
        "aug": 8,
        "sep": 9,
        "oct": 10,
        "nov": 11,
        "dec": 12,
    }}
    dow = bounds{0, 6, map[string]uint{
        "sun": 0,
        "mon": 1,
        "tue": 2,
        "wed": 3,
        "thu": 4,
        "fri": 5,
        "sat": 6,
    }}
)

const (
    // 如果表达式中包含星号，则设置最高位。
    starBit = 1 << 63
)

// Next 返回下一次激活此计划的时间，大于指定时间。
// 如果找不到时间来满足计划，则返回零时间。
func (s *SpecSchedule) Next(t time.Time) time.Time {
    // 一般的做法
    //
    // 月、日、时、分、秒：
    // 检查时间值是否匹配。 如果是，请继续下一个字段。
    // 如果该字段与计划不匹配，则递增该字段直到匹配。
    // 在增加字段时，环绕将其带回到开头
    // 字段列表的（因为需要重新验证之前的字段值）

    // 如果指定了，则将指定时间转换为时间表的时区。
    // 保存原始时区，以便我们在找到时间后转换回来。
    // 请注意，未指定时区 (time.Local) 的时间表将被处理
    // 作为提供时间的本地时间。
    origLocation := t.Location()
    loc := s.Location
    if loc == time.Local {
        loc = t.Location()
    }
    if s.Location != time.Local {
        t = t.In(s.Location)
    }

    // 尽早开始（即将到来的第二个）。
    t = t.Add(1*time.Second - time.Duration(t.Nanosecond())*time.Nanosecond)

    // 此标志指示字段是否已递增。
    added := false

    // 如果五年内没有找到时间，则返回零。
    yearLimit := t.Year() + 5

WRAP:
    if t.Year() > yearLimit {
        return time.Time{}
    }

    // 查找第一个适用的月份。
    // 如果是这个月，那么什么都不做。
    for 1<<uint(t.Month())&s.Month == 0 {
        // 如果我们必须添加一个月，请将其他部分重置为 0。
        if !added {
            added = true
            // 否则，将日期设置在开头（因为当前时间无关紧要）。
            t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
        }
        t = t.AddDate(0, 1, 0)

        // 包起来。
        if t.Month() == time.January {
            goto WRAP
        }
    }

    // 现在得到那个月的一天。
    //
    // 注意：这会导致夏令时出现问题
    // 不存在。 例如：圣保罗有 DST，它在午夜变换
    // 11/3 到凌晨 1 点。 通过注意小时结束的时间来处理它!=0。
    for !dayMatches(s, t) {
        if !added {
            added = true
            t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
        }
        t = t.AddDate(0, 0, 1)
        // 注意由于 DST，小时是否不再是午夜。
        // 如果是 23 则加一小时，如果是 1 则减一小时。
        if t.Hour() != 0 {
            if t.Hour() > 12 {
                t = t.Add(time.Duration(24-t.Hour()) * time.Hour)
            } else {
                t = t.Add(time.Duration(-t.Hour()) * time.Hour)
            }
        }

        if t.Day() == 1 {
            goto WRAP
        }
    }

    for 1<<uint(t.Hour())&s.Hour == 0 {
        if !added {
            added = true
            t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, loc)
        }
        t = t.Add(1 * time.Hour)

        if t.Hour() == 0 {
            goto WRAP
        }
    }

    for 1<<uint(t.Minute())&s.Minute == 0 {
        if !added {
            added = true
            t = t.Truncate(time.Minute)
        }
        t = t.Add(1 * time.Minute)

        if t.Minute() == 0 {
            goto WRAP
        }
    }

    for 1<<uint(t.Second())&s.Second == 0 {
        if !added {
            added = true
            t = t.Truncate(time.Second)
        }
        t = t.Add(1 * time.Second)

        if t.Second() == 0 {
            goto WRAP
        }
    }

    return t.In(origLocation)
}

// dayMatches 如果指定时间满足计划的星期几和每月几日的限制，则返回 true。
func dayMatches(s *SpecSchedule, t time.Time) bool {
    var (
        domMatch = 1<<uint(t.Day())&s.Dom > 0
        dowMatch = 1<<uint(t.Weekday())&s.Dow > 0
    )
    if s.Dom&starBit > 0 || s.Dow&starBit > 0 {
        return domMatch && dowMatch
    }
    return domMatch || dowMatch
}
