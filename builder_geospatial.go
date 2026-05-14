package redisson

import (
	"fmt"
	"strconv"
	"strings"
)

func (b builder) GeoAddCompleted(key string, geoLocation ...GeoLocation) Completed {
	cmd := b.Geoadd().Key(key).LongitudeLatitudeMember()
	for _, loc := range geoLocation {
		cmd = cmd.LongitudeLatitudeMember(loc.Longitude, loc.Latitude, loc.Name)
	}
	return cmd.Build()
}

func (b builder) GeoDistCompleted(key, member1, member2, unit string) Completed {
	switch strings.ToUpper(unit) {
	case M:
		return b.Geodist().Key(key).Member1(member1).Member2(member2).M().Build()
	case MI:
		return b.Geodist().Key(key).Member1(member1).Member2(member2).Mi().Build()
	case FT:
		return b.Geodist().Key(key).Member1(member1).Member2(member2).Ft().Build()
	case EMPTY, KM:
		return b.Geodist().Key(key).Member1(member1).Member2(member2).Km().Build()
	default:
		panic(fmt.Sprintf("invalid unit %s", unit))
	}
}

func (b builder) GeoHashCompleted(key string, members ...string) Completed {
	return b.Geohash().Key(key).Member(members...).Build()
}

func (b builder) GeoPosCompleted(key string, members ...string) Completed {
	return b.Geopos().Key(key).Member(members...).Build()
}

func (b builder) GeoRadiusByMemberCompleted(key, member string, query GeoRadiusQuery) Completed {
	cmd := b.Arbitrary(XXX_GEORADIUSBYMEMBER_RO).Keys(key).Args(member)
	if query.Store != "" || query.StoreDist != "" {
		panic("GeoRadiusByMember does not support Store or StoreDist")
	}
	return cmd.Args(geoRadiusQueryArgs(query)...).Build()
}

func (b builder) GeoRadiusByMemberStoreCompleted(key, member string, query GeoRadiusQuery) Completed {
	cmd := b.Arbitrary(XXX_GEORADIUSBYMEMBER).Keys(key).Args(member)
	if query.Store == "" && query.StoreDist == "" {
		panic("GeoRadiusByMemberStore requires Store or StoreDist")
	}
	return cmd.Args(geoRadiusQueryArgs(query)...).Build()
}

func (b builder) GeoRadiusCompleted(key string, longitude, latitude float64, query GeoRadiusQuery) Completed {
	cmd := b.Arbitrary(XXX_GEORADIUS_RO).Keys(key).Args(strconv.FormatFloat(longitude, 'f', -1, 64), strconv.FormatFloat(latitude, 'f', -1, 64))
	if query.Store != "" || query.StoreDist != "" {
		panic("GeoRadius does not support Store or StoreDist")
	}
	return cmd.Args(geoRadiusQueryArgs(query)...).Build()
}

func (b builder) GeoRadiusStoreCompleted(key string, longitude, latitude float64, query GeoRadiusQuery) Completed {
	cmd := b.Arbitrary(XXX_GEORADIUS).Keys(key).Args(strconv.FormatFloat(longitude, 'f', -1, 64), strconv.FormatFloat(latitude, 'f', -1, 64))
	if query.Store == "" && query.StoreDist == "" {
		panic("GeoRadiusStore requires Store or StoreDist")
	}
	return cmd.Args(geoRadiusQueryArgs(query)...).Build()
}

func (b builder) GeoSearchCompleted(key string, q GeoSearchQuery) Completed {
	return b.Arbitrary(XXX_GEOSEARCH).Keys(key).Args(geoSearchQueryArgs(q)...).Build()
}

func (b builder) GeoSearchLocationCompleted(key string, q GeoSearchLocationQuery) Completed {
	return b.Arbitrary(XXX_GEOSEARCH).Keys(key).Args(geoSearchLocationQueryArgs(q)...).Build()
}

func (b builder) GeoSearchStoreCompleted(src, dest string, q GeoSearchStoreQuery) Completed {
	cmd := b.Arbitrary(XXX_GEOSEARCHSTORE).Keys(dest, src)
	cmd = cmd.Args(geoSearchQueryArgs(q.GeoSearchQuery)...)
	if q.StoreDist {
		cmd = cmd.Args(XXX_STOREDIST)
	}
	return cmd.Build()
}
