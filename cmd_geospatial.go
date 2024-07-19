package redisson

import (
	"context"
)

type GeospatialCmdable interface {
	GeospatialWriter
	GeospatialReader
}

type GeospatialWriter interface {
	// GeoAdd
	// Available since: 3.2.0
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @geo @slow
	GeoAdd(ctx context.Context, key string, geoLocation ...GeoLocation) IntCmd

	// GeoRadiusStore
	// Available since: 3.2.0
	// Time complexity: O(N+log(M)) where N is the number of elements inside the bounding box of the circular area delimited by center and radius and M is the number of items inside the index.
	// ACL categories: @write @geo @slow
	GeoRadiusStore(ctx context.Context, key string, longitude, latitude float64, query GeoRadiusQuery) IntCmd

	// GeoRadiusByMemberStore
	// Available since: 3.2.0
	// Time complexity: O(N+log(M)) where N is the number of elements inside the bounding box of the circular area delimited by center and radius and M is the number of items inside the index.
	// ACL categories: @write @geo @slow
	GeoRadiusByMemberStore(ctx context.Context, key, member string, query GeoRadiusQuery) IntCmd

	// GeoSearchStore
	// Available since: 6.2.0
	// Time complexity: O(N+log(M)) where N is the number of elements in the grid-aligned bounding box area around the shape provided as the filter and M is the number of items inside the shape
	// ACL categories: @write @geo @slow
	GeoSearchStore(ctx context.Context, key, store string, q GeoSearchStoreQuery) IntCmd
}

type GeospatialReader interface {
	// GeoSearchLocation
	// Available since: 6.2.0
	// Time complexity: O(N+log(M)) where N is the number of elements in the grid-aligned bounding box area around the shape provided as the filter and M is the number of items inside the shape
	// ACL categories: @read @geo @slow
	GeoSearchLocation(ctx context.Context, key string, q GeoSearchLocationQuery) GeoSearchLocationCmd
}

type GeospatialCacheCmdable interface {
	// GeoDist
	// Available since: 3.2.0
	// Time complexity: O(log(N))
	// ACL categories: @read @geo @slow
	GeoDist(ctx context.Context, key string, member1, member2, unit string) FloatCmd

	// GeoHash
	// Available since: 3.2.0
	// Time complexity: O(log(N)) for each member requested, where N is the number of elements in the sorted set.
	// ACL categories: @read @geo @slow
	GeoHash(ctx context.Context, key string, members ...string) StringSliceCmd

	// GeoPos
	// Available since: 3.2.0
	// Time complexity: O(N) where N is the number of members requested.
	// ACL categories: @read @geo @slow
	GeoPos(ctx context.Context, key string, members ...string) GeoPosCmd

	// GeoRadius
	// Available since: 3.2.10
	// Time complexity: O(N+log(M)) where N is the number of elements inside the bounding box of the circular area delimited by center and radius and M is the number of items inside the index.
	// ACL categories: @read @geo @slow
	GeoRadius(ctx context.Context, key string, longitude, latitude float64, query GeoRadiusQuery) GeoLocationCmd

	// GeoRadiusByMember
	// Available since: 3.2.10
	// Time complexity: O(N+log(M)) where N is the number of elements inside the bounding box of the circular area delimited by center and radius and M is the number of items inside the index.
	// ACL categories: @read @geo @slow
	GeoRadiusByMember(ctx context.Context, key, member string, query GeoRadiusQuery) GeoLocationCmd

	// GeoSearch
	// Available since: 6.2.0
	// Time complexity: O(N+log(M)) where N is the number of elements in the grid-aligned bounding box area around the shape provided as the filter and M is the number of items inside the shape
	// ACL categories: @read @geo @slow
	GeoSearch(ctx context.Context, key string, q GeoSearchQuery) StringSliceCmd
}

