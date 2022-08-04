package redisson

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func testGeoAdd(ctx context.Context, c Cmdable) []string {
	var key, member1, member2 = "Sicily", "Palermo", "Catania"
	geoAdd := c.GeoAdd(ctx, key,
		GeoLocation{Name: member1, Longitude: 13.361389, Latitude: 38.115556},
		GeoLocation{Name: member2, Longitude: 15.087269, Latitude: 37.502669},
	)
	So(geoAdd.Err(), ShouldBeNil)
	So(geoAdd.Val(), ShouldEqual, 2)

	geoAdd = c.GeoAdd(ctx, key,
		GeoLocation{Name: member1, Longitude: 13.361389, Latitude: 38.115556},
	)
	So(geoAdd.Err(), ShouldBeNil)
	So(geoAdd.Val(), ShouldEqual, 0)

	return []string{key}
}

func testGeoDist(ctx context.Context, c Cmdable) []string {
	var key, member1, member2 = "Sicily", "Palermo", "Catania"
	geoAdd := c.GeoAdd(ctx, key,
		GeoLocation{Name: member1, Longitude: 13.361389, Latitude: 38.115556},
		GeoLocation{Name: member2, Longitude: 15.087269, Latitude: 37.502669},
	)
	So(geoAdd.Err(), ShouldBeNil)
	So(geoAdd.Val(), ShouldEqual, 2)

	geoDist := c.GeoDist(ctx, key, member1, member2, "m")
	So(geoDist.Err(), ShouldBeNil)
	So(geoDist.Val(), ShouldEqual, 166274.1516)

	geoDist = c.GeoDist(ctx, key, member1, member2, "")
	So(geoDist.Err(), ShouldBeNil)
	So(geoDist.Val(), ShouldEqual, 166.2742)

	geoDist = c.GeoDist(ctx, key, member1, member2, "km")
	So(geoDist.Err(), ShouldBeNil)
	So(geoDist.Val(), ShouldEqual, 166.2742)

	geoDist = c.GeoDist(ctx, key, member1, member2, "mi")
	So(geoDist.Err(), ShouldBeNil)
	So(geoDist.Val(), ShouldEqual, 103.3182)

	geoDist = c.GeoDist(ctx, key, "Foo", "Bar", "")
	So(geoDist.Err(), ShouldNotBeNil)
	So(IsNil(geoDist.Err()), ShouldBeTrue)

	return []string{key}
}

func testGeoHash(ctx context.Context, c Cmdable) []string {
	var key, member1, member2 = "Sicily", "Palermo", "Catania"
	geoAdd := c.GeoAdd(ctx, key,
		GeoLocation{Name: member1, Longitude: 13.361389, Latitude: 38.115556},
		GeoLocation{Name: member2, Longitude: 15.087269, Latitude: 37.502669},
	)
	So(geoAdd.Err(), ShouldBeNil)
	So(geoAdd.Val(), ShouldEqual, 2)

	geoHash := c.GeoHash(ctx, key, member1, member2)
	So(geoHash.Err(), ShouldBeNil)
	So(stringSliceEqual(geoHash.Val(), []string{"sqc8b49rny0", "sqdtr74hyu0"}, true), ShouldBeTrue)

	return []string{key}
}

func testGeoPos(ctx context.Context, c Cmdable) []string {
	var key, member1, member2 = "Sicily", "Palermo", "Catania"
	geoAdd := c.GeoAdd(ctx, key,
		GeoLocation{Name: member1, Longitude: 13.361389, Latitude: 38.115556},
		GeoLocation{Name: member2, Longitude: 15.087269, Latitude: 37.502669},
	)
	So(geoAdd.Err(), ShouldBeNil)
	So(geoAdd.Val(), ShouldEqual, 2)

	geoPos := c.GeoPos(ctx, key, member1, member2, "NonExisting")
	So(geoPos.Err(), ShouldBeNil)
	So(len(geoPos.Val()), ShouldEqual, 3)
	So(geoPos.Val()[0].Longitude, ShouldEqual, 13.36138933897018433)
	So(geoPos.Val()[0].Latitude, ShouldEqual, 38.11555639549629859)
	So(geoPos.Val()[1].Longitude, ShouldEqual, 15.08726745843887329)
	So(geoPos.Val()[1].Latitude, ShouldEqual, 37.50266842333162032)
	So(geoPos.Val()[2], ShouldBeNil)

	return []string{key}
}

