package dove

import (
    "reflect"
    "testing"
    "time"
)

func TestNew(t *testing.T) {
    app := New(
        Name("dove"),
        Version("v1.0.0"),
    )
    time.AfterFunc(time.Second, func() {
        _ = app.Stop()
    })
    if err := app.Run(); err != nil {
        t.Fatal(err)
    }
}

func TestApp_ID(t *testing.T) {
    v := "123"
    o := New(ID(v))
    if !reflect.DeepEqual(v, o.ID()) {
        t.Fatalf("o.ID():%s is not equal to v:%s", o.ID(), v)
    }
}

func TestApp_Name(t *testing.T) {
    v := "123"
    o := New(Name(v))
    if !reflect.DeepEqual(v, o.Name()) {
        t.Fatalf("o.Name():%s is not equal to v:%s", o.Name(), v)
    }
}

func TestApp_Version(t *testing.T) {
    v := "123"
    o := New(Version(v))
    if !reflect.DeepEqual(v, o.Version()) {
        t.Fatalf("o.Version():%s is not equal to v:%s", o.Version(), v)
    }
}
