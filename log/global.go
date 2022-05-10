package log

import (
    "sync"
)

// globalLogger 被设计为当前进程中的全局日志记录器。
var global = &loggerAppliance{}

// loggerAppliance 是 `Logger` 的代理器，使 logger 更改将影响所有子 logger.
type loggerAppliance struct {
    lock sync.Mutex
    Logger
    helper *Helper
}

// init 初始化全局默认 Logger。
func init() {
    global.SetLogger(DefaultLogger)
}

// SetLogger 配置 Logger。
func (a *loggerAppliance) SetLogger(in Logger) {
    a.lock.Lock()
    defer a.lock.Unlock()
    a.Logger = in
    a.helper = NewHelper(a.Logger)
}

// GetLogger 获取 Logger。
func (a *loggerAppliance) GetLogger() Logger {
    return a.Logger
}

// SetLogger 应该在任何其他日志调用之前调用。
// 而且它不是线程安全的。
func SetLogger(logger Logger) {
    global.SetLogger(logger)
}

// GetLogger 将全局日志记录器器设备作为当前进程中的记录器返回。
func GetLogger() Logger {
    return global
}

// Log 按级别和键值打印日志。
func Log(level Level, keyvals ...any) {
    global.helper.Log(level, keyvals...)
}

// Debug 打印调试级别的日志。
func Debug(a ...any) {
    global.helper.Debug(a...)
}

// Debugf 按 fmt.Sprintf 格式打印调试级别的日志。
func Debugf(format string, a ...any) {
    global.helper.Debugf(format, a...)
}

// Debugw 按键值对打印调试级别的日志。
func Debugw(keyvals ...any) {
    global.helper.Debugw(keyvals...)
}

// Info 打印信息级别的日志。
func Info(a ...any) {
    global.helper.Info(a...)
}

// Infof 按 fmt.Sprintf 格式打印信息级别的日志。
func Infof(format string, a ...any) {
    global.helper.Infof(format, a...)
}

// Infow 按键值对打印信息级别的日志。
func Infow(keyvals ...any) {
    global.helper.Infow(keyvals...)
}

// Warn 打印警告级别的日志。
func Warn(a ...any) {
    global.helper.Warn(a...)
}

// Warnf 按 fmt.Sprintf 格式打印警告级别的日志。
func Warnf(format string, a ...any) {
    global.helper.Warnf(format, a...)
}

// Warnw 按键值对打印警告级别的日志。
func Warnw(keyvals ...any) {
    global.helper.Warnw(keyvals...)
}

// Error 打印错误级别的日志。
func Error(a ...any) {
    global.helper.Error(a...)
}

// Errorf 按 fmt.Sprintf 格式打印错误级别的日志。
func Errorf(format string, a ...any) {
    global.helper.Errorf(format, a...)
}

// Errorw 按键值对打印错误级别的日志。
func Errorw(keyvals ...any) {
    global.helper.Errorw(keyvals...)
}

// Fatal 打印致命级别的日志。
func Fatal(a ...any) {
    global.helper.Fatal(a...)
}

// Fatalf 按 fmt.Sprintf 格式打印致命级别的日志。
func Fatalf(format string, a ...any) {
    global.helper.Fatalf(format, a...)
}

// Fatalw 按键值对打印致命级别的日志。
func Fatalw(keyvals ...any) {
    global.helper.Fatalw(keyvals...)
}