func testGeoRadiusRO(ctx context.Context, c Cmdable) []string {
	var key, member1, member2 = "Sicily", "Palermo", "Catania"
	geoAdd := c.GeoAdd(ctx, key,
		GeoLocation{Name: member1, Longitude: 13.361389, Latitude: 38.115556},
		GeoLocation{Name: member2, Longitude: 15.087269, Latitude: 37.502669},
	)
	So(geoAdd.Err(), ShouldBeNil)
	So(geoAdd.Val(), ShouldEqual, 2)

	geoRadius := c.GeoRadius(ctx, key, 15, 37, GeoRadiusQuery{
		Radius:   200,
		Unit:     "km",
		WithDist: true,
	})
	So(geoRadius.Err(), ShouldBeNil)
	So(len(geoRadius.Val()), ShouldEqual, 2)
	So(geoRadius.Val()[0].Name, ShouldEqual, member1)
	So(geoRadius.Val()[0].Dist, ShouldEqual, 190.4424)
	So(geoRadius.Val()[1].Name, ShouldEqual, member2)
	So(geoRadius.Val()[1].Dist, ShouldEqual, 56.4413)

	geoRadius = c.GeoRadius(ctx, key, 15, 37, GeoRadiusQuery{
		Radius:    200,
		Unit:      "km",
		WithCoord: true,
	})
	So(geoRadius.Err(), ShouldBeNil)
	So(len(geoRadius.Val()), ShouldEqual, 2)
	So(geoRadius.Val()[0].Name, ShouldEqual, member1)
	So(geoRadius.Val()[0].Longitude, ShouldEqual, 13.36138933897018433)
	So(geoRadius.Val()[0].Latitude, ShouldEqual, 38.11555639549629859)
	So(geoRadius.Val()[1].Name, ShouldEqual, member2)
	So(geoRadius.Val()[1].Longitude, ShouldEqual, 15.08726745843887329)
	So(geoRadius.Val()[1].Latitude, ShouldEqual, 37.50266842333162032)

	geoRadius = c.GeoRadius(ctx, key, 15, 37, GeoRadiusQuery{
		Radius:    200,
		Unit:      "km",
		WithCoord: true,
		WithDist:  true,
	})
	So(geoRadius.Err(), ShouldBeNil)
	So(len(geoRadius.Val()), ShouldEqual, 2)
	So(geoRadius.Val()[0].Name, ShouldEqual, member1)
	So(geoRadius.Val()[0].Dist, ShouldEqual, 190.4424)
	So(geoRadius.Val()[0].Longitude, ShouldEqual, 13.36138933897018433)
	So(geoRadius.Val()[0].Latitude, ShouldEqual, 38.11555639549629859)
	So(geoRadius.Val()[1].Name, ShouldEqual, member2)
	So(geoRadius.Val()[1].Longitude, ShouldEqual, 15.08726745843887329)
	So(geoRadius.Val()[1].Latitude, ShouldEqual, 37.50266842333162032)
	So(geoRadius.Val()[1].Dist, ShouldEqual, 56.4413)

	return []string{key}
}

func testGeoRadiusStore(ctx context.Context, c Cmdable) []string {
	var key, key1, member1, member2 = "Sicily", "Sicily1", "Palermo", "Catania"
	geoAdd := c.GeoAdd(ctx, key,
		GeoLocation{Name: member1, Longitude: 13.361389, Latitude: 38.115556},
		GeoLocation{Name: member2, Longitude: 15.087269, Latitude: 37.502669},
	)
	So(geoAdd.Err(), ShouldBeNil)
	So(geoAdd.Val(), ShouldEqual, 2)

	geoRadius := c.GeoRadiusStore(ctx, key, 15, 37, GeoRadiusQuery{
		Radius:    200,
		Unit:      "km",
		StoreDist: key1,
	})
	So(geoRadius.Err(), ShouldBeNil)
	So(geoRadius.Val(), ShouldEqual, 2)

	return []string{key, key1}
}

