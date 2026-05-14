package redisson

// Deprecated: 旧的 ALL_CAPS 命名常量；保留为新 Kw* 常量的 alias。
// 请改用 Kw* 名称（详见 keywords.go）。
//
//nolint:revive,staticcheck // legacy aliases
const (
	XXX_BITFIELD             = KwBitField
	XXX_SCAN                 = KwScan
	XXX_HSCAN                = KwHScan
	XXX_SSCAN                = KwSScan
	XXX_ZSCAN                = KwZScan
	XXX_MATCH                = KwMatch
	XXX_COUNT                = KwCount
	XXX_GEORADIUSBYMEMBER    = KwGeoRadiusByMember
	XXX_GEORADIUSBYMEMBER_RO = KwGeoRadiusByMemberRO
	XXX_GEORADIUS            = KwGeoRadius
	XXX_GEORADIUS_RO         = KwGeoRadiusRO
	XXX_GEOSEARCH            = KwGeoSearch
	XXX_GEOSEARCHSTORE       = KwGeoSearchStore
	XXX_STORE                = KwStore
	XXX_STOREDIST            = KwStoreDist
	XXX_LMOVE                = KwLMove
	XXX_LMPOP                = KwLMPop
	XXX_LPOS                 = KwLPos
	XXX_RANK                 = KwRank
	XXX_MAXLEN               = KwMaxLen
	XXX_MINID                = KwMinID
	XXX_FUNCTION             = KwFunction
	XXX_LIST                 = KwList
	XXX_LIBRARYNAME          = KwLibraryName
	XXX_WITHCODE             = KwWithCode
	XXX_XADD                 = KwXAdd
	XXX_NOMKSTREAM           = KwNoMkStream
	XXX_LIMIT                = KwLimit
	XXX_XPENDING             = KwXPending
	XXX_IDLE                 = KwIdle
	XXX_XTRIM                = KwXTrim
	XXX_SET                  = KwSet
	XXX_KEEPTTL              = KwKeepTTL
	XXX_EXAT                 = KwExat
	XXX_PX                   = KwPx
	XXX_EX                   = KwEx
	XXX_GET                  = KwGet
	XXX_ZADD                 = KwZAdd
	XXX_GT                   = KwGt
	XXX_LT                   = KwLt
	XXX_CH                   = KwCh
	XXX_INCR                 = KwIncr
	XXX_ZINTER               = KwZInter
	XXX_ZINTERSTORE          = KwZInterStore
	XXX_WEIGHTS              = KwWeights
	XXX_AGGREGATE            = KwAggregate
	XXX_WITHSCORES           = KwWithScores
	XXX_ZMPOP                = KwZMPop
	XXX_ZRANGE               = KwZRange
	XXX_ZRANGESTORE          = KwZRangeStore
	XXX_ZUNION               = KwZUnion
	XXX_ZUNIONSTORE          = KwZUnionStore
	XXX_BYSCORE              = KwByScore
	XXX_BY                   = KwBy
	XXX_BYLEX                = KwByLex
	XXX_REV                  = KwRev
	XXX_BYRADIUS             = KwByRadius
	XXX_BYBOX                = KwByBox
	XXX_FROMMEMBER           = KwFromMember
	XXX_FROMLONLAT           = KwFromLonLat
	XXX_WITHCOORD            = KwWithCoord
	XXX_WITHDIST             = KwWithDist
	XXX_WITHHASH             = KwWithHash
	XXX_ANY                  = KwAny
	XXX_SERVER               = KwServer
	XXX_CLUSTER              = KwCluster
	XXX_ALPHA                = KwAlpha
	XXX_BLMOVE               = KwBLMove
)
