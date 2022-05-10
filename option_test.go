package dove

import (
    "context"
    "os"
    "reflect"
    "testing"
    "time"
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

type mockSignal struct{}

func (m *mockSignal) String() string { return "sig" }
func (m *mockSignal) Signal()        {}

func TestSignal(t *testing.T) {
    o := &option{}
    v := []os.Signal{
        &mockSignal{}, &mockSignal{},
    }
    Signals(v...)(o)
    if !reflect.DeepEqual(v, o.signals) {
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
