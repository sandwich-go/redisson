package redisson

func (b builder) EvalCompleted(script string, keys []string, args ...any) Completed {
	return b.Eval().Script(script).Numkeys(int64(len(keys))).Key(keys...).Arg(argsToSlice(args)...).Build()
}

func (b builder) EvalShaCompleted(sha1 string, keys []string, args ...any) Completed {
	return b.Evalsha().Sha1(sha1).Numkeys(int64(len(keys))).Key(keys...).Arg(argsToSlice(args)...).Build()
}

func (b builder) EvalROCompleted(script string, keys []string, args ...any) Completed {
	return b.EvalRo().Script(script).Numkeys(int64(len(keys))).Key(keys...).Arg(argsToSlice(args)...).Build()
}

func (b builder) EvalShaROCompleted(sha1 string, keys []string, args ...any) Completed {
	return b.EvalshaRo().Sha1(sha1).Numkeys(int64(len(keys))).Key(keys...).Arg(argsToSlice(args)...).Build()
}

func (b builder) FunctionListCompleted(q FunctionListQuery) Completed {
	cmd := b.Arbitrary(KwFunction, KwList)
	if q.LibraryNamePattern != "" {
		cmd = cmd.Args(KwLibraryName, q.LibraryNamePattern)
	}
	if q.WithCode {
		cmd = cmd.Args(KwWithCode)
	}
	return cmd.Build()
}

func (b builder) FunctionDumpCompleted() Completed { return b.FunctionDump().Build() }

func (b builder) FCallCompleted(function string, keys []string, args ...any) Completed {
	return b.Fcall().Function(function).Numkeys(int64(len(keys))).Key(keys...).Arg(argsToSlice(args)...).Build()
}

func (b builder) FCallROCompleted(function string, keys []string, args ...any) Completed {
	return b.FcallRo().Function(function).Numkeys(int64(len(keys))).Key(keys...).Arg(argsToSlice(args)...).Build()
}

func (b builder) ACLDryRunCompleted(username string, command ...any) Completed {
	return b.AclDryrun().Username(username).Command(command[0].(string)).Arg(argsToSlice(command[1:])...).Build()
}
