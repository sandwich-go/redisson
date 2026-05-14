package redisson

func (b builder) PFAddCompleted(key string, els ...any) Completed {
	return b.Pfadd().Key(key).Element(argsToSlice(els)...).Build()
}

func (b builder) PFCountCompleted(keys ...string) Completed {
	return b.Pfcount().Key(keys...).Build()
}

func (b builder) PFMergeCompleted(dest string, keys ...string) Completed {
	return b.Pfmerge().Destkey(dest).Sourcekey(keys...).Build()
}
