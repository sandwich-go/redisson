package redisson

import (
	"testing"

	"github.com/redis/rueidis"
)

func TestGeoPosCmd_BaseAccess(t *testing.T) {
	c := &geoPosCmd{}
	c.SetVal([]*GeoPos{{Longitude: 13.5, Latitude: 41.9}, nil})
	v := c.Val()
	if len(v) != 2 {
		t.Fatalf("len=%d", len(v))
	}
	if v[0] == nil || v[0].Longitude != 13.5 {
		t.Errorf("0=%+v", v[0])
	}
	if v[1] != nil {
		t.Errorf("1 should be nil, got %+v", v[1])
	}
}

func TestGeoLocationCmd_BaseAccess(t *testing.T) {
	c := &geoLocationCmd{}
	c.SetVal([]rueidis.GeoLocation{{Name: "p1"}})
	if v := c.Val(); len(v) != 1 || v[0].Name != "p1" {
		t.Errorf("got %+v", v)
	}
}

func TestKeyValueSliceCmd_BaseAccess(t *testing.T) {
	c := &keyValueSliceCmd{}
	c.SetVal([]KeyValue{{Key: "k", Value: "v"}})
	if v := c.Val(); len(v) != 1 || v[0].Key != "k" {
		t.Errorf("got %+v", v)
	}
}

func TestCommandsInfoCmd_BaseAccess(t *testing.T) {
	c := &commandsInfoCmd{}
	c.SetVal(map[string]CommandInfo{
		"GET": {Name: "GET", Arity: 2, ReadOnly: true},
	})
	v := c.Val()
	if len(v) != 1 || v["GET"].Arity != 2 || !v["GET"].ReadOnly {
		t.Errorf("got %+v", v)
	}
}
