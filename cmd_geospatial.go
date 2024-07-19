package redisson

import (
	"context"
)

type GeospatialCmdable interface {
	GeospatialWriter
}

type GeospatialWriter interface {
	// GeoAdd
	// Available since: 3.2.0
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @geo @slow
	// Options
	// 	- XX: Only update elements that already exist. Never add elements.
	// 	- NX: Don't update already existing elements. Always add new elements.
	// 	- CH: Modify the return value from the number of new elements added, to the total number of elements changed (CH is an abbreviation of changed).
	//		Changed elements are new elements added and elements already existing for which the coordinates was updated. So elements specified in the
	//		command line having the same score as they had in the past are not counted. Note: normally, the return value of GEOADD only counts the number of new elements added.
	// Note: The XX and NX options are mutually exclusive.
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: When used without optional arguments, the number of elements added to the sorted set (excluding score updates).
	//		If the CH option is specified, the number of elements that were changed (added or updated).
	// History:
	//	- Starting with Redis version 6.2.0: Added the CH, NX and XX options.
	GeoAdd(ctx context.Context, key string, geoLocation ...GeoLocation) IntCmd

	// GeoRadiusStore
	// Available since: 3.2.0
	// Time complexity: O(N+log(M)) where N is the number of elements inside the bounding box of the circular area delimited by center and radius and M is the number of items inside the index.
	// ACL categories: @write @geo @slow
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- If no WITH* option is specified, an Array reply of matched member names
	//		- If WITHCOORD, WITHDIST, or WITHHASH options are specified, the command returns an Array reply of arrays, where each sub-array represents a single item:
	//			1. The distance from the center as a floating point number, in the same unit specified in the radius.
	//			2. The Geohash integer.
	//			3. The coordinates as a two items x,y array (longitude,latitude).
	// History:
	//	- Starting with Redis version 6.2.0: Added the ANY option for COUNT.
	//	- Starting with Redis version 7.0.0: Added support for uppercase unit names.
	GeoRadiusStore(ctx context.Context, key string, longitude, latitude float64, query GeoRadiusQuery) IntCmd

	// GeoRadiusByMemberStore
	// Available since: 3.2.0
	// Time complexity: O(N+log(M)) where N is the number of elements inside the bounding box of the circular area delimited by center and radius and M is the number of items inside the index.
	// ACL categories: @write @geo @slow
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- If no WITH* option is specified, an Array reply of matched member names
	//		- If WITHCOORD, WITHDIST, or WITHHASH options are specified, the command returns an Array reply of arrays, where each sub-array represents a single item:
	//			1. The distance from the center as a floating point number, in the same unit specified in the radius.
	//			2. The Geohash integer.
	//			3. The coordinates as a two items x,y array (longitude,latitude).
	// History:
	//	- Starting with Redis version 7.0.0: Added support for uppercase unit names.
	GeoRadiusByMemberStore(ctx context.Context, key, member string, query GeoRadiusQuery) IntCmd

	// GeoSearchStore
	// Available since: 6.2.0
	// Time complexity: O(N+log(M)) where N is the number of elements in the grid-aligned bounding box area around the shape provided as the filter and M is the number of items inside the shape
	// ACL categories: @write @geo @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of elements in the resulting set
	// History:
	//	- Starting with Redis version 7.0.0: Added support for uppercase unit names.
	GeoSearchStore(ctx context.Context, key, store string, q GeoSearchStoreQuery) IntCmd
}

