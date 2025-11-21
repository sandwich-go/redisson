package redisson

import (
	goredis "github.com/redis/go-redis/v9"
)

const KeepTTL = goredis.KeepTTL

//------------------------------------------------------------------------------

type SetArgs = goredis.SetArgs

//------------------------------------------------------------------------------

type Sort = goredis.Sort

//------------------------------------------------------------------------------

type LPosArgs = goredis.LPosArgs

//------------------------------------------------------------------------------

type (
	ZStore     = goredis.ZStore
	ZAddArgs   = goredis.ZAddArgs
	ZRangeBy   = goredis.ZRangeBy
	ZRangeArgs = goredis.ZRangeArgs
)

//------------------------------------------------------------------------------

type (
	XAddArgs        = goredis.XAddArgs
	XAutoClaimArgs  = goredis.XAutoClaimArgs
	XClaimArgs      = goredis.XClaimArgs
	XPendingExtArgs = goredis.XPendingExtArgs
	XReadArgs       = goredis.XReadArgs
	XReadGroupArgs  = goredis.XReadGroupArgs
)

//------------------------------------------------------------------------------

type BitCount = goredis.BitCount

//------------------------------------------------------------------------------

type (
	GeoSearchQuery         = goredis.GeoSearchQuery
	GeoSearchLocationQuery = goredis.GeoSearchLocationQuery
	GeoSearchStoreQuery    = goredis.GeoSearchStoreQuery
	GeoRadiusQuery         = goredis.GeoRadiusQuery
)

func getGeoSearchQueryArgs(q GeoSearchQuery) []string {
	args := make([]string, 0, 11)
	return _getGeoSearchQueryArgs(q, args)
}

func _getGeoSearchQueryArgs(q GeoSearchQuery, args []string) []string {
	if len(q.Member) > 0 {
		args = append(args, FROMMEMBER, q.Member)
	} else {
		args = append(args, FROMLONLAT, str(q.Longitude), str(q.Latitude))
	}
	if q.Radius > 0 {
		if len(q.RadiusUnit) == 0 {
			q.RadiusUnit = KM
		}
		args = append(args, BYRADIUS, str(q.Radius), q.RadiusUnit)
	} else {
		if len(q.BoxUnit) == 0 {
			q.BoxUnit = KM
		}
		args = append(args, BYBOX, str(q.BoxWidth), str(q.BoxHeight), q.BoxUnit)
	}
	if len(q.Sort) > 0 {
		args = append(args, q.Sort)
	}
	if q.Count > 0 {
		args = append(args, COUNT, str(int64(q.Count)))
		if q.CountAny {
			args = append(args, ANY)
		}
	}
	return args
}

func getGeoSearchLocationQueryArgs(q GeoSearchLocationQuery) []string {
	args := make([]string, 0, 14)
	args = _getGeoSearchQueryArgs(q.GeoSearchQuery, args)
	if q.WithCoord {
		args = append(args, WITHCOORD)
	}
	if q.WithDist {
		args = append(args, WITHDIST)
	}
	if q.WithHash {
		args = append(args, WITHHASH)
	}
	return args
}

func getGeoRadiusQueryArgs(q GeoRadiusQuery) []string {
	args := make([]string, 0, 12)
	args = append(args, str(q.Radius))
	if len(q.Unit) > 0 {
		args = append(args, q.Unit)
	} else {
		args = append(args, KM)
	}
	if q.WithCoord {
		args = append(args, WITHCOORD)
	}
	if q.WithDist {
		args = append(args, WITHDIST)
	}
	if q.WithGeoHash {
		args = append(args, WITHHASH)
	}
	if q.Count > 0 {
		args = append(args, COUNT, str(int64(q.Count)))
	}
	if len(q.Sort) > 0 {
		args = append(args, q.Sort)
	}
	if len(q.Store) > 0 {
		args = append(args, STORE, q.Store)
	}
	if len(q.StoreDist) > 0 {
		args = append(args, STOREDIST, q.StoreDist)
	}
	return args
}
