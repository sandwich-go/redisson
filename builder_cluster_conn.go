package redisson

func (b builder) ClusterReplicasCompleted(nodeID string) Completed {
	return b.ClusterReplicas().NodeId(nodeID).Build()
}

func (b builder) ClientGetNameCompleted() Completed   { return b.ClientGetname().Build() }
func (b builder) ClientListCompleted() Completed      { return b.ClientList().Build() }
func (b builder) EchoCompleted(message any) Completed { return b.Echo().Message(str(message)).Build() }
func (b builder) PingCompleted() Completed            { return b.Ping().Build() }
