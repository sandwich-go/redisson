package redisson

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func beforeStream(ctx context.Context, key string, c Cmdable) {
	id := c.XAdd(ctx, XAddArgs{
		Stream: key,
		ID:     "1-0",
		Values: map[string]interface{}{"uno": "un"},
	})
	So(id.Err(), ShouldBeNil)
	So(id.Val(), ShouldEqual, "1-0")

	// Values supports []interface{}.
	id = c.XAdd(ctx, XAddArgs{
		Stream: key,
		ID:     "2-0",
		Values: []interface{}{"dos", "deux"},
	})
	So(id.Err(), ShouldBeNil)
	So(id.Val(), ShouldEqual, "2-0")

	// Value supports []string.
	id = c.XAdd(ctx, XAddArgs{
		Stream: key,
		ID:     "3-0",
		Values: []string{"tres", "troix"},
	})
	So(id.Err(), ShouldBeNil)
	So(id.Val(), ShouldEqual, "3-0")
}

func testXAdd(ctx context.Context, c Cmdable) []string {
	var key = "stream"
	id := c.XAdd(ctx, XAddArgs{
		Stream: key,
		ID:     "1-0",
		Values: map[string]interface{}{"uno": "un"},
	})
	So(id.Err(), ShouldBeNil)
	So(id.Val(), ShouldEqual, "1-0")

	xlen := c.XLen(ctx, key)
	So(xlen.Err(), ShouldBeNil)
	So(xlen.Val(), ShouldEqual, 1)

	return []string{key}
}

func testXLen(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	xlen := c.XLen(ctx, key)
	So(xlen.Err(), ShouldBeNil)
	So(xlen.Val(), ShouldEqual, 3)

	return []string{key}
}

func testXAck(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	err := c.XGroupCreateMkStream(ctx, key, "group", "0").Err()
	So(err, ShouldBeNil)

	err = c.XReadGroup(ctx, XReadGroupArgs{
		Group:    "group",
		Consumer: "consumer",
		Streams:  []string{key, ">"},
		Count:    3,
	}).Err()
	So(err, ShouldBeNil)

	n := c.XAck(ctx, key, "group", "1-0", "2-0", "4-0")
	So(n.Err(), ShouldBeNil)
	So(n.Val(), ShouldEqual, 2)

	return []string{key}
}

func testXDel(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	xlen := c.XLen(ctx, key)
	So(xlen.Err(), ShouldBeNil)
	So(xlen.Val(), ShouldEqual, 3)

	xdel := c.XDel(ctx, key, "1-0", "2-0", "3-0", "4-0")
	So(xdel.Err(), ShouldBeNil)
	So(xdel.Val(), ShouldEqual, 3)

	xlen = c.XLen(ctx, key)
	So(xlen.Err(), ShouldBeNil)
	So(xlen.Val(), ShouldEqual, 0)

	return []string{key}
}

func xMessagesEqual(a, b []XMessage) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil && b != nil {
		return false
	}
	if b == nil && a != nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if !xMessageEqual(b[k], v) {
			return false
		}
	}
	return true
}

func testXRange(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	xlen := c.XLen(ctx, key)
	So(xlen.Err(), ShouldBeNil)
	So(xlen.Val(), ShouldEqual, 3)

	xrange := c.XRange(ctx, key, "-", "+")
	So(xrange.Err(), ShouldBeNil)
	So(len(xrange.Val()), ShouldEqual, 3)
	So(xMessagesEqual(xrange.Val(), []XMessage{
		{ID: "1-0", Values: map[string]interface{}{"uno": "un"}},
		{ID: "2-0", Values: map[string]interface{}{"dos": "deux"}},
		{ID: "3-0", Values: map[string]interface{}{"tres": "troix"}},
	}), ShouldBeTrue)

	xrange = c.XRange(ctx, key, "1-0", "2-0")
	So(xrange.Err(), ShouldBeNil)
	So(len(xrange.Val()), ShouldEqual, 2)
	So(xMessagesEqual(xrange.Val(), []XMessage{
		{ID: "1-0", Values: map[string]interface{}{"uno": "un"}},
		{ID: "2-0", Values: map[string]interface{}{"dos": "deux"}},
	}), ShouldBeTrue)

	return []string{key}
}

