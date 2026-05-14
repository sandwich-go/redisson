package redisson

func (b builder) CommandCompleted() Completed { return b.Command().Build() }

func (b builder) CommandListCompleted(filter FilterBy) Completed {
	if filter.Module != "" {
		return b.CommandList().FilterbyModuleName(filter.Module).Build()
	} else if filter.Pattern != "" {
		return b.CommandList().FilterbyPatternPattern(filter.Pattern).Build()
	} else if filter.ACLCat != "" {
		return b.CommandList().FilterbyAclcatCategory(filter.ACLCat).Build()
	} else {
		return b.CommandList().Build()
	}
}

func (b builder) CommandGetKeysCompleted(commands ...any) Completed {
	return b.CommandGetkeys().Command(commands[0].(string)).Arg(argsToSlice(commands[1:])...).Build()
}

func (b builder) CommandGetKeysAndFlagsCompleted(commands ...any) Completed {
	return b.CommandGetkeysandflags().Command(commands[0].(string)).Arg(argsToSlice(commands[1:])...).Build()
}

func (b builder) ConfigGetCompleted(parameter string) Completed {
	return b.ConfigGet().Parameter(parameter).Build()
}

func (b builder) InfoCompleted(section ...string) Completed {
	return b.Info().Section(section...).Build()
}

func (b builder) LastSaveCompleted() Completed { return b.Lastsave().Build() }

func (b builder) DebugObjectCompleted(key string) Completed {
	return b.DebugObject().Key(key).Build()
}

func (b builder) MemoryUsageCompleted(key string, samples ...int64) Completed {
	switch len(samples) {
	case 0:
		return b.MemoryUsage().Key(key).Build()
	case 1:
		return b.MemoryUsage().Key(key).Samples(samples[0]).Build()
	default:
		panic("too many arguments")
	}
}

func (b builder) TimeCompleted() Completed {
	return b.Time().Build()
}
