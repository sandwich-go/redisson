package redisson

import (
	"strconv"
)

func (b builder) XAckCompleted(stream, group string, ids ...string) Completed {
	return b.Xack().Key(stream).Group(group).Id(ids...).Build()
}

func (b builder) XAddCompleted(a XAddArgs) Completed {
	cmd := b.Arbitrary(KwXAdd).Keys(a.Stream)
	if a.NoMkStream {
		cmd = cmd.Args(KwNoMkStream)
	}
	switch {
	case a.MaxLen > 0:
		if a.Approx {
			cmd = cmd.Args(KwMaxLen, "~", strconv.FormatInt(a.MaxLen, 10))
		} else {
			cmd = cmd.Args(KwMaxLen, strconv.FormatInt(a.MaxLen, 10))
		}
	case a.MinID != "":
		if a.Approx {
			cmd = cmd.Args(KwMinID, "~", a.MinID)
		} else {
			cmd = cmd.Args(KwMinID, a.MinID)
		}
	}
	if a.Limit > 0 {
		cmd = cmd.Args(KwLimit, strconv.FormatInt(a.Limit, 10))
	}
	if a.ID != "" {
		cmd = cmd.Args(a.ID)
	} else {
		cmd = cmd.Args("*")
	}
	cmd = cmd.Args(argToSlice(a.Values)...)
	return cmd.Build()
}

func (b builder) XAutoClaimCompleted(a XAutoClaimArgs) Completed {
	if a.Count > 0 {
		return b.Xautoclaim().Key(a.Stream).Group(a.Group).Consumer(a.Consumer).MinIdleTime(strconv.FormatInt(formatMs(a.MinIdle), 10)).Start(a.Start).Count(a.Count).Build()
	} else {
		return b.Xautoclaim().Key(a.Stream).Group(a.Group).Consumer(a.Consumer).MinIdleTime(strconv.FormatInt(formatMs(a.MinIdle), 10)).Start(a.Start).Build()
	}
}

func (b builder) XAutoClaimJustIDCompleted(a XAutoClaimArgs) Completed {
	if a.Count > 0 {
		return b.Xautoclaim().Key(a.Stream).Group(a.Group).Consumer(a.Consumer).MinIdleTime(strconv.FormatInt(formatMs(a.MinIdle), 10)).Start(a.Start).Count(a.Count).Justid().Build()
	} else {
		return b.Xautoclaim().Key(a.Stream).Group(a.Group).Consumer(a.Consumer).MinIdleTime(strconv.FormatInt(formatMs(a.MinIdle), 10)).Start(a.Start).Justid().Build()
	}
}

func (b builder) XClaimCompleted(a XClaimArgs) Completed {
	return b.Xclaim().Key(a.Stream).Group(a.Group).Consumer(a.Consumer).MinIdleTime(strconv.FormatInt(formatMs(a.MinIdle), 10)).Id(a.Messages...).Build()
}

func (b builder) XClaimJustIDCompleted(a XClaimArgs) Completed {
	return b.Xclaim().Key(a.Stream).Group(a.Group).Consumer(a.Consumer).MinIdleTime(strconv.FormatInt(formatMs(a.MinIdle), 10)).Id(a.Messages...).Justid().Build()
}

func (b builder) XDelCompleted(stream string, ids ...string) Completed {
	return b.Xdel().Key(stream).Id(ids...).Build()
}

func (b builder) XGroupCreateCompleted(stream, group, start string) Completed {
	return b.XgroupCreate().Key(stream).Group(group).Id(start).Build()
}

func (b builder) XGroupCreateMkStreamCompleted(stream, group, start string) Completed {
	return b.XgroupCreate().Key(stream).Group(group).Id(start).Mkstream().Build()
}

func (b builder) XGroupCreateConsumerCompleted(stream, group, consumer string) Completed {
	return b.XgroupCreateconsumer().Key(stream).Group(group).Consumer(consumer).Build()
}

func (b builder) XGroupDelConsumerCompleted(stream, group, consumer string) Completed {
	return b.XgroupDelconsumer().Key(stream).Group(group).Consumername(consumer).Build()
}

func (b builder) XGroupDestroyCompleted(stream, group string) Completed {
	return b.XgroupDestroy().Key(stream).Group(group).Build()
}

func (b builder) XGroupSetIDCompleted(stream, group, start string) Completed {
	return b.XgroupSetid().Key(stream).Group(group).Id(start).Build()
}

func (b builder) XInfoConsumersCompleted(key, group string) Completed {
	return b.XinfoConsumers().Key(key).Group(group).Build()
}

func (b builder) XInfoGroupsCompleted(key string) Completed {
	return b.XinfoGroups().Key(key).Build()
}

func (b builder) XInfoStreamCompleted(key string) Completed {
	return b.XinfoStream().Key(key).Build()
}

func (b builder) XInfoStreamFullCompleted(key string, count int64) Completed {
	return b.XinfoStream().Key(key).Full().Count(count).Build()
}

func (b builder) XLenCompleted(stream string) Completed {
	return b.Xlen().Key(stream).Build()
}

func (b builder) XPendingCompleted(stream, group string) Completed {
	return b.Xpending().Key(stream).Group(group).Build()
}

func (b builder) XPendingExtCompleted(a XPendingExtArgs) Completed {
	cmd := b.Arbitrary(KwXPending).Keys(a.Stream).Args(a.Group)
	if a.Idle != 0 {
		cmd = cmd.Args(KwIdle, strconv.FormatInt(formatMs(a.Idle), 10))
	}
	cmd = cmd.Args(a.Start, a.End, strconv.FormatInt(a.Count, 10))
	if a.Consumer != "" {
		cmd = cmd.Args(a.Consumer)
	}
	return cmd.Build()
}

func (b builder) XRangeCompleted(stream, start, stop string) Completed {
	return b.Xrange().Key(stream).Start(start).End(stop).Build()
}

func (b builder) XRangeNCompleted(stream, start, stop string, count int64) Completed {
	return b.Xrange().Key(stream).Start(start).End(stop).Count(count).Build()
}

func (b builder) XRevRangeCompleted(stream, stop, start string) Completed {
	return b.Xrevrange().Key(stream).End(stop).Start(start).Build()
}

func (b builder) XRevRangeNCompleted(stream, stop, start string, count int64) Completed {
	return b.Xrevrange().Key(stream).End(stop).Start(start).Count(count).Build()
}

func (b builder) xTrim(key, strategy string,
	approx bool, threshold string, limit int64) Completed {
	cmd := b.Arbitrary(KwXTrim).Keys(key).Args(strategy)
	if approx {
		cmd = cmd.Args("~")
	}
	cmd = cmd.Args(threshold)
	if limit > 0 {
		cmd = cmd.Args(KwLimit, strconv.FormatInt(limit, 10))
	}
	return cmd.Build()
}

func (b builder) XTrimCompleted(key string, maxLen int64) Completed {
	return b.xTrim(key, KwMaxLen, false, strconv.FormatInt(maxLen, 10), 0)
}

func (b builder) XTrimMaxLenApproxCompleted(key string, maxLen, limit int64) Completed {
	return b.xTrim(key, KwMaxLen, true, strconv.FormatInt(maxLen, 10), limit)
}

func (b builder) XTrimMinIDCompleted(key string, minID string) Completed {
	return b.xTrim(key, KwMinID, false, minID, 0)
}

func (b builder) XTrimMinIDApproxCompleted(key string, minID string, limit int64) Completed {
	return b.xTrim(key, KwMinID, true, minID, limit)
}
