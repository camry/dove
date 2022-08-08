package cron

import (
    "context"
    "strings"
    "testing"
    "time"

    "github.com/camry/g/glog"
)

func TestWithLocation(t *testing.T) {
    c := New(WithLocation(time.UTC))
    if c.location != time.UTC {
        t.Errorf("expected UTC, got %v", c.location)
    }
}

func TestWithParser(t *testing.T) {
    var parser = NewParser(Dow)
    c := New(WithParser(parser))
    if c.parser != parser {
        t.Error("expected provided parser")
    }
}

func TestWithVerboseLogger(t *testing.T) {
    var buf syncWriter
    logger := glog.NewHelper(glog.NewStdLogger(&buf))
    c := New(WithLogger(logger))
    if c.logger != logger {
        t.Error("expected provided logger")
    }

    c.AddFunc("@every 1s", func() {})
    c.Start()
    time.Sleep(OneSecond)
    c.Stop(context.Background())
    out := buf.String()
    if !strings.Contains(out, "action=schedule") ||
        !strings.Contains(out, "action=run") {
        t.Error("expected to see some actions, got:", out)
    }
}
