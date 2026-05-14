package redisson

import (
	"strconv"
)

func (b builder) zAddArgs(key string, incr bool, args ZAddArgs) Completed {
	cmd := b.Arbitrary(KwZAdd).Keys(key)
	// The GT, LT and NX options are mutually exclusive.
	if args.NX {
		cmd = cmd.Args(NX)
	} else {
		if args.XX {
			cmd = cmd.Args(XX)
		}
		if args.GT {
			cmd = cmd.Args(KwGt)
		} else if args.LT {
			cmd = cmd.Args(KwLt)
		}
	}
	if args.Ch {
		cmd = cmd.Args(KwCh)
	}
	if incr {
		cmd = cmd.Args(KwIncr)
	}
	for _, v := range args.Members {
		cmd = cmd.Args(strconv.FormatFloat(v.Score, 'f', -1, 64), v.Member)
	}
	return cmd.Build()
}

func (b builder) ZAddCompleted(key string, members ...Z) Completed {
	return b.zAddArgs(key, false, ZAddArgs{Members: members})
}

func (b builder) ZAddNXCompleted(key string, members ...Z) Completed {
	return b.zAddArgs(key, false, ZAddArgs{Members: members, NX: true})
}

func (b builder) ZAddXXCompleted(key string, members ...Z) Completed {
	return b.zAddArgs(key, false, ZAddArgs{Members: members, XX: true})
}

func (b builder) ZAddLTCompleted(key string, members ...Z) Completed {
	return b.zAddArgs(key, false, ZAddArgs{Members: members, LT: true})
}

func (b builder) ZAddGTCompleted(key string, members ...Z) Completed {
	return b.zAddArgs(key, false, ZAddArgs{Members: members, GT: true})
}

func (b builder) ZAddChCompleted(key string, members ...Z) Completed {
	return b.zAddArgs(key, false, ZAddArgs{Members: members, Ch: true})
}

func (b builder) ZAddArgsCompleted(key string, args ZAddArgs) Completed {
	return b.zAddArgs(key, false, args)
}

func (b builder) ZAddArgsIncrCompleted(key string, args ZAddArgs) Completed {
	return b.zAddArgs(key, true, args)
}

func (b builder) ZCardCompleted(key string) Completed {
	return b.Zcard().Key(key).Build()
}

func (b builder) ZCountCompleted(key, min, max string) Completed {
	return b.Zcount().Key(key).Min(min).Max(max).Build()
}

func (b builder) ZLexCountCompleted(key, min, max string) Completed {
	return b.Zlexcount().Key(key).Min(min).Max(max).Build()
}

func (b builder) ZIncrByCompleted(key string, increment float64, member string) Completed {
	return b.Zincrby().Key(key).Increment(increment).Member(member).Build()
}

func (b builder) zInter(store ZStore, withScores bool) Completed {
	cmd := b.Arbitrary(KwZInter).Args(strconv.Itoa(len(store.Keys))).Keys(store.Keys...)
	if len(store.Weights) > 0 {
		cmd = cmd.Args(KwWeights)
		for _, w := range store.Weights {
			cmd = cmd.Args(strconv.FormatInt(w, 10))
		}
	}
	if store.Aggregate != "" {
		cmd = cmd.Args(KwAggregate, store.Aggregate)
	}
	if withScores {
		cmd = cmd.Args(KwWithScores)
	}
	return cmd.ReadOnly()
}

func (b builder) ZInterCompleted(store ZStore) Completed           { return b.zInter(store, false) }
func (b builder) ZInterWithScoresCompleted(store ZStore) Completed { return b.zInter(store, true) }

func (b builder) ZInterCardCompleted(limit int64, keys ...string) Completed {
	return b.Zintercard().Numkeys(int64(len(keys))).Key(keys...).Limit(limit).Build()
}

func (b builder) ZInterStoreCompleted(destination string, store ZStore) Completed {
	cmd := b.Arbitrary(KwZInterStore).Keys(destination).Args(strconv.Itoa(len(store.Keys))).Keys(store.Keys...)
	if len(store.Weights) > 0 {
		cmd = cmd.Args(KwWeights)
		for _, w := range store.Weights {
			cmd = cmd.Args(strconv.FormatInt(w, 10))
		}
	}
	if store.Aggregate != "" {
		cmd = cmd.Args(KwAggregate, store.Aggregate)
	}
	return cmd.Build()
}

func (b builder) ZMPopCompleted(order string, count int64, keys ...string) Completed {
	cmd := b.Arbitrary(KwZMPop, strconv.Itoa(len(keys))).Keys(keys...)
	cmd = cmd.Args(order)
	if count > 0 {
		cmd = cmd.Args(KwCount, strconv.FormatInt(count, 10))
	}
	return cmd.Build()
}

func (b builder) ZMScoreCompleted(key string, members ...string) Completed {
	return b.Zmscore().Key(key).Member(members...).Build()
}

func (b builder) ZPopMaxCompleted(key string, count ...int64) Completed {
	switch len(count) {
	case 0:
		return b.Zpopmax().Key(key).Build()
	case 1:
		return b.Zpopmax().Key(key).Count(count[0]).Build()
	default:
		panic("too many arguments")
	}
}

