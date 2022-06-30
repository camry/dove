package cron

import (
    "fmt"
    "math"
    "strconv"
    "strings"
    "time"
)

// ParseOption 用于创建解析器的配置选项。
// 大多数选项指定应包含哪些字段，而其他选项则启用功能。
// 如果未包含字段，则解析器将采用默认值。
// 这些选项不会更改解析的顺序字段。
type ParseOption int

const (
    Second         ParseOption = 1 << iota // 秒，默认 0
    SecondOptional                         // 可选秒，默认0
    Minute                                 // 分，默认 0
    Hour                                   // 时，默认 0
    Dom                                    // 日，默认 *
    Month                                  // 月，默认 *
    Dow                                    // 周，默认 *
    DowOptional                            // 可选周，默认 *
    Descriptor                             // 允许使用 @monthly、@weekly 等描述符。
)

var places = []ParseOption{
    Second,
    Minute,
    Hour,
    Dom,
    Month,
    Dow,
}

var defaults = []string{
    "0",
    "0",
    "0",
    "*",
    "*",
    "*",
}

// Parser 可以配置的自定义解析器。
type Parser struct {
    options ParseOption
}

// NewParser 使用自定义选项创建解析器。
//
// 如果给出了多个 Optional，它会出现 panic，因为通常无法正确推断提供或缺少哪个可选。
//
// 例子
//
//  // 没有描述符的标准解析器
//  specParser := NewParser(Minute | Hour | Dom | Month | Dow)
//  sched, err := specParser.Parse("0 0 15 */3 *")
//
//  // 同上，只是不包括时间字段
//  specParser := NewParser(Dom | Month | Dow)
//  sched, err := specParser.Parse("15 */3 *")
//
//  // 同上，只是使 Dow 可选
//  specParser := NewParser(Dom | Month | DowOptional)
//  sched, err := specParser.Parse("15 */3")
//
func NewParser(options ParseOption) Parser {
    optionals := 0
    if options&DowOptional > 0 {
        optionals++
    }
    if options&SecondOptional > 0 {
        optionals++
    }
    if optionals > 1 {
        panic("multiple optionals may not be configured")
    }
    return Parser{options}
}

// Parse 返回代表指定规范的新 crontab 计划。
// 如果规范无效，则返回描述性错误。
// 它接受由 NewParser 配置的 crontab 规范和特性。
func (p Parser) Parse(spec string) (Schedule, error) {
    if len(spec) == 0 {
        return nil, fmt.Errorf("empty spec string")
    }

    // 提取时区（如果存在）
    var loc = time.Local
    if strings.HasPrefix(spec, "TZ=") || strings.HasPrefix(spec, "CRON_TZ=") {
        var err error
        i := strings.Index(spec, " ")
        eq := strings.Index(spec, "=")
        if loc, err = time.LoadLocation(spec[eq+1 : i]); err != nil {
            return nil, fmt.Errorf("provided bad location %s: %v", spec[eq+1:i], err)
        }
        spec = strings.TrimSpace(spec[i:])
    }

    // 处理命名计划（描述符），如果已配置
    if strings.HasPrefix(spec, "@") {
        if p.options&Descriptor == 0 {
            return nil, fmt.Errorf("parser does not accept descriptors: %v", spec)
        }
        return parseDescriptor(spec, loc)
    }

    // 在空白处拆分。
    fields := strings.Fields(spec)

    // 验证并填写任何省略或可选的字段
    var err error
    fields, err = normalizeFields(fields, p.options)
    if err != nil {
        return nil, err
    }

    field := func(field string, r bounds) uint64 {
        if err != nil {
            return 0
        }
        var bits uint64
        bits, err = getField(field, r)
        return bits
    }

    var (
        second     = field(fields[0], seconds)
        minute     = field(fields[1], minutes)
        hour       = field(fields[2], hours)
        dayofmonth = field(fields[3], dom)
        month      = field(fields[4], months)
        dayofweek  = field(fields[5], dow)
    )
    if err != nil {
        return nil, err
    }

    return &SpecSchedule{
        Second:   second,
        Minute:   minute,
        Hour:     hour,
        Dom:      dayofmonth,
        Month:    month,
        Dow:      dayofweek,
        Location: loc,
    }, nil
}

