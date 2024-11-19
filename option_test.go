package dove

import (
    "context"
    "log"
    "os"
    "reflect"
    "testing"
    "time"

    "github.com/camry/g/glog"
)

func TestID(t *testing.T) {
    o := &option{}
    v := "123"
    ID(v)(o)
    if !reflect.DeepEqual(v, o.id) {
        t.Fatalf("o.id:%s is not equal to v:%s", o.id, v)
    }
}

func TestName(t *testing.T) {
    o := &option{}
    v := "abc"
    Name(v)(o)
    if !reflect.DeepEqual(v, o.name) {
        t.Fatalf("o.name:%s is not equal to v:%s", o.name, v)
    }
}

func TestVersion(t *testing.T) {
    o := &option{}
    v := "123"
    Version(v)(o)
    if !reflect.DeepEqual(v, o.version) {
        t.Fatalf("o.version:%s is not equal to v:%s", o.version, v)
    }
}

func TestContext(t *testing.T) {
    type ctxKey = struct{}
    o := &option{}
    v := context.WithValue(context.TODO(), ctxKey{}, "b")
    Context(v)(o)
    if !reflect.DeepEqual(v, o.ctx) {
        t.Fatalf("o.ctx:%s is not equal to v:%s", o.ctx, v)
    }
}

func TestLogger(t *testing.T) {
    o := &option{}
    v := glog.NewStdLogger(log.Writer())
    Logger(v)(o)
    if !reflect.DeepEqual(v, o.logger) {
        t.Fatalf("o.logger:%v is not equal to xlog.NewHelper(v):%v", o.logger, glog.NewHelper(v))
    }
}

type mockSignal struct{}

func (m *mockSignal) String() string { return "sig" }
func (m *mockSignal) Signal()        {}

func TestSignal(t *testing.T) {
    o := &option{}
    v := []os.Signal{
        &mockSignal{}, &mockSignal{},
    }
    Signal(v...)(o)
    if !reflect.DeepEqual(v, o.sigs) {
        t.Fatal("o.sigs is not equal to v")
    }
}

func TestStopTimeout(t *testing.T) {
    o := &option{}
    v := time.Duration(123)
    StopTimeout(v)(o)
    if !reflect.DeepEqual(v, o.stopTimeout) {
        t.Fatal("o.stopTimeout is not equal to v")
    }
}

func TestBeforeStart(t *testing.T) {
    o := &option{}
    v := func(_ context.Context) error {
        t.Log("BeforeStart...")
        return nil
    }
    BeforeStart(v)(o)
}

func TestBeforeStop(t *testing.T) {
    o := &option{}
    v := func(_ context.Context) error {
        t.Log("BeforeStop...")
        return nil
    }
    BeforeStop(v)(o)
}

func TestAfterStart(t *testing.T) {
    o := &option{}
    v := func(_ context.Context) error {
        t.Log("AfterStart...")
        return nil
    }
    AfterStart(v)(o)
}

func TestAfterStop(t *testing.T) {
    o := &option{}
    v := func(_ context.Context) error {
        t.Log("AfterStop...")
        return nil
    }
    AfterStop(v)(o)
}