func (c *client) GeoAdd(ctx context.Context, key string, geoLocation ...GeoLocation) IntCmd {
	ctx = c.handler.before(ctx, CommandGeoAdd)
	r := c.adapter.GeoAdd(ctx, key, geoLocation...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GeoDist(ctx context.Context, key string, member1, member2, unit string) FloatCmd {
	ctx = c.handler.before(ctx, CommandGeoDist)
	var r FloatCmd
	if c.ttl > 0 {
		r = newFloatCmd(c.Do(ctx, c.builder.GeoDistCompleted(key, member1, member2, unit)))
	} else {
		r = c.adapter.GeoDist(ctx, key, member1, member2, unit)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GeoHash(ctx context.Context, key string, members ...string) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandGeoHash)
	var r StringSliceCmd
	if c.ttl > 0 {
		r = newStringSliceCmd(c.Do(ctx, c.builder.GeoHashCompleted(key, members...)))
	} else {
		r = c.adapter.GeoHash(ctx, key, members...)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GeoPos(ctx context.Context, key string, members ...string) GeoPosCmd {
	ctx = c.handler.before(ctx, CommandGeoPos)
	var r GeoPosCmd
	if c.ttl > 0 {
		r = newGeoPosCmd(c.Do(ctx, c.builder.GeoPosCompleted(key, members...)))
	} else {
		r = c.adapter.GeoPos(ctx, key, members...)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GeoRadius(ctx context.Context, key string, longitude, latitude float64, query GeoRadiusQuery) GeoLocationCmd {
	if query.Count > 0 {
		ctx = c.handler.before(ctx, CommandGeoRadiusROCount)
	} else {
		ctx = c.handler.before(ctx, CommandGeoRadiusRO)
	}
	var r GeoLocationCmd
	if c.ttl > 0 {
		r = newGeoLocationCmd(c.Do(ctx, c.builder.GeoRadiusCompleted(key, longitude, latitude, query)))
	} else {
		r = c.adapter.GeoRadius(ctx, key, longitude, latitude, query)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GeoRadiusStore(ctx context.Context, key string, longitude, latitude float64, query GeoRadiusQuery) IntCmd {
	f := func() []string {
		if len(query.Store) > 0 && len(query.StoreDist) > 0 {
			return appendString(key, query.Store, query.StoreDist)
		} else if len(query.Store) > 0 {
			return appendString(key, query.Store)
		} else if len(query.StoreDist) > 0 {
			return appendString(key, query.StoreDist)
		}
		return nil
	}
	if query.Count > 0 {
		ctx = c.handler.beforeWithKeys(ctx, CommandGeoRadiusStoreCount, f)
	} else {
		ctx = c.handler.beforeWithKeys(ctx, CommandGeoRadiusStore, f)
	}
	r := c.adapter.GeoRadiusStore(ctx, key, longitude, latitude, query)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GeoRadiusByMember(ctx context.Context, key, member string, query GeoRadiusQuery) GeoLocationCmd {
	ctx = c.handler.before(ctx, CommandGeoRadiusByMemberRO)
	var r GeoLocationCmd
	if c.ttl > 0 {
		r = newGeoLocationCmd(c.Do(ctx, c.builder.GeoRadiusByMemberCompleted(key, member, query)))
	} else {
		r = c.adapter.GeoRadiusByMember(ctx, key, member, query)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GeoRadiusByMemberStore(ctx context.Context, key, member string, query GeoRadiusQuery) IntCmd {
	ctx = c.handler.before(ctx, CommandGeoRadiusByMemberStore)
	r := c.adapter.GeoRadiusByMemberStore(ctx, key, member, query)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GeoSearch(ctx context.Context, key string, q GeoSearchQuery) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandGeoSearch)
	var r StringSliceCmd
	if c.ttl > 0 {
		r = newStringSliceCmd(c.Do(ctx, c.builder.GeoSearchCompleted(key, q)))
	} else {
		r = c.adapter.GeoSearch(ctx, key, q)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GeoSearchLocation(ctx context.Context, key string, q GeoSearchLocationQuery) GeoSearchLocationCmd {
	ctx = c.handler.before(ctx, CommandGeoSearch)
	r := c.adapter.GeoSearchLocation(ctx, key, q)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GeoSearchStore(ctx context.Context, key, store string, q GeoSearchStoreQuery) IntCmd {
	ctx = c.handler.before(ctx, CommandGeoSearchStore)
	r := c.adapter.GeoSearchStore(ctx, key, store, q)
	c.handler.after(ctx, r.Err())
	return r
}