// normalizeFields 获取时间字段的子集并返回完整的集合，其中填充了未设置字段的默认值（零）。
//
// 作为执行此功能的一部分，它还验证提供的字段是否与配置的选项兼容。
func normalizeFields(fields []string, options ParseOption) ([]string, error) {
    // 验证选项并将其字段添加到选项
    optionals := 0
    if options&SecondOptional > 0 {
        options |= Second
        optionals++
    }
    if options&DowOptional > 0 {
        options |= Dow
        optionals++
    }
    if optionals > 1 {
        return nil, fmt.Errorf("multiple optionals may not be configured")
    }

    // 弄清楚我们需要多少个字段
    max := 0
    for _, place := range places {
        if options&place > 0 {
            max++
        }
    }
    min := max - optionals

    // 验证字段数
    if count := len(fields); count < min || count > max {
        if min == max {
            return nil, fmt.Errorf("expected exactly %d fields, found %d: %s", min, count, fields)
        }
        return nil, fmt.Errorf("expected %d to %d fields, found %d: %s", min, max, count, fields)
    }

    // 如果未提供，则填充可选字段
    if min < max && len(fields) == min {
        switch {
        case options&DowOptional > 0:
            fields = append(fields, defaults[5]) // TODO: improve access to default
        case options&SecondOptional > 0:
            fields = append([]string{defaults[0]}, fields...)
        default:
            return nil, fmt.Errorf("unknown optional field")
        }
    }

    // 使用默认值填充不属于选项的所有字段
    n := 0
    expandedFields := make([]string, len(places))
    copy(expandedFields, defaults)
    for i, place := range places {
        if options&place > 0 {
            expandedFields[i] = fields[n]
            n++
        }
    }
    return expandedFields, nil
}

var standardParser = NewParser(
    Minute | Hour | Dom | Month | Dow | Descriptor,
)

// ParseStandard 返回代表指定的新 crontab 计划。
// 标准规范 (https://en.wikipedia.org/wiki/Cron)。
// 它需要 5 个条目，依次代表：分钟、小时、月中的某天、月和周中的某天。
// 如果规范无效，它会返回描述性错误。
//
// 它接受
// - 标准 crontab 规范，例如 “* * * * ？”
// - 描述符，例如 “@midnight”、“@每 1 小时 30 分”
func ParseStandard(standardSpec string) (Schedule, error) {
    return standardParser.Parse(standardSpec)
}

// getField 返回一个 Int，其位设置表示该字段表示的所有时间或错误解析字段值。
// “字段”是逗号分隔的“范围”列表。
func getField(field string, r bounds) (uint64, error) {
    var bits uint64
    ranges := strings.FieldsFunc(field, func(r rune) bool { return r == ',' })
    for _, expr := range ranges {
        bit, err := getRange(expr, r)
        if err != nil {
            return bits, err
        }
        bits |= bit
    }
    return bits, nil
}

// getRange 返回指定表达式指示的位：
// 数字 | 数字“-”数字[“/”数字]
// 或错误解析范围。
func getRange(expr string, r bounds) (uint64, error) {
    var (
        start, end, step uint
        rangeAndStep     = strings.Split(expr, "/")
        lowAndHigh       = strings.Split(rangeAndStep[0], "-")
        singleDigit      = len(lowAndHigh) == 1
        err              error
    )

    var extra uint64
    if lowAndHigh[0] == "*" || lowAndHigh[0] == "?" {
        start = r.min
        end = r.max
        extra = starBit
    } else {
        start, err = parseIntOrName(lowAndHigh[0], r.names)
        if err != nil {
            return 0, err
        }
        switch len(lowAndHigh) {
        case 1:
            end = start
        case 2:
            end, err = parseIntOrName(lowAndHigh[1], r.names)
            if err != nil {
                return 0, err
            }
        default:
            return 0, fmt.Errorf("too many hyphens: %s", expr)
        }
    }

    switch len(rangeAndStep) {
    case 1:
        step = 1
    case 2:
        step, err = mustParseInt(rangeAndStep[1])
        if err != nil {
            return 0, err
        }

        // 特殊处理：“N/step”表示“N-max/step”。
        if singleDigit {
            end = r.max
        }
        if step > 1 {
            extra = 0
        }
    default:
        return 0, fmt.Errorf("too many slashes: %s", expr)
    }

    if start < r.min {
        return 0, fmt.Errorf("beginning of range (%d) below minimum (%d): %s", start, r.min, expr)
    }
    if end > r.max {
        return 0, fmt.Errorf("end of range (%d) above maximum (%d): %s", end, r.max, expr)
    }
    if start > end {
        return 0, fmt.Errorf("beginning of range (%d) beyond end of range (%d): %s", start, end, expr)
    }
    if step == 0 {
        return 0, fmt.Errorf("step of range should be a positive number: %s", expr)
    }

    return getBits(start, end, step) | extra, nil
}

