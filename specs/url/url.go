package url

import (
	"strings"

	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
)

func init() {
	events.SubscribeAsync("flod:floData", onFloData)
	events.SubscribeAsync("specs:url", onUrl)
}

func onFloData(floData string, tx *datastore.TransactionData) {
	if strings.HasPrefix(floData, "http://") || strings.HasPrefix(floData, "https://") {
		events.Publish("specs:url", floData)
		return
	}
}

func onUrl(floData string) {

}
