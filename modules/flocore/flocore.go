package flocore

import (
	"fmt"
	"github.com/bitspill/flod/rpcclient"
)

func AddFloCore(host, user, pass string, tls bool) (*rpcclient.Client, error) {
	cfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
		DisableTLS:   tls,
		HTTPPostMode: true,
	}
	c, err := rpcclient.New(cfg, nil)
	fmt.Println(c)
	return c, err
}