func testXRangeN(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	xlen := c.XLen(ctx, key)
	So(xlen.Err(), ShouldBeNil)
	So(xlen.Val(), ShouldEqual, 3)

	xrange := c.XRangeN(ctx, key, "-", "+", 2)
	So(xrange.Err(), ShouldBeNil)
	So(len(xrange.Val()), ShouldEqual, 2)
	So(xMessagesEqual(xrange.Val(), []XMessage{
		{ID: "1-0", Values: map[string]interface{}{"uno": "un"}},
		{ID: "2-0", Values: map[string]interface{}{"dos": "deux"}},
	}), ShouldBeTrue)

	xrange = c.XRangeN(ctx, key, "1-0", "2-0", 1)
	So(xrange.Err(), ShouldBeNil)
	So(len(xrange.Val()), ShouldEqual, 1)
	So(xMessagesEqual(xrange.Val(), []XMessage{
		{ID: "1-0", Values: map[string]interface{}{"uno": "un"}},
	}), ShouldBeTrue)

	return []string{key}
}

func testXRevRange(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	xlen := c.XLen(ctx, key)
	So(xlen.Err(), ShouldBeNil)
	So(xlen.Val(), ShouldEqual, 3)

	xrange := c.XRevRange(ctx, key, "+", "-")
	So(xrange.Err(), ShouldBeNil)
	So(len(xrange.Val()), ShouldEqual, 3)
	So(xMessagesEqual(xrange.Val(), []XMessage{
		{ID: "3-0", Values: map[string]interface{}{"tres": "troix"}},
		{ID: "2-0", Values: map[string]interface{}{"dos": "deux"}},
		{ID: "1-0", Values: map[string]interface{}{"uno": "un"}},
	}), ShouldBeTrue)

	xrange = c.XRevRange(ctx, key, "2-0", "1-0")
	So(xrange.Err(), ShouldBeNil)
	So(len(xrange.Val()), ShouldEqual, 2)
	So(xMessagesEqual(xrange.Val(), []XMessage{
		{ID: "2-0", Values: map[string]interface{}{"dos": "deux"}},
		{ID: "1-0", Values: map[string]interface{}{"uno": "un"}},
	}), ShouldBeTrue)

	return []string{key}
}

func testXRevRangeN(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	xlen := c.XLen(ctx, key)
	So(xlen.Err(), ShouldBeNil)
	So(xlen.Val(), ShouldEqual, 3)

	xrange := c.XRevRangeN(ctx, key, "+", "-", 2)
	So(xrange.Err(), ShouldBeNil)
	So(len(xrange.Val()), ShouldEqual, 2)
	So(xMessagesEqual(xrange.Val(), []XMessage{
		{ID: "3-0", Values: map[string]interface{}{"tres": "troix"}},
		{ID: "2-0", Values: map[string]interface{}{"dos": "deux"}},
	}), ShouldBeTrue)

	xrange = c.XRevRangeN(ctx, key, "2-0", "1-0", 1)
	So(xrange.Err(), ShouldBeNil)
	So(len(xrange.Val()), ShouldEqual, 1)
	So(xMessagesEqual(xrange.Val(), []XMessage{
		{ID: "2-0", Values: map[string]interface{}{"dos": "deux"}},
	}), ShouldBeTrue)

	return []string{key}
}

func xMessageEqual(a, b XMessage) bool {
	if a.ID != b.ID {
		return false
	}
	if len(a.Values) != len(b.Values) {
		return false
	}
	for k, v := range a.Values {
		v1, ok := b.Values[k]
		if !ok || v != v1 {
			return false
		}
	}
	return true
}

