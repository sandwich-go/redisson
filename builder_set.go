package redisson

import (
	"strconv"
)

func (b builder) SAddCompleted(key string, members ...any) Completed {
	cmd := b.Sadd().Key(key).Member()
	for _, m := range argsToSlice(members) {
		cmd = cmd.Member(str(m))
	}
	return cmd.Build()
}

func (b builder) SCardCompleted(key string) Completed {
	return b.Scard().Key(key).Build()
}

func (b builder) SDiffCompleted(keys ...string) Completed {
	return b.Sdiff().Key(keys...).Build()
}

func (b builder) SDiffStoreCompleted(destination string, keys ...string) Completed {
	return b.Sdiffstore().Destination(destination).Key(keys...).Build()
}

func (b builder) SInterCompleted(keys ...string) Completed {
	return b.Sinter().Key(keys...).Build()
}

func (b builder) SInterStoreCompleted(destination string, keys ...string) Completed {
	return b.Sinterstore().Destination(destination).Key(keys...).Build()
}

func (b builder) SInterCardCompleted(limit int64, keys ...string) Completed {
	return b.Sintercard().Numkeys(int64(len(keys))).Key(keys...).Limit(limit).Build()
}

func (b builder) SIsMemberCompleted(key string, member any) Completed {
	return b.Sismember().Key(key).Member(str(member)).Build()
}

func (b builder) SMIsMemberCompleted(key string, members ...any) Completed {
	return b.Smismember().Key(key).Member(argsToSlice(members)...).Build()
}

func (b builder) SMembersCompleted(key string) Completed {
	return b.Smembers().Key(key).Build()
}

func (b builder) SMoveCompleted(source, destination string, member any) Completed {
	return b.Smove().Source(source).Destination(destination).Member(str(member)).Build()
}

func (b builder) SPopCompleted(key string) Completed {
	return b.Spop().Key(key).Build()
}

func (b builder) SPopNCompleted(key string, count int64) Completed {
	return b.Spop().Key(key).Count(count).Build()
}

func (b builder) SRandMemberCompleted(key string) Completed {
	return b.Srandmember().Key(key).Build()
}

func (b builder) SRandMemberNCompleted(key string, count int64) Completed {
	return b.Srandmember().Key(key).Count(count).Build()
}

func (b builder) SRemCompleted(key string, members ...any) Completed {
	return b.Srem().Key(key).Member(argsToSlice(members)...).Build()
}

func (b builder) SScanCompleted(key string, cursor uint64, match string, count int64) Completed {
	cmd := b.Arbitrary(KwSScan).Keys(key).Args(strconv.FormatInt(int64(cursor), 10))
	if match != "" {
		cmd = cmd.Args(KwMatch, match)
	}
	if count > 0 {
		cmd = cmd.Args(KwCount, strconv.FormatInt(count, 10))
	}
	return cmd.ReadOnly()
}

func (b builder) SUnionCompleted(keys ...string) Completed {
	return b.Sunion().Key(keys...).Build()
}

func (b builder) SUnionStoreCompleted(destination string, keys ...string) Completed {
	return b.Sunionstore().Destination(destination).Key(keys...).Build()
}
