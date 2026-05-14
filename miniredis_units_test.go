//go:build miniredis_test

package redisson

// miniredis_units_test.go 收录 miniredis 兼容的命令族 TestUnit 子集。
//
// 选取标准: 不依赖真实时钟（无 TTL/expiration 等待）、命令 miniredis 已实现。
// 完整命令套见 cmd_*_test.go（//go:build integration），需要真实 Redis。

// miniredisStringUnits 仅包含单 key 的 String 命令；
// 跨 key 命令（MSET/MGET/MSETNX）在 rueidis 内部对未带 hashtag 的多 key 会 panic（CROSSSLOT 校验），
// 故不放入 miniredis 套。
func miniredisStringUnits() []TestUnit {
	return []TestUnit{
		{CommandAppend, testAppend},
		{CommandDecr, testDecr},
		{CommandDecrBy, testDecrBy},
		{CommandGetDel, testGetDel},
		{CommandGetSet, testGetSet},
		{CommandIncr, testIncr},
		{CommandIncrBy, testIncrBy},
		{CommandIncrByFloat, testIncrByFloat},
		{CommandSetRange, testSetRange},
		{CommandGet, testGet},
		{CommandGetRange, testGetRange},
		{CommandStrLen, testStrLen},
	}
}

func miniredisHashUnits() []TestUnit {
	return []TestUnit{
		{CommandHDel, testHDel},
		{CommandHGet, testHGet},
		{CommandHGetAll, testHGetAll},
		{CommandHIncrBy, testHIncrBy},
		{CommandHIncrByFloat, testHIncrByFloat},
		{CommandHKeys, testHKeys},
		{CommandHLen, testHLen},
		{CommandHMGet, testHMGet},
		{CommandHSet, testHSet},
		{CommandHSetNX, testHSetNX},
		{CommandHVals, testHVals},
	}
}

// miniredisSetUnits 仅含单 key 的 Set 命令；
// 多 key 集合命令（SDIFF/SINTER/SUNION 系列、SMOVE）会触发 rueidis CROSSSLOT 校验。
func miniredisSetUnits() []TestUnit {
	return []TestUnit{
		{CommandSAdd, testSAdd},
		{CommandSCard, testSCard},
		{CommandSIsMember, testSIsMember},
		{CommandSMembers, testSMembers},
		{CommandSPop, testSPop},
		{CommandSRandMember, testSRandMember},
		{CommandSRem, testSRem},
	}
}