func testXAutoClaim(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	err := c.XGroupCreateMkStream(ctx, key, "group", "0").Err()
	So(err, ShouldBeNil)

	err = c.XReadGroup(ctx, XReadGroupArgs{
		Group:    "group",
		Consumer: "consumer",
		Streams:  []string{key, ">"},
		Count:    3,
	}).Err()
	So(err, ShouldBeNil)

	xca := XAutoClaimArgs{
		Stream:   key,
		Group:    "group",
		Consumer: "consumer",
		Start:    "-",
		Count:    2,
	}
	msgs, start, err := c.XAutoClaim(ctx, xca).Result()
	So(err, ShouldBeNil)
	So(start, ShouldEqual, "3-0")
	So(xMessagesEqual(msgs, []XMessage{{
		ID:     "1-0",
		Values: map[string]interface{}{"uno": "un"},
	}, {
		ID:     "2-0",
		Values: map[string]interface{}{"dos": "deux"},
	}}), ShouldBeTrue)

	xca.Start = start
	msgs, start, err = c.XAutoClaim(ctx, xca).Result()
	So(err, ShouldBeNil)
	So(start, ShouldEqual, "0-0")
	So(xMessagesEqual(msgs, []XMessage{{
		ID:     "3-0",
		Values: map[string]interface{}{"tres": "troix"},
	}}), ShouldBeTrue)

	ids, start1, err1 := c.XAutoClaimJustID(ctx, xca).Result()
	So(err1, ShouldBeNil)
	So(start1, ShouldEqual, "0-0")
	So(stringSliceEqual(ids, []string{"3-0"}, true), ShouldBeTrue)

	return []string{key}
}

func testXClaim(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	err := c.XGroupCreateMkStream(ctx, key, "group", "0").Err()
	So(err, ShouldBeNil)

	err = c.XReadGroup(ctx, XReadGroupArgs{
		Group:    "group",
		Consumer: "consumer",
		Streams:  []string{key, ">"},
		Count:    3,
	}).Err()
	So(err, ShouldBeNil)

	msgs, err1 := c.XClaim(ctx, XClaimArgs{
		Stream:   key,
		Group:    "group",
		Consumer: "consumer",
		Messages: []string{"1-0", "2-0", "3-0"},
	}).Result()
	So(err1, ShouldBeNil)
	So(xMessagesEqual(msgs, []XMessage{{
		ID:     "1-0",
		Values: map[string]interface{}{"uno": "un"},
	}, {
		ID:     "2-0",
		Values: map[string]interface{}{"dos": "deux"},
	}, {
		ID:     "3-0",
		Values: map[string]interface{}{"tres": "troix"},
	}}), ShouldBeTrue)

	ids, err2 := c.XClaimJustID(ctx, XClaimArgs{
		Stream:   key,
		Group:    "group",
		Consumer: "consumer",
		Messages: []string{"1-0", "2-0", "3-0"},
	}).Result()
	So(err2, ShouldBeNil)
	So(stringSliceEqual(ids, []string{"1-0", "2-0", "3-0"}, true), ShouldBeTrue)

	return []string{key}
}

func testXGroupCreate(ctx context.Context, c Cmdable) []string {
	var key = "stream2"

	err := c.XGroupCreate(ctx, key, "group", "0").Err()
	So(err, ShouldNotBeNil)

	xadd := c.XAdd(ctx, XAddArgs{
		Stream:     key,
		NoMkStream: false,
		ID:         "1-0",
		Values:     map[string]interface{}{"uno": "un"},
	})
	So(xadd.Err(), ShouldBeNil)
	So(xadd.Val(), ShouldEqual, "1-0")

	err = c.XGroupCreate(ctx, key, "group", "0").Err()
	So(err, ShouldBeNil)

	return []string{key}
}