func (b builder) ZPopMinCompleted(key string, count ...int64) Completed {
	switch len(count) {
	case 0:
		return b.Zpopmin().Key(key).Build()
	case 1:
		return b.Zpopmin().Key(key).Count(count[0]).Build()
	default:
		panic("too many arguments")
	}
}

func (b builder) zRangeArgs(withScores bool, z ZRangeArgs) Completed {
	cmd := b.Arbitrary(KwZRange).Keys(z.Key)
	if z.Rev && (z.ByScore || z.ByLex) {
		cmd = cmd.Args(str(z.Stop), str(z.Start))
	} else {
		cmd = cmd.Args(str(z.Start), str(z.Stop))
	}
	if z.ByScore {
		cmd = cmd.Args(KwByScore)
	} else if z.ByLex {
		cmd = cmd.Args(KwByLex)
	}
	if z.Rev {
		cmd = cmd.Args(KwRev)
	}
	if z.Offset != 0 || z.Count != 0 {
		cmd = cmd.Args(KwLimit, strconv.FormatInt(z.Offset, 10), strconv.FormatInt(z.Count, 10))
	}
	if withScores {
		cmd = cmd.Args(KwWithScores)
	}
	return cmd.Build()
}

func (b builder) ZRangeCompleted(key string, start, stop int64) Completed {
	return b.zRangeArgs(false, ZRangeArgs{
		Key:   key,
		Start: start,
		Stop:  stop,
	})
}

func (b builder) ZRangeWithScoresCompleted(key string, start, stop int64) Completed {
	return b.zRangeArgs(true, ZRangeArgs{
		Key:   key,
		Start: start,
		Stop:  stop,
	})
}

func (b builder) ZRangeByScoreCompleted(key string, opt ZRangeBy) Completed {
	if opt.Offset != 0 || opt.Count != 0 {
		return b.Zrangebyscore().Key(key).Min(opt.Min).Max(opt.Max).Limit(opt.Offset, opt.Count).Build()
	} else {
		return b.Zrangebyscore().Key(key).Min(opt.Min).Max(opt.Max).Build()
	}
}

func (b builder) ZRangeByLexCompleted(key string, opt ZRangeBy) Completed {
	if opt.Offset != 0 || opt.Count != 0 {
		return b.Zrangebylex().Key(key).Min(opt.Min).Max(opt.Max).Limit(opt.Offset, opt.Count).Build()
	} else {
		return b.Zrangebylex().Key(key).Min(opt.Min).Max(opt.Max).Build()
	}
}

func (b builder) ZRangeByScoreWithScoresCompleted(key string, opt ZRangeBy) Completed {
	if opt.Offset != 0 || opt.Count != 0 {
		return b.Zrangebyscore().Key(key).Min(opt.Min).Max(opt.Max).Withscores().Limit(opt.Offset, opt.Count).Build()
	} else {
		return b.Zrangebyscore().Key(key).Min(opt.Min).Max(opt.Max).Withscores().Build()
	}
}

func (b builder) ZRangeArgsCompleted(z ZRangeArgs) Completed {
	return b.zRangeArgs(false, z)
}

func (b builder) ZRangeArgsWithScoresCompleted(z ZRangeArgs) Completed {
	return b.zRangeArgs(true, z)
}

func (b builder) ZRangeStoreCompleted(dst string, z ZRangeArgs) Completed {
	cmd := b.Arbitrary(KwZRangeStore).Keys(dst, z.Key)
	if z.Rev && (z.ByScore || z.ByLex) {
		cmd = cmd.Args(str(z.Stop), str(z.Start))
	} else {
		cmd = cmd.Args(str(z.Start), str(z.Stop))
	}
	if z.ByScore {
		cmd = cmd.Args(KwByScore)
	} else if z.ByLex {
		cmd = cmd.Args(KwByLex)
	}
	if z.Rev {
		cmd = cmd.Args(KwRev)
	}
	if z.Offset != 0 || z.Count != 0 {
		cmd = cmd.Args(KwLimit, strconv.FormatInt(z.Offset, 10), strconv.FormatInt(z.Count, 10))
	}
	return cmd.Build()
}

func (b builder) ZRankCompleted(key, member string) Completed {
	return b.Zrank().Key(key).Member(member).Build()
}

func (b builder) ZRankWithScoreCompleted(key, member string) Completed {
	return b.Zrank().Key(key).Member(member).Withscore().Build()
}

func (b builder) ZRemCompleted(key string, members ...any) Completed {
	return b.Zrem().Key(key).Member(argsToSlice(members)...).Build()
}

func (b builder) ZRemRangeByRankCompleted(key string, start, stop int64) Completed {
	return b.Zremrangebyrank().Key(key).Start(start).Stop(stop).Build()
}
func (b builder) ZRemRangeByScoreCompleted(key, min, max string) Completed {
	return b.Zremrangebyscore().Key(key).Min(min).Max(max).Build()
}