func testGeoRadiusByMemberRO(ctx context.Context, c Cmdable) []string {
	var key, member1, member2, member3 = "Sicily", "Palermo", "Catania", "Agrigento"
	geoAdd := c.GeoAdd(ctx, key,
		GeoLocation{Name: member3, Longitude: 13.583333, Latitude: 37.316667},
	)
	So(geoAdd.Err(), ShouldBeNil)
	So(geoAdd.Val(), ShouldEqual, 1)

	geoAdd = c.GeoAdd(ctx, key,
		GeoLocation{Name: member1, Longitude: 13.361389, Latitude: 38.115556},
		GeoLocation{Name: member2, Longitude: 15.087269, Latitude: 37.502669},
	)
	So(geoAdd.Err(), ShouldBeNil)
	So(geoAdd.Val(), ShouldEqual, 2)

	g := c.GeoRadiusByMember(ctx, key, member3, GeoRadiusQuery{
		Radius: 100,
		Unit:   "km",
	})
	So(g.Err(), ShouldBeNil)
	So(len(g.Val()), ShouldEqual, 2)
	So(g.Val()[0].Name, ShouldEqual, member3)
	So(g.Val()[1].Name, ShouldEqual, member1)

	return []string{key}
}

func testGeoRadiusByMember(ctx context.Context, c Cmdable) []string {
	var key, key1, member1, member2, member3 = "Sicily", "Sicily1", "Palermo", "Catania", "Agrigento"
	geoAdd := c.GeoAdd(ctx, key,
		GeoLocation{Name: member3, Longitude: 13.583333, Latitude: 37.316667},
		GeoLocation{Name: member1, Longitude: 13.361389, Latitude: 38.115556},
		GeoLocation{Name: member2, Longitude: 15.087269, Latitude: 37.502669},
	)
	So(geoAdd.Err(), ShouldBeNil)
	So(geoAdd.Val(), ShouldEqual, 3)

	g := c.GeoRadiusByMemberStore(ctx, key, member3, GeoRadiusQuery{
		Radius:    100,
		Unit:      "km",
		StoreDist: key1,
	})
	So(g.Err(), ShouldBeNil)
	So(g.Val(), ShouldEqual, 2)

	return []string{key, key1}
}

func testGeoSearch(ctx context.Context, c Cmdable) []string {
	var key, member1, member2, member3, member4 = "Sicily", "Palermo", "Catania", "edge1", "edge2"
	geoAdd := c.GeoAdd(ctx, key,
		GeoLocation{Name: member1, Longitude: 13.361389, Latitude: 38.115556},
		GeoLocation{Name: member2, Longitude: 15.087269, Latitude: 37.502669},
	)
	So(geoAdd.Err(), ShouldBeNil)
	So(geoAdd.Val(), ShouldEqual, 2)

	geoAdd = c.GeoAdd(ctx, key,
		GeoLocation{Name: member3, Longitude: 12.758489, Latitude: 38.788135},
		GeoLocation{Name: member4, Longitude: 17.241510, Latitude: 38.788135},
	)
	So(geoAdd.Err(), ShouldBeNil)
	So(geoAdd.Val(), ShouldEqual, 2)

	geoSearch := c.GeoSearch(ctx, key, GeoSearchQuery{Longitude: 15, Latitude: 37, Radius: 200, RadiusUnit: "km", Sort: "asc"})
	So(geoSearch.Err(), ShouldBeNil)
	So(len(geoSearch.Val()), ShouldEqual, 2)
	So(geoSearch.Val()[0], ShouldEqual, member2)
	So(geoSearch.Val()[1], ShouldEqual, member1)

	return []string{key}
}

func testGeoSearchStore(ctx context.Context, c Cmdable) []string {
	var key, key1, member1, member2, member3, member4 = "Sicily", "Sicily1", "Palermo", "Catania", "edge1", "edge2"
	geoAdd := c.GeoAdd(ctx, key,
		GeoLocation{Name: member1, Longitude: 13.361389, Latitude: 38.115556},
		GeoLocation{Name: member2, Longitude: 15.087269, Latitude: 37.502669},
		GeoLocation{Name: member3, Longitude: 12.758489, Latitude: 38.788135},
		GeoLocation{Name: member4, Longitude: 17.241510, Latitude: 38.788135},
	)
	So(geoAdd.Err(), ShouldBeNil)
	So(geoAdd.Val(), ShouldEqual, 4)

	geoSearch := c.GeoSearchStore(ctx, key, key1, GeoSearchStoreQuery{GeoSearchQuery: GeoSearchQuery{Longitude: 15, Latitude: 37, Radius: 200, RadiusUnit: "km", Sort: "asc"}})
	So(geoSearch.Err(), ShouldBeNil)
	So(geoSearch.Val(), ShouldEqual, 2)

	return []string{key, key1}
}