func testXGroupCreateMkStream(ctx context.Context, c Cmdable) []string {
	var key = "stream2"

	err := c.XGroupCreateMkStream(ctx, key, "group", "0").Err()
	So(err, ShouldBeNil)

	err = c.XGroupCreateMkStream(ctx, key, "group", "0").Err()
	So(err, ShouldNotBeNil)
	So(err.Error(), ShouldEqual, "BUSYGROUP Consumer Group name already exists")

	n, err1 := c.XGroupDestroy(ctx, key, "group").Result()
	So(err1, ShouldBeNil)
	So(n, ShouldEqual, 1)

	n, err = c.Del(ctx, key).Result()
	So(err, ShouldBeNil)
	So(n, ShouldEqual, 1)

	return []string{key}
}

func testXGroupCreateConsumer(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	err := c.XGroupCreate(ctx, key, "group", "0").Err()
	So(err, ShouldBeNil)

	n, err := c.XGroupCreateConsumer(ctx, key, "group", "c1").Result()
	So(err, ShouldBeNil)
	So(n, ShouldEqual, 1)

	return []string{key}
}

func testXGroupDelConsumer(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	err := c.XGroupCreate(ctx, key, "group", "0").Err()
	So(err, ShouldBeNil)

	n, err := c.XGroupCreateConsumer(ctx, key, "group", "c1").Result()
	So(err, ShouldBeNil)
	So(n, ShouldEqual, 1)

	n, err = c.XGroupDelConsumer(ctx, key, "group", "c1").Result()
	So(err, ShouldBeNil)
	So(n, ShouldEqual, 0)

	return []string{key}
}

func testXGroupDestroy(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	err := c.XGroupCreate(ctx, key, "group", "0").Err()
	So(err, ShouldBeNil)

	n, err1 := c.XGroupDestroy(ctx, key, "group").Result()
	So(err1, ShouldBeNil)
	So(n, ShouldEqual, 1)

	n, err = c.XGroupDestroy(ctx, key, "group2").Result()
	So(err, ShouldBeNil)
	So(n, ShouldEqual, 0)

	return []string{key}
}

func testXGroupSetID(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	err := c.XGroupCreate(ctx, key, "group", "0").Err()
	So(err, ShouldBeNil)

	xGroupSetID := c.XGroupSetID(ctx, key, "group", "2-0")
	So(xGroupSetID.Err(), ShouldBeNil)
	So(xGroupSetID.Val(), ShouldEqual, OK)

	xReadGroup := c.XReadGroup(ctx, XReadGroupArgs{
		Group:    "group",
		Consumer: "consumer",
		Streams:  []string{key, ">"},
		Count:    3,
	})
	So(xReadGroup.Err(), ShouldBeNil)
	So(len(xReadGroup.Val()), ShouldEqual, 1)
	So(xMessagesEqual(xReadGroup.Val()[0].Messages, []XMessage{{
		ID:     "3-0",
		Values: map[string]interface{}{"tres": "troix"},
	}}), ShouldBeTrue)

	return []string{key}
}

func testXReadGroup(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	err := c.XGroupCreate(ctx, key, "group", "0").Err()
	So(err, ShouldBeNil)

	xReadGroup := c.XReadGroup(ctx, XReadGroupArgs{
		Group:    "group",
		Consumer: "consumer",
		Streams:  []string{key, ">"},
		Count:    3,
	})
	So(xReadGroup.Err(), ShouldBeNil)
	So(len(xReadGroup.Val()), ShouldEqual, 1)
	So(xMessagesEqual(xReadGroup.Val()[0].Messages, []XMessage{{
		ID:     "1-0",
		Values: map[string]interface{}{"uno": "un"},
	}, {
		ID:     "2-0",
		Values: map[string]interface{}{"dos": "deux"},
	}, {
		ID:     "3-0",
		Values: map[string]interface{}{"tres": "troix"},
	}}), ShouldBeTrue)

	return []string{key}
}

