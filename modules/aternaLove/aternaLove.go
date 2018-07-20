package aternaLove

import (
	"strings"

	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"gopkg.in/olivere/elastic.v6"
)

func init() {
	log.Info("init aterna")
	events.Bus.SubscribeAsync("flo:floData", onFloData, false)
	events.Bus.SubscribeAsync("modules:aternaLove:alove", onAlove, false)
	datastore.RegisterMapping("aterna", aternaMapping)
}

func onFloData(floData string, tx datastore.TransactionData) {
	if tx.Block < 500000 || tx.Block > 1000000 {
		return
	}

	prefix := "t1:ALOVE>"
	if strings.HasPrefix(floData, prefix) {
		events.Bus.Publish("modules:aternaLove:alove", strings.TrimPrefix(floData, prefix))
		return
	}
}

func onAlove(floData string) {
	var message, to, from string

	chunks := strings.SplitN(floData, "|", 2)
	lc := len(chunks)
	if lc == 3 {
		message = chunks[0]
		to = chunks[1]
		from = chunks[2]
	} else if lc > 3 {
		message = strings.Join(chunks[0:lc-2], "|")
		to = chunks[lc-2]
		from = chunks[lc-1]
	} else {
		return
	}

	a := Alove{
		Message: message,
		From:    from,
		To:      to,
		// TxId: txid,
	}
	bir := elastic.NewBulkIndexRequest().Index("aterna").Type("_doc"). /*Id(txid).*/ Doc(a)
	datastore.AutoBulk.Add(bir)
}

type Alove struct {
	Message string `json:"message"`
	To      string `json:"to"`
	From    string `json:"from"`
	TxId    string `json:"txId"`
}

const aternaMapping = `{
  "settings": {
  },
  "mappings": {
    "_doc": {
      "dynamic": "strict",
      "properties": {
        "from": {
          "type": "text",
          "fields": {
            "keyword": {
              "type": "keyword",
              "ignore_above": 256
            }
          }
        },
        "message": {
          "type": "text",
          "fields": {
            "keyword": {
              "type": "keyword",
              "ignore_above": 256
            }
          }
        },
        "to": {
          "type": "text",
          "fields": {
            "keyword": {
              "type": "keyword",
              "ignore_above": 256
            }
          }
        },
        "txId": {
          "type": "text",
          "fields": {
            "keyword": {
              "type": "keyword",
              "ignore_above": 256
            }
          }
        }
      }
    }
  }
}`