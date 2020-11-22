package flotorizer

import (
	"strings"

	"gopkg.in/olivere/elastic.v6"

	"github.com/oipwg/oip/config"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
)

func init() {
	log.Info("init flotorizer")
	if !config.IsTestnet() {
		events.SubscribeAsync("flod:floData", onFloData)
		events.SubscribeAsync("specs:flotorizer:flotorized", onFlotorized)
		datastore.RegisterMapping("flotorizer", "flotorizer.json")
	}
}

func onFloData(floData string, tx *datastore.TransactionData) {
	if tx.Block < 1500000 {
		return
	}
	prefix := "This document has been flotorized: "
	if strings.HasPrefix(floData, prefix) {
		events.Publish("specs:flotorizer:flotorized", strings.TrimPrefix(floData, prefix))
		return
	}
}

func onFlotorized(floData string) {
	f := Flotorized{
		Hash: floData,
		// TxId: txid,
	}
	bir := elastic.NewBulkIndexRequest().Index(datastore.Index("flotorizer")).Type("_doc"). /*Id(txid).*/ Doc(f)
	datastore.AutoBulk.Add(bir)
}

type Flotorized struct {
	Hash string `json:"hash"`
	TxId string `json:"txId"`
}
