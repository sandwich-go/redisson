package redisson

import "context"

type GeospatialCmdable interface {
	GeospatialWriter
	GeospatialReader
}

type GeospatialWriter interface {
	// GeoAdd
	// Available since: 3.2.0
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @geo @slow
	// Adds the specified geospatial items (longitude, latitude, name) to the specified key. Data is stored into the key as a sorted set, in a way that makes it possible to query the items with the GEOSEARCH command.
	// The command takes arguments in the standard format x,y so the longitude must be specified before the latitude. There are limits to the coordinates that can be indexed: areas very near to the poles are not indexable.
	// The exact limits, as specified by EPSG:900913 / EPSG:3785 / OSGEO:41001 are the following:
	// 	Valid longitudes are from -180 to 180 degrees.
	// 	Valid latitudes are from -85.05112878 to 85.05112878 degrees.
	// The command will report an error when the user attempts to index coordinates outside the specified ranges.
	// Note: there is no GEODEL command because you can use ZREM to remove elements. The Geo index structure is just a sorted set.
	// GEOADD options
	// GEOADD also provides the following options:
	//	XX: Only update elements that already exist. Never add elements.
	//	NX: Don't update already existing elements. Always add new elements.
	//	CH: Modify the return value from the number of new elements added, to the total number of elements changed (CH is an abbreviation of changed). Changed elements are new elements added and elements already existing for which the coordinates was updated. So elements specified in the command line having the same score as they had in the past are not counted. Note: normally, the return value of GEOADD only counts the number of new elements added.
	// Note: The XX and NX options are mutually exclusive.
	// Return:
	//	Integer reply, specifically:
	//	When used without optional arguments, the number of elements added to the sorted set (excluding score updates).
	//	If the CH option is specified, the number of elements that were changed (added or updated).
	// Starting with Redis version 6.2.0: Added the CH, NX and XX options.
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
	// As of Redis version 6.2.0, this command is regarded as deprecated.
	// It can be replaced by GEOSEARCH and GEOSEARCHSTORE with the BYRADIUS and FROMMEMBER arguments when migrating or writing new code.
	// This command is exactly like GEORADIUS with the sole difference that instead of taking, as the center of the area to query, a longitude and latitude value, it takes the name of a member already existing inside the geospatial index represented by the sorted set.
	// The position of the specified member is used as the center of the query.
	// Please check the example below and the GEORADIUS documentation for more information about the command and its options.
	// Note that GEORADIUSBYMEMBER_RO is also available since Redis 3.2.10 and Redis 4.0.0 in order to provide a read-only command that can be used in replicas. See the GEORADIUS page for more information.
	GeoRadiusByMemberStore(ctx context.Context, key, member string, query GeoRadiusQuery) IntCmd

	// GeoSearchStore
	// Available since: 6.2.0
	// Time complexity: O(N+log(M)) where N is the number of elements in the grid-aligned bounding box area around the shape provided as the filter and M is the number of items inside the shape
	// ACL categories: @write @geo @slow
	// This command is like GEOSEARCH, but stores the result in destination key.
	// This command comes in place of the now deprecated GEORADIUS and GEORADIUSBYMEMBER.
	// By default, it stores the results in the destination sorted set with their geospatial information.
	// When using the STOREDIST option, the command stores the items in a sorted set populated with their distance from the center of the circle or box, as a floating-point number, in the same unit specified for that shape.
	// Return:
	//	Integer reply: the number of elements in the resulting set.
	GeoSearchStore(ctx context.Context, key, store string, q GeoSearchStoreQuery) IntCmd
}

type GeospatialReader interface{}