// parseIntOrName 返回 expr 中包含的（可能命名的）整数。
func parseIntOrName(expr string, names map[string]uint) (uint, error) {
    if names != nil {
        if namedInt, ok := names[strings.ToLower(expr)]; ok {
            return namedInt, nil
        }
    }
    return mustParseInt(expr)
}

// mustParseInt 将指定的表达式解析为 int 或返回错误。
func mustParseInt(expr string) (uint, error) {
    num, err := strconv.Atoi(expr)
    if err != nil {
        return 0, fmt.Errorf("failed to parse int from %s: %s", expr, err)
    }
    if num < 0 {
        return 0, fmt.Errorf("negative number (%d) not allowed: %s", num, expr)
    }

    return uint(num), nil
}

// getBits 设置 [min, max] 范围内的所有位，以指定步长取模。
func getBits(min, max, step uint) uint64 {
    var bits uint64

    if step == 1 {
        return ^(math.MaxUint64 << (max + 1)) & (math.MaxUint64 << min)
    }

    for i := min; i <= max; i += step {
        bits |= 1 << i
    }
    return bits
}

// all 返回指定范围内的所有位。（加上星位）
func all(r bounds) uint64 {
    return getBits(r.min, r.max, 1) | starBit
}

// parseDescriptor 返回表达式的预定义计划，如果没有匹配则返回错误。
func parseDescriptor(descriptor string, loc *time.Location) (Schedule, error) {
    switch descriptor {
    case "@yearly", "@annually":
        return &SpecSchedule{
            Second:   1 << seconds.min,
            Minute:   1 << minutes.min,
            Hour:     1 << hours.min,
            Dom:      1 << dom.min,
            Month:    1 << months.min,
            Dow:      all(dow),
            Location: loc,
        }, nil

    case "@monthly":
        return &SpecSchedule{
            Second:   1 << seconds.min,
            Minute:   1 << minutes.min,
            Hour:     1 << hours.min,
            Dom:      1 << dom.min,
            Month:    all(months),
            Dow:      all(dow),
            Location: loc,
        }, nil

    case "@weekly":
        return &SpecSchedule{
            Second:   1 << seconds.min,
            Minute:   1 << minutes.min,
            Hour:     1 << hours.min,
            Dom:      all(dom),
            Month:    all(months),
            Dow:      1 << dow.min,
            Location: loc,
        }, nil

    case "@daily", "@midnight":
        return &SpecSchedule{
            Second:   1 << seconds.min,
            Minute:   1 << minutes.min,
            Hour:     1 << hours.min,
            Dom:      all(dom),
            Month:    all(months),
            Dow:      all(dow),
            Location: loc,
        }, nil

    case "@hourly":
        return &SpecSchedule{
            Second:   1 << seconds.min,
            Minute:   1 << minutes.min,
            Hour:     all(hours),
            Dom:      all(dom),
            Month:    all(months),
            Dow:      all(dow),
            Location: loc,
        }, nil

    }

    const every = "@every "
    if strings.HasPrefix(descriptor, every) {
        duration, err := time.ParseDuration(descriptor[len(every):])
        if err != nil {
            return nil, fmt.Errorf("failed to parse duration %s: %s", descriptor, err)
        }
        return Every(duration), nil
    }

    return nil, fmt.Errorf("unrecognized descriptor: %s", descriptor)
}