type GeospatialCacheCmdable interface {
	// GeoDist
	// Available since: 3.2.0
	// Time complexity: O(1)
	// ACL categories: @read @geo @slow
	// Options
	// 	- M for meters.
	// 	- KM for kilometers.
	// 	- MI for miles.
	// 	- FT for feet.
	// RESP2 Reply:
	//	One of the following:
	//		- Nil reply: one or both of the elements are missing.
	//		- Bulk string reply: distance as a double (represented as a string) in the specified units.
	// RESP3 Reply:
	//	One of the following:
	//		- Null reply: one or both of the elements are missing.
	//		- Bulk string reply: distance as a double (represented as a string) in the specified units.
	GeoDist(ctx context.Context, key string, member1, member2, unit string) FloatCmd

	// GeoHash
	// Available since: 3.2.0
	// Time complexity: O(1) for each member requested.
	// ACL categories: @read @geo @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: an array where each element is the Geohash corresponding to each member name passed as an argument to the command.
	GeoHash(ctx context.Context, key string, members ...string) StringSliceCmd

	// GeoPos
	// Available since: 3.2.0
	// Time complexity: O(1) for each member requested.
	// ACL categories: @read @geo @slow
	// RESP2 Reply:
	// 	- Array reply: An array where each element is a two elements array representing longitude and latitude (x,y) of
	//		each member name passed as argument to the command. Non-existing elements are reported as Nil reply elements of the array.
	// RESP3 Reply:
	// 	- Array reply: An array where each element is a two elements array representing longitude and latitude (x,y) of
	//		each member name passed as argument to the command. Non-existing elements are reported as Null reply elements of the array.
	GeoPos(ctx context.Context, key string, members ...string) GeoPosCmd

	// GeoRadius
	// Available since: 3.2.10
	// Time complexity: O(N+log(M)) where N is the number of elements inside the bounding box of the circular area delimited by center and radius and M is the number of items inside the index.
	// ACL categories: @read @geo @slow
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- If no WITH* option is specified, an Array reply of matched member names
	//		- If WITHCOORD, WITHDIST, or WITHHASH options are specified, the command returns an Array reply of arrays, where each sub-array represents a single item:
	//			1. The distance from the center as a floating point number, in the same unit specified in the radius.
	//			2. The Geohash integer.
	//			3. The coordinates as a two items x,y array (longitude,latitude).
	// History:
	//	- Starting with Redis version 6.2.0: Added the ANY option for COUNT.
	GeoRadius(ctx context.Context, key string, longitude, latitude float64, query GeoRadiusQuery) GeoLocationCmd

	// GeoRadiusByMember
	// Available since: 3.2.10
	// Time complexity: O(N+log(M)) where N is the number of elements inside the bounding box of the circular area delimited by center and radius and M is the number of items inside the index.
	// ACL categories: @read @geo @slow
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- If no WITH* option is specified, an Array reply of matched member names
	//		- If WITHCOORD, WITHDIST, or WITHHASH options are specified, the command returns an Array reply of arrays, where each sub-array represents a single item:
	//			1. The distance from the center as a floating point number, in the same unit specified in the radius.
	//			2. The Geohash integer.
	//			3. The coordinates as a two items x,y array (longitude,latitude).
	GeoRadiusByMember(ctx context.Context, key, member string, query GeoRadiusQuery) GeoLocationCmd

	// GeoSearch
	// Available since: 6.2.0
	// Time complexity: O(N+log(M)) where N is the number of elements in the grid-aligned bounding box area around the shape provided as the filter and M is the number of items inside the shape
	// ACL categories: @read @geo @slow
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- If no WITH* option is specified, an Array reply of matched member names
	//		- If WITHCOORD, WITHDIST, or WITHHASH options are specified, the command returns an Array reply of arrays, where each sub-array represents a single item:
	//			1. The distance from the center as a floating point number, in the same unit specified in the radius.
	//			2. The Geohash integer.
	//			3. The coordinates as a two items x,y array (longitude,latitude).
	// History:
	//	- Starting with Redis version 7.0.0: Added support for uppercase unit names.
	GeoSearch(ctx context.Context, key string, q GeoSearchQuery) StringSliceCmd
	GeoSearchLocation(ctx context.Context, key string, q GeoSearchLocationQuery) GeoSearchLocationCmd
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
	ctx = c.handler.before(ctx, CommandGeoRadiusRO)
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
	ctx = c.handler.beforeWithKeys(ctx, CommandGeoRadiusStore, f)
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
