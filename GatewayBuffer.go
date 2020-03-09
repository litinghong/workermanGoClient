package workermanGoClient

import "github.com/techoner/gophp/serialize"

type GatewayBuffer struct {
	GatewayAddress string
	Buffer         []byte
}

func (g *GatewayBuffer) UnMarshal() (interface{}, error) {
	return serialize.UnMarshal(g.Buffer)
}
