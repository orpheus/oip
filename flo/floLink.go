package flo

import (
	"context"
	"github.com/azer/logger"
	"github.com/bitspill/flod/flojson"
	"github.com/bitspill/flod/rpcclient"
	"github.com/bitspill/flod/wire"
	"github.com/bitspill/floutil"
	"github.com/cloudflare/backoff"
	"github.com/oipwg/oip/config"
	"github.com/oipwg/oip/events"
	"github.com/orpheus/oip/links"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"io/ioutil"
	"net"
	"time"
)

func init() {
	links.AddLink(flodLink)
}

type Flod struct {
	ID     links.LinkerID
	Client *rpcclient.Client
}

var flodLink = &Flod{
	ID: "FLOD",
}

func (l *Flod) GetId() links.LinkerID {
	return l.ID
}

func (l *Flod) AddNode() error {
	host := viper.GetString("flod.host")
	user := viper.GetString("flod.user")
	pass := viper.GetString("flod.pass")
	tls := viper.GetBool("flod.tls")

	var certs []byte
	var err error

	if tls {
		certFile := config.GetFilePath("flod.certFile")
		certs, err = ioutil.ReadFile(certFile)
		if err != nil {
			return errors.Wrap(err, "unable to read rpc.cert")
		}
	}

	ntfnHandlers := rpcclient.NotificationHandlers{
		OnFilteredBlockConnected: func(height int32, header *wire.BlockHeader, txns []*floutil.Tx) {
			log.Info("Block connected: %v (%d) %v",
				header.BlockHash(), height, header.Timestamp)
			events.Publish("flo:notify:onFilteredBlockConnected", height, header, txns)
		},
		OnFilteredBlockDisconnected: func(height int32, header *wire.BlockHeader) {
			log.Info("Block disconnected:  %v (%d) %v",
				header.BlockHash(), height, header.Timestamp)
			events.Publish("flo:notify:onFilteredBlockDisconnected", height, header)
		},
		OnTxAcceptedVerbose: func(txDetails *flojson.TxRawResult) {
			log.Info("Incoming TX: %v (Block: %v) floData: %v", txDetails.Txid, txDetails.BlockHash, txDetails.FloData)
			events.Publish("flo:notify:onTxAcceptedVerbose", txDetails)
		},
	}

	cfg := &rpcclient.ConnConfig{
		Host:         host,
		Endpoint:     "ws",
		User:         user,
		Pass:         pass,
		DisableTLS:   !tls,
		Certificates: certs,
	}
	c, err := rpcclient.New(cfg, &ntfnHandlers)
	if err != nil {
		return errors.Wrap(err, "unable to create new rpc client")
	}
	l.Client = c
	return nil
}

func (l *Flod) RemoveNode() {
	// disconnect client
	l.Client.Disconnect()
	// remove from linkMap
	links.RemoveLink(l)
}

func (l *Flod) WaitForNode(ctx context.Context, host, user, pass string, tls bool) error {
	attempts := 0
	a := logger.Attrs{"host": host, "attempts": attempts}
	b := backoff.NewWithoutJitter(10*time.Minute, 1*time.Second)
	t := log.Timer()
	defer t.End("WaitForFlod", a)
	for {
		attempts++
		a["attempts"] = attempts
		log.Info("attempting connection to flod", a)
		err := AddFlod(host, user, pass, tls)
		if err != nil {
			a["err"] = err
			log.Error("unable to connect to flod", a)
			delete(a, "err")
			c := errors.Cause(err)
			if _, ok := c.(*net.OpError); !ok {
				// not a network error, something else is wrong
				return err
			}
			// it's a network error, delay and retry
			d := b.Duration()
			a["delay"] = d
			log.Info("delaying connection to flod retry", a)
			delete(a, "delay")
			select {
			case <-ctx.Done():
				a["err"] = ctx.Err()
				log.Error("context timeout/cancelled", a)
				return ctx.Err()
			case <-time.After(d):
				// loop around for another try
			}
		} else {
			break
		}
	}
	return nil
}