type GeospatialCacheCmdable interface {
	// GeoDist
	// Available since: 3.2.0
	// Time complexity: O(log(N))
	// ACL categories: @read @geo @slow
	// Return the distance between two members in the geospatial index represented by the sorted set.
	// Given a sorted set representing a geospatial index, populated using the GEOADD command, the command returns the distance between the two specified members in the specified unit.
	// If one or both the members are missing, the command returns NULL.
	// The unit must be one of the following, and defaults to meters:
	// 	m for meters.
	//	km for kilometers.
	//	mi for miles.
	//	ft for feet.
	// The distance is computed assuming that the Earth is a perfect sphere, so errors up to 0.5% are possible in edge cases.
	// Return:
	// Bulk string reply, specifically:
	// The command returns the distance as a double (represented as a string) in the specified unit, or NULL if one or both the elements are missing.
	GeoDist(ctx context.Context, key string, member1, member2, unit string) FloatCmd

	// GeoHash
	// Available since: 3.2.0
	// Time complexity: O(log(N)) for each member requested, where N is the number of elements in the sorted set.
	// ACL categories: @read @geo @slow
	// Return valid Geohash strings representing the position of one or more elements in a sorted set value representing a geospatial index (where elements were added using GEOADD).
	// Normally Redis represents positions of elements using a variation of the Geohash technique where positions are encoded using 52 bit integers. The encoding is also different compared to the standard because the initial min and max coordinates used during the encoding and decoding process are different. This command however returns a standard Geohash in the form of a string as described in the Wikipedia article and compatible with the geohash.org web site.
	// Geohash string properties
	// The command returns 11 characters Geohash strings, so no precision is lost compared to the Redis internal 52 bit representation. The returned Geohashes have the following properties:
	//	They can be shortened removing characters from the right. It will lose precision but will still point to the same area.
	//	It is possible to use them in geohash.org URLs such as http://geohash.org/<geohash-string>. This is an example of such URL.
	//	Strings with a similar prefix are nearby, but the contrary is not true, it is possible that strings with different prefixes are nearby too.
	// Return:
	// Array reply, specifically:
	// The command returns an array where each element is the Geohash corresponding to each member name passed as argument to the command.
	GeoHash(ctx context.Context, key string, members ...string) StringSliceCmd

	// GeoPos
	// Available since: 3.2.0
	// Time complexity: O(N) where N is the number of members requested.
	// ACL categories: @read @geo @slow
	// Return the positions (longitude,latitude) of all the specified members of the geospatial index represented by the sorted set at key.
	// Given a sorted set representing a geospatial index, populated using the GEOADD command, it is often useful to obtain back the coordinates of specified members. When the geospatial index is populated via GEOADD the coordinates are converted into a 52 bit geohash, so the coordinates returned may not be exactly the ones used in order to add the elements, but small errors may be introduced.
	// The command can accept a variable number of arguments so it always returns an array of positions even when a single element is specified.
	// Return:
	//Array reply, specifically:
	// The command returns an array where each element is a two elements array representing longitude and latitude (x,y) of each member name passed as argument to the command.
	// Non existing elements are reported as NULL elements of the array.
	GeoPos(ctx context.Context, key string, members ...string) GeoPosCmd

	// GeoRadius
	// Available since: 3.2.10
	// Time complexity: O(N+log(M)) where N is the number of elements inside the bounding box of the circular area delimited by center and radius and M is the number of items inside the index.
	// ACL categories: @read @geo @slow
	// As of Redis version 6.2.0, this command is regarded as deprecated.
	// It can be replaced by GEOSEARCH with the BYRADIUS argument when migrating or writing new code.
	// Read-only variant of the GEORADIUS command.
	// This command is identical to the GEORADIUS command, except that it doesn't support the optional STORE and STOREDIST parameters.
	// Return:
	//	Array reply: An array with each entry being the corresponding result of the subcommand given at the same position.
	GeoRadius(ctx context.Context, key string, longitude, latitude float64, query GeoRadiusQuery) GeoLocationCmd

	// GeoRadiusByMember
	// Available since: 3.2.10
	// Time complexity: O(N+log(M)) where N is the number of elements inside the bounding box of the circular area delimited by center and radius and M is the number of items inside the index.
	// ACL categories: @read @geo @slow
	// As of Redis version 6.2.0, this command is regarded as deprecated.
	// It can be replaced by GEOSEARCH with the BYRADIUS and FROMMEMBER arguments when migrating or writing new code.
	// Read-only variant of the GEORADIUSBYMEMBER command.
	// This command is identical to the GEORADIUSBYMEMBER command, except that it doesn't support the optional STORE and STOREDIST parameters.
	GeoRadiusByMember(ctx context.Context, key, member string, query GeoRadiusQuery) GeoLocationCmd

	// GeoSearch
	// Available since: 6.2.0
	// Time complexity: O(N+log(M)) where N is the number of elements in the grid-aligned bounding box area around the shape provided as the filter and M is the number of items inside the shape
	// ACL categories: @read @geo @slow
	// Return the members of a sorted set populated with geospatial information using GEOADD, which are within the borders of the area specified by a given shape. This command extends the GEORADIUS command, so in addition to searching within circular areas, it supports searching within rectangular areas.
	// This command should be used in place of the deprecated GEORADIUS and GEORADIUSBYMEMBER commands.
	// The query's center point is provided by one of these mandatory options:
	//	FROMMEMBER: Use the position of the given existing <member> in the sorted set.
	//	FROMLONLAT: Use the given <longitude> and <latitude> position.
	// The query's shape is provided by one of these mandatory options:
	//	BYRADIUS: Similar to GEORADIUS, search inside circular area according to given <radius>.
	//	BYBOX: Search inside an axis-aligned rectangle, determined by <height> and <width>.
	// The command optionally returns additional information using the following options:
	//	WITHDIST: Also return the distance of the returned items from the specified center point. The distance is returned in the same unit as specified for the radius or height and width arguments.
	//	WITHCOORD: Also return the longitude and latitude of the matching items.
	//	WITHHASH: Also return the raw geohash-encoded sorted set score of the item, in the form of a 52 bit unsigned integer. This is only useful for low level hacks or debugging and is otherwise of little interest for the general user.
	// Matching items are returned unsorted by default. To sort them, use one of the following two options:
	//	ASC: Sort returned items from the nearest to the farthest, relative to the center point.
	//	DESC: Sort returned items from the farthest to the nearest, relative to the center point.
	// All matching items are returned by default. To limit the results to the first N matching items, use the COUNT <count> option. When the ANY option is used, the command returns as soon as enough matches are found. This means that the results returned may not be the ones closest to the specified point, but the effort invested by the server to generate them is significantly less. When ANY is not provided, the command will perform an effort that is proportional to the number of items matching the specified area and sort them, so to query very large areas with a very small COUNT option may be slow even if just a few results are returned.
	// Return:
	// Array reply, specifically:
	//	Without any WITH option specified, the command just returns a linear array like ["New York","Milan","Paris"].
	//	If WITHCOORD, WITHDIST or WITHHASH options are specified, the command returns an array of arrays, where each sub-array represents a single item.
	// When additional information is returned as an array of arrays for each item, the first item in the sub-array is always the name of the returned item. The other information is returned in the following order as successive elements of the sub-array.
	//	The distance from the center as a floating point number, in the same unit specified in the shape.
	//	The geohash integer.
	//	The coordinates as a two items x,y array (longitude,latitude).
	GeoSearch(ctx context.Context, key string, q GeoSearchQuery) StringSliceCmd

	// GeoSearchLocation
	// Available since: 6.2.0
	// Time complexity: O(N+log(M)) where N is the number of elements in the grid-aligned bounding box area around the shape provided as the filter and M is the number of items inside the shape
	// ACL categories: @read @geo @slow
	// See GeoSearch
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
		r = c.adapter.Cache(c.ttl).GeoDist(ctx, key, member1, member2, unit)
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
		r = c.adapter.Cache(c.ttl).GeoHash(ctx, key, members...)
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
		r = c.adapter.Cache(c.ttl).GeoPos(ctx, key, members...)
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
		r = c.adapter.Cache(c.ttl).GeoRadius(ctx, key, longitude, latitude, query)
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
		r = c.adapter.Cache(c.ttl).GeoRadiusByMember(ctx, key, member, query)
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
		r = c.adapter.Cache(c.ttl).GeoSearch(ctx, key, q)
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