func (b builder) ZRemRangeByLexCompleted(key string, min, max string) Completed {
	return b.Zremrangebylex().Key(key).Min(min).Max(max).Build()
}

func (b builder) ZRevRangeCompleted(key string, start, stop int64) Completed {
	return b.Zrevrange().Key(key).Start(start).Stop(stop).Build()
}

func (b builder) ZRevRangeWithScoresCompleted(key string, start, stop int64) Completed {
	return b.Zrevrange().Key(key).Start(start).Stop(stop).Withscores().Build()
}

func (b builder) ZRevRangeByScoreCompleted(key string, opt ZRangeBy) Completed {
	if opt.Offset != 0 || opt.Count != 0 {
		return b.Zrevrangebyscore().Key(key).Max(opt.Max).Min(opt.Min).Limit(opt.Offset, opt.Count).Build()
	} else {
		return b.Zrevrangebyscore().Key(key).Max(opt.Max).Min(opt.Min).Build()
	}
}

func (b builder) ZRevRangeByLexCompleted(key string, opt ZRangeBy) Completed {
	if opt.Offset != 0 || opt.Count != 0 {
		return b.Zrevrangebylex().Key(key).Max(opt.Max).Min(opt.Min).Limit(opt.Offset, opt.Count).Build()
	} else {
		return b.Zrevrangebylex().Key(key).Max(opt.Max).Min(opt.Min).Build()
	}
}

func (b builder) ZRevRangeByScoreWithScoresCompleted(key string, opt ZRangeBy) Completed {
	if opt.Offset != 0 || opt.Count != 0 {
		return b.Zrevrangebyscore().Key(key).Max(opt.Max).Min(opt.Min).Withscores().Limit(opt.Offset, opt.Count).Build()
	} else {
		return b.Zrevrangebyscore().Key(key).Max(opt.Max).Min(opt.Min).Withscores().Build()
	}
}

func (b builder) ZRevRankCompleted(key, member string) Completed {
	return b.Zrevrank().Key(key).Member(member).Build()
}

func (b builder) ZRevRankWithScoreCompleted(key, member string) Completed {
	return b.Zrevrank().Key(key).Member(member).Withscore().Build()
}

func (b builder) ZScoreCompleted(key, member string) Completed {
	return b.Zscore().Key(key).Member(member).Build()
}

func (b builder) zUnion(store ZStore, withScores bool) Completed {
	cmd := b.Arbitrary(KwZUnion).Args(strconv.Itoa(len(store.Keys))).Keys(store.Keys...)
	if len(store.Weights) > 0 {
		cmd = cmd.Args(KwWeights)
		for _, w := range store.Weights {
			cmd = cmd.Args(strconv.FormatInt(w, 10))
		}
	}
	if store.Aggregate != "" {
		cmd = cmd.Args(KwAggregate, store.Aggregate)
	}
	if withScores {
		cmd = cmd.Args(KwWithScores)
	}
	return cmd.ReadOnly()
}

func (b builder) ZUnionStoreCompleted(dest string, store ZStore) Completed {
	cmd := b.Arbitrary(KwZUnionStore).Keys(dest).Args(strconv.Itoa(len(store.Keys))).Keys(store.Keys...)
	if len(store.Weights) > 0 {
		cmd = cmd.Args(KwWeights)
		for _, w := range store.Weights {
			cmd = cmd.Args(strconv.FormatInt(w, 10))
		}
	}
	if store.Aggregate != "" {
		cmd = cmd.Args(KwAggregate, store.Aggregate)
	}
	return cmd.Build()
}

func (b builder) ZUnionCompleted(store ZStore) Completed           { return b.zUnion(store, false) }
func (b builder) ZUnionWithScoresCompleted(store ZStore) Completed { return b.zUnion(store, true) }

func (b builder) ZRandMemberCompleted(key string, count int64) Completed {
	return b.Zrandmember().Key(key).Count(count).Build()
}

func (b builder) ZRandMemberWithScoresCompleted(key string, count int64) Completed {
	return b.Zrandmember().Key(key).Count(count).Withscores().Build()
}

func (b builder) ZDiffCompleted(keys ...string) Completed {
	return b.Zdiff().Numkeys(int64(len(keys))).Key(keys...).Build()
}

func (b builder) ZDiffWithScoresCompleted(keys ...string) Completed {
	return b.Zdiff().Numkeys(int64(len(keys))).Key(keys...).Withscores().Build()
}

func (b builder) ZDiffStoreCompleted(destination string, keys ...string) Completed {
	return b.Zdiffstore().Destination(destination).Numkeys(int64(len(keys))).Key(keys...).Build()
}

func (b builder) ZScanCompleted(key string, cursor uint64, match string, count int64) Completed {
	cmd := b.Arbitrary(KwZScan).Keys(key).Args(strconv.FormatInt(int64(cursor), 10))
	if match != "" {
		cmd = cmd.Args(KwMatch, match)
	}
	if count > 0 {
		cmd = cmd.Args(KwCount, strconv.FormatInt(count, 10))
	}
	return cmd.ReadOnly()
}