func testXRead(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	xread := c.XRead(ctx, XReadArgs{
		Streams: []string{key, "0"},
		Count:   3,
	})

	So(xread.Err(), ShouldBeNil)
	So(len(xread.Val()), ShouldEqual, 1)
	So(xMessagesEqual(xread.Val()[0].Messages, []XMessage{{
		ID:     "1-0",
		Values: map[string]interface{}{"uno": "un"},
	}, {
		ID:     "2-0",
		Values: map[string]interface{}{"dos": "deux"},
	}, {
		ID:     "3-0",
		Values: map[string]interface{}{"tres": "troix"},
	}}), ShouldBeTrue)

	return []string{key}
}

func testXReadStreams(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	xread := c.XReadStreams(ctx, key, "0")

	So(xread.Err(), ShouldBeNil)
	So(len(xread.Val()), ShouldEqual, 1)
	So(xMessagesEqual(xread.Val()[0].Messages, []XMessage{{
		ID:     "1-0",
		Values: map[string]interface{}{"uno": "un"},
	}, {
		ID:     "2-0",
		Values: map[string]interface{}{"dos": "deux"},
	}, {
		ID:     "3-0",
		Values: map[string]interface{}{"tres": "troix"},
	}}), ShouldBeTrue)

	return []string{key}
}

func testXPending(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	err := c.XGroupCreate(ctx, key, "group", "0").Err()
	So(err, ShouldBeNil)

	xReadGroup := c.XReadGroup(ctx, XReadGroupArgs{
		Group:    "group",
		Consumer: "consumer",
		Streams:  []string{key, ">"},
		Count:    3,
	})
	So(xReadGroup.Err(), ShouldBeNil)

	info, err1 := c.XPending(ctx, key, "group").Result()
	So(err1, ShouldBeNil)
	So(info.Count, ShouldEqual, 3)
	So(info.Lower, ShouldEqual, "1-0")
	So(info.Higher, ShouldEqual, "3-0")
	So(len(info.Consumers), ShouldEqual, 1)
	So(info.Consumers["consumer"], ShouldEqual, 3)

	args := XPendingExtArgs{
		Stream:   key,
		Group:    "group",
		Start:    "-",
		End:      "+",
		Count:    10,
		Consumer: "consumer",
	}
	infoExt, err1 := c.XPendingExt(ctx, args).Result()
	So(err1, ShouldBeNil)
	So(len(infoExt), ShouldEqual, 3)
	var infoExtMapping = make(map[string]XPendingExt)
	for i := range infoExt {
		infoExt[i].Idle = 0
		infoExtMapping[infoExt[i].ID] = infoExt[i]
	}
	So(infoExtMapping["1-0"].Consumer, ShouldEqual, "consumer")
	So(infoExtMapping["1-0"].RetryCount, ShouldEqual, 1)
	So(infoExtMapping["2-0"].Consumer, ShouldEqual, "consumer")
	So(infoExtMapping["2-0"].RetryCount, ShouldEqual, 1)
	So(infoExtMapping["3-0"].Consumer, ShouldEqual, "consumer")
	So(infoExtMapping["3-0"].RetryCount, ShouldEqual, 1)

	args.Idle = 72 * time.Hour
	infoExt, err = c.XPendingExt(ctx, args).Result()
	So(err, ShouldBeNil)
	So(len(infoExt), ShouldEqual, 0)

	return []string{key}
}

func testXTrim(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	n, err := c.XTrim(ctx, key, 0).Result()
	So(err, ShouldBeNil)
	So(n, ShouldEqual, 3)

	return []string{key}
}

func testXTrimApprox(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	n, err := c.XTrimApprox(ctx, key, 0).Result()
	So(err, ShouldBeNil)
	So(n, ShouldEqual, 3)

	return []string{key}
}

