package aternaLove

import (
	"strings"

	"github.com/azer/logger"
	"gopkg.in/olivere/elastic.v6"

	"github.com/oipwg/oip/config"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
)

func init() {
	log.Info("init aterna")
	if !config.IsTestnet() {
		events.SubscribeAsync("flod:floData", onFloData)
		events.SubscribeAsync("specs:aternaLove:alove", onAlove)
		datastore.RegisterMapping("aterna", "aterna.json")
	}
}

func onFloData(floData string, tx *datastore.TransactionData) {
	if tx.Block < 500000 {
		return
	}
	if tx.Block > 1000000 {
		events.Unsubscribe("flod:floData", onFloData)
		events.Unsubscribe("specs:aternaLove:alove", onAlove)
	}

	prefix := "t1:ALOVE>"
	if strings.HasPrefix(floData, prefix) {
		events.Publish("specs:aternaLove:alove", strings.TrimPrefix(floData, prefix), tx)
		return
	}
}

func onAlove(floData string, tx *datastore.TransactionData) {
	chunks := strings.SplitN(floData, "|", 3)
	if len(chunks) != 3 {
		log.Error("invalid aterna", logger.Attrs{"txid": tx.Transaction.Txid, "floData": floData})
		return
	}

	a := Alove{
		Message: chunks[0],
		To:      chunks[1],
		From:    chunks[2],
		TxId:    tx.Transaction.Txid,
	}
	bir := elastic.NewBulkIndexRequest().Index(datastore.Index("aterna")).Type("_doc").Id(tx.Transaction.Txid).Doc(a)
	datastore.AutoBulk.Add(bir)
}

type Alove struct {
	Message string `json:"message"`
	To      string `json:"to"`
	From    string `json:"from"`
	TxId    string `json:"txId"`
}
