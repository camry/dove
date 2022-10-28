package gudp_test

import (
    "reflect"
    "testing"

    "github.com/camry/dove/server/gudp"
)

func TestGetFreePort(t *testing.T) {
    _, err := gudp.GetFreePort()
    if err != nil {
        t.Error(err)
    }
}

func TestGetFreePorts(t *testing.T) {
    ports, err := gudp.GetFreePorts(2)
    if err != nil {
        t.Error(err)
    }
    if !reflect.DeepEqual(len(ports), 2) {
        t.Fatalf("len(ports):%d is not equal to v:2", len(ports))
    }
}
