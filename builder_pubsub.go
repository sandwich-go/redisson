package redisson

func (b builder) PublishCompleted(channel string, message any) Completed {
	return b.Publish().Channel(channel).Message(str(message)).Build()
}

func (b builder) SPublishCompleted(channel string, message any) Completed {
	return b.Spublish().Channel(channel).Message(str(message)).Build()
}

func (b builder) PubSubChannelsCompleted(pattern string) Completed {
	return b.PubsubChannels().Pattern(pattern).Build()
}

func (b builder) PubSubNumSubCompleted(channels ...string) Completed {
	return b.PubsubNumsub().Channel(channels...).Build()
}

func (b builder) PubSubNumPatCompleted() Completed {
	return b.PubsubNumpat().Build()
}

func (b builder) PubSubShardChannelsCompleted(pattern string) Completed {
	return b.PubsubShardchannels().Pattern(pattern).Build()
}

func (b builder) PubSubShardNumSubCompleted(channels ...string) Completed {
	return b.PubsubShardnumsub().Channel(channels...).Build()
}