func testXTrimMaxLen(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	n, err := c.XTrimMaxLen(ctx, key, 0).Result()
	So(err, ShouldBeNil)
	So(n, ShouldEqual, 3)

	return []string{key}
}

func testXTrimMaxLenApprox(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	n, err := c.XTrimMaxLenApprox(ctx, key, 0, 0).Result()
	So(err, ShouldBeNil)
	So(n, ShouldEqual, 3)

	return []string{key}
}

func testXTrimMinID(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	n, err := c.XTrimMinID(ctx, key, "4-0").Result()
	So(err, ShouldBeNil)
	So(n, ShouldEqual, 3)

	return []string{key}
}

func testXTrimMinIDApprox(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	n, err := c.XTrimMinIDApprox(ctx, key, "4-0", 0).Result()
	So(err, ShouldBeNil)
	So(n, ShouldEqual, 3)

	return []string{key}
}

func testXInfoConsumers(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	err := c.XGroupCreate(ctx, key, "group", "0").Err()
	So(err, ShouldBeNil)

	n, err1 := c.XGroupCreateConsumer(ctx, key, "group", "c1").Result()
	So(err1, ShouldBeNil)
	So(n, ShouldEqual, 1)

	res, err2 := c.XInfoConsumers(ctx, key, "group").Result()
	So(err2, ShouldBeNil)
	for i := range res {
		res[i].Idle = 0
	}
	So(len(res), ShouldEqual, 1)

	return []string{key}
}

func testXInfoGroups(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	err := c.XGroupCreate(ctx, key, "group", "0").Err()
	So(err, ShouldBeNil)

	res, err2 := c.XInfoGroups(ctx, key).Result()
	So(err2, ShouldBeNil)
	So(len(res), ShouldEqual, 1)

	return []string{key}
}

func testXInfoStream(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	err := c.XInfoStream(ctx, key).Err()
	So(err, ShouldBeNil)

	return []string{key}
}

func testXInfoStreamFull(ctx context.Context, c Cmdable) []string {
	var key = "stream"

	beforeStream(ctx, key, c)

	err := c.XInfoStreamFull(ctx, key, 1).Err()
	So(err, ShouldBeNil)

	return []string{key}
}

func streamTestUnits() []TestUnit {
	return []TestUnit{
		{CommandXAdd, testXAdd},
		{CommandXAck, testXAck},
		{CommandXLen, testXLen},
		{CommandXDel, testXDel},
		{CommandXRange, testXRange},
		{CommandXRange, testXRangeN},
		{CommandXRevRange, testXRevRange},
		{CommandXRevRange, testXRevRangeN},
		{CommandXAutoClaim, testXAutoClaim},
		{CommandXClaim, testXClaim},
		{CommandXGroupCreate, testXGroupCreate},
		{CommandXGroupCreate, testXGroupCreateMkStream},
		{CommandXGroupCreateConsumer, testXGroupCreateConsumer},
		{CommandXGroupDelConsumer, testXGroupDelConsumer},
		{CommandXGroupDestroy, testXGroupDestroy},
		{CommandXGroupSetID, testXGroupSetID},
		{CommandXReadGroup, testXReadGroup},
		{CommandXRead, testXRead},
		{CommandXRead, testXReadStreams},
		{CommandXPending, testXPending},
		{CommandXTrim, testXTrim},
		{CommandXTrim, testXTrimApprox},
		{CommandXTrim, testXTrimMaxLen},
		{CommandXTrim, testXTrimMaxLenApprox},
		{CommandXTrim, testXTrimMinID},
		{CommandXTrim, testXTrimMinIDApprox},
		{CommandXInfoConsumers, testXInfoConsumers},
		{CommandXInfoGroups, testXInfoGroups},
		{CommandXInfoStream, testXInfoStream},
		{CommandXInfoStreamFull, testXInfoStreamFull},
	}
}

func TestResp2Client_Stream(t *testing.T) { doTestUnits(t, RESP2, streamTestUnits) }
