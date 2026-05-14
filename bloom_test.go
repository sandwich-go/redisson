//go:build integration

package redisson

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBloomFilter_Basic(t *testing.T) {
	c := MustNewClient(NewConf(WithDevelopment(false)))
	t.Cleanup(func() { _ = c.Close() })

	bf, err := c.NewBloomFilter("rss-test-bloom", 1000, 0.01)
	if err != nil {
		t.Fatalf("NewBloomFilter: %v", err)
	}

	ctx := context.Background()
	t.Cleanup(func() { _ = bf.Delete(ctx) })

	Convey("Bloom Add/Exists", t, func() {
		So(bf.Add(ctx, "foo"), ShouldBeNil)
		So(bf.Add(ctx, "bar"), ShouldBeNil)

		ok, err := bf.Exists(ctx, "foo")
		So(err, ShouldBeNil)
		So(ok, ShouldBeTrue)

		ok, err = bf.Exists(ctx, "never-added-zxcvbn")
		So(err, ShouldBeNil)
		// 误判率 0.01，对单个未加项几乎应为 false
		So(ok, ShouldBeFalse)
	})

	Convey("Bloom AddMulti/ExistsMulti", t, func() {
		So(bf.AddMulti(ctx, []string{"a", "b", "c"}), ShouldBeNil)
		got, err := bf.ExistsMulti(ctx, []string{"a", "b", "c", "neverNeverNever"})
		So(err, ShouldBeNil)
		So(len(got), ShouldEqual, 4)
		So(got[0], ShouldBeTrue)
		So(got[1], ShouldBeTrue)
		So(got[2], ShouldBeTrue)
	})

	Convey("Bloom Reset", t, func() {
		So(bf.Add(ctx, "to-be-reset"), ShouldBeNil)
		So(bf.Reset(ctx), ShouldBeNil)
		ok, err := bf.Exists(ctx, "to-be-reset")
		So(err, ShouldBeNil)
		So(ok, ShouldBeFalse)
	})
}
