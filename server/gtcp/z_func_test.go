package gtcp_test

import (
    "reflect"
    "testing"

    "github.com/camry/dove/server/gtcp"
)

func TestGetFreePort(t *testing.T) {
    _, err := gtcp.GetFreePort()
    if err != nil {
        t.Error(err)
    }
}

func TestGetFreePorts(t *testing.T) {
    ports, err := gtcp.GetFreePorts(2)
    if err != nil {
        t.Error(err)
    }
    if !reflect.DeepEqual(len(ports), 2) {
        t.Fatalf("len(ports):%d is not equal to v:2", len(ports))
    }
}
