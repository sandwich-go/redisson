package redisson

import (
	"strings"
)

func (b builder) GetBitCompleted(key string, offset int64) Completed {
	return b.Getbit().Key(key).Offset(offset).Build()
}

func (b builder) SetBitCompleted(key string, offset int64, value int64) Completed {
	return b.Setbit().Key(key).Offset(offset).Value(value).Build()
}

func (b builder) BitCountCompleted(key string, bc *BitCount) Completed {
	if bc == nil {
		return b.Bitcount().Key(key).Build()
	}
	if bc.Unit == "" {
		return b.Bitcount().Key(key).Start(bc.Start).End(bc.End).Build()
	}
	switch strings.ToUpper(bc.Unit) {
	case BYTE:
		return b.Bitcount().Key(key).Start(bc.Start).End(bc.End).Byte().Build()
	case BIT:
		return b.Bitcount().Key(key).Start(bc.Start).End(bc.End).Bit().Build()
	default:
		panic(NewParameterError("invalid unit %s", bc.Unit))
	}
}

func (b builder) BitOpAndCompleted(destKey string, keys ...string) Completed {
	return b.Bitop().And().Destkey(destKey).Key(keys...).Build()
}

func (b builder) BitOpOrCompleted(destKey string, keys ...string) Completed {
	return b.Bitop().Or().Destkey(destKey).Key(keys...).Build()
}

func (b builder) BitOpXorCompleted(destKey string, keys ...string) Completed {
	return b.Bitop().Xor().Destkey(destKey).Key(keys...).Build()
}

func (b builder) BitOpNotCompleted(destKey string, key string) Completed {
	return b.Bitop().Not().Destkey(destKey).Key(key).Build()
}

func (b builder) BitPosCompleted(key string, bit int64, pos ...int64) Completed {
	switch len(pos) {
	case 0:
		return b.Bitpos().Key(key).Bit(bit).Build()
	case 1:
		return b.Bitpos().Key(key).Bit(bit).Start(pos[0]).Build()
	case 2:
		return b.Bitpos().Key(key).Bit(bit).Start(pos[0]).End(pos[1]).Build()
	default:
		panic(NewParameterError("too many arguments"))
	}
}

func (b builder) BitPosSpanCompleted(key string, bit, start, end int64, span string) Completed {
	if strings.ToUpper(span) == BIT {
		return b.Bitpos().Key(key).Bit(bit).Start(start).End(end).Bit().Build()
	} else {
		return b.Bitpos().Key(key).Bit(bit).Start(start).End(end).Byte().Build()
	}
}

func (b builder) BitFieldCompleted(key string, args ...any) Completed {
	cmd := b.Arbitrary(KwBitField).Keys(key)
	for _, v := range args {
		cmd = cmd.Args(str(v))
	}
	return cmd.Build()
}
