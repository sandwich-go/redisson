package redisson

import "github.com/alicebob/miniredis/v2"

type mock struct {
	*resp3
}

func connectMock(v ConfInterface, h handler) (*mock, error) {
	_ = v.ApplyOption(WithAddrs(miniredis.RunT(v.GetT()).Addr()))
	c, err := connectResp3(v, h)
	if err != nil {
		return nil, err
	}
	return &mock{resp3: c}, nil
}