func testGeoSearchLocation(ctx context.Context, c Cmdable) []string {
	var key, member1, member2, member3, member4 = "Sicily", "Palermo", "Catania", "edge1", "edge2"
	geoAdd := c.GeoAdd(ctx, key,
		GeoLocation{Name: member1, Longitude: 13.361389, Latitude: 38.115556},
		GeoLocation{Name: member2, Longitude: 15.087269, Latitude: 37.502669},
	)
	So(geoAdd.Err(), ShouldBeNil)
	So(geoAdd.Val(), ShouldEqual, 2)

	geoAdd = c.GeoAdd(ctx, key,
		GeoLocation{Name: member3, Longitude: 12.758489, Latitude: 38.788135},
		GeoLocation{Name: member4, Longitude: 17.241510, Latitude: 38.788135},
	)
	So(geoAdd.Err(), ShouldBeNil)
	So(geoAdd.Val(), ShouldEqual, 2)

	geoSearch := c.GeoSearchLocation(ctx, key, GeoSearchLocationQuery{
		GeoSearchQuery: GeoSearchQuery{Longitude: 15, Latitude: 37, BoxHeight: 400, BoxWidth: 400, BoxUnit: "km", Sort: "asc"},
		WithCoord:      true,
		WithDist:       true,
	},
	)
	So(geoSearch.Err(), ShouldBeNil)
	So(len(geoSearch.Val()), ShouldEqual, 4)
	So(geoSearch.Val()[0].Name, ShouldEqual, member2)
	So(geoSearch.Val()[0].Dist, ShouldEqual, 56.4413)
	So(geoSearch.Val()[0].Longitude, ShouldEqual, 15.08726745843887329)
	So(geoSearch.Val()[0].Latitude, ShouldEqual, 37.50266842333162032)

	So(geoSearch.Val()[1].Name, ShouldEqual, member1)
	So(geoSearch.Val()[1].Dist, ShouldEqual, 190.4424)
	So(geoSearch.Val()[1].Longitude, ShouldEqual, 13.36138933897018433)
	So(geoSearch.Val()[1].Latitude, ShouldEqual, 38.11555639549629859)

	So(geoSearch.Val()[2].Name, ShouldEqual, member4)
	So(geoSearch.Val()[2].Dist, ShouldEqual, 279.7403)
	So(geoSearch.Val()[2].Longitude, ShouldEqual, 17.24151045083999634)
	So(geoSearch.Val()[2].Latitude, ShouldEqual, 38.78813451624225195)

	So(geoSearch.Val()[3].Name, ShouldEqual, member3)
	So(geoSearch.Val()[3].Dist, ShouldEqual, 279.7405)
	So(geoSearch.Val()[3].Longitude, ShouldEqual, 12.7584877610206604)
	So(geoSearch.Val()[3].Latitude, ShouldEqual, 38.78813451624225195)

	return []string{key}
}

func geospatialTestUnits() []TestUnit {
	return []TestUnit{
		{CommandGeoAdd, testGeoAdd},
		{CommandGeoDist, testGeoDist},
		{CommandGeoHash, testGeoHash},
		{CommandGeoPos, testGeoPos},
		{CommandGeoRadiusRO, testGeoRadiusRO},
		{CommandGeoRadiusStore, testGeoRadiusStore},
		{CommandGeoRadiusByMemberRO, testGeoRadiusByMemberRO},
		{CommandGeoRadiusByMemberStore, testGeoRadiusByMember},
		{CommandGeoSearch, testGeoSearch},
		{CommandGeoSearchStore, testGeoSearchStore},
		{CommandGeoSearch, testGeoSearchLocation},
	}
}

func TestResp2Client_Geospatial(t *testing.T) { doTestUnits(t, RESP2, geospatialTestUnits) }
func TestResp3Client_Geospatial(t *testing.T) { doTestUnits(t, RESP3, geospatialTestUnits) }
