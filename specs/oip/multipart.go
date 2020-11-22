package oip

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/azer/logger"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/oipwg/proto/go/pb_oip"
	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v6"

	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
	"github.com/oipwg/oip/flo"
	"github.com/oipwg/oip/httpapi"
	oipSync "github.com/oipwg/oip/sync"
)

const multipartIndex = "oip-multipart-single"

var multiPartCommitMutex sync.Mutex
var mpRouter = httpapi.NewSubRoute("/multipart")

var previousMultipartCount int

func init() {
	log.Info("init multipart")
	datastore.RegisterMapping(multipartIndex, "multipart.json")
	events.SubscribeAsync("specs:oip:multipartSingle", onMultipartSingle)
	events.SubscribeAsync("specs:oip:multipartProto", onMultipartProto)
	events.SubscribeAsync("datastore:commit", onDatastoreCommit)

	mpRouter.HandleFunc("/get/ref/{ref:[a-f0-9]+}", handleGetRef)
	mpRouter.HandleFunc("/get/id/{id:[a-f0-9]+}", handleGetId)
}

var (
	mpIndices = []string{multipartIndex}
	mpFsc     = elastic.NewFetchSourceContext(true).Include("*")
)

func handleGetId(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	q := elastic.NewBoolQuery().Must(
		elastic.NewPrefixQuery("meta.txid", opts["id"]),
	)

	searchService := httpapi.BuildCommonSearchService(r.Context(), mpIndices, q, []elastic.SortInfo{{Field: "meta.time", Ascending: false}}, mpFsc)
	httpapi.RespondSearch(r.Context(), w, searchService)
}

func handleGetRef(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	q := elastic.NewBoolQuery().Must(
		elastic.NewPrefixQuery("reference", opts["ref"]),
	)

	searchService := httpapi.BuildCommonSearchService(r.Context(), mpIndices, q, []elastic.SortInfo{{Field: "meta.time", Ascending: false}}, mpFsc)
	httpapi.RespondSearch(r.Context(), w, searchService)
}

func onDatastoreCommit() {
	// If we are still working on the initial sync, don't attempt to complete multiparts.
	if oipSync.IsInitialSync {
		return
	}

	multiPartCommitMutex.Lock()
	defer multiPartCommitMutex.Unlock()

	wasInitialSync := oipSync.IsInitialSync

moreMultiparts:
	multiparts := make(map[string]Multipart)
	var after []interface{}

	after, err := queryMultiparts(multiparts, after)
	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
	}

	// Check if there are no more multiparts, if so, mark the multipart sync as complete so that we can start marking Edits as invalid
	if len(multiparts) == 0 {
		oipSync.MultipartSyncComplete = true
	} else {
		oipSync.MultipartSyncComplete = false
	}

	potentialChanges := false
	for k, mp := range multiparts {
		if mp.Count >= mp.Total {
			if mp.Count > mp.Total {
				log.Info("extra parts", k)
			}
			tryCompleteMultipart(mp)
			potentialChanges = true
		}
	}

	if potentialChanges {
		events.Publish("specs:oip:mpCompleted")
	}

	if after != nil {
		goto moreMultiparts
	}

	// If we are not still syncing for the first time, and the remaining count is exactly the same as last time we checked,
	// then it is a good indicator that these Multiparts are stale
	if !wasInitialSync && previousMultipartCount == len(multiparts) && previousMultipartCount < 10000 {
		// ToDo: Consider re-enabling after further tests under high volume
		markStale()
	}

	previousMultipartCount = len(multiparts)
}

func queryMultiparts(multiparts map[string]Multipart, after []interface{}) ([]interface{}, error) {
	var nextAfter []interface{}
	searchSize := 10000

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.complete", false),
		elastic.NewTermQuery("meta.stale", false),
	)
	search := datastore.Client().
		Search(datastore.Index(multipartIndex)).
		Type("_doc").
		Query(q).
		Size(searchSize).
		Sort("meta.time", false).
		Sort("reference", false)

	if after != nil {
		search.SearchAfter(after...)
	}

	results, err := search.Do(context.TODO())
	if err != nil {
		return nil, err
	}

	log.Info("Collecting multiparts to attempt assembly", logger.Attrs{"newParts": len(results.Hits.Hits), "totalParts": len(results.Hits.Hits) + len(multiparts)})

	for i, v := range results.Hits.Hits {
		var mps MultipartSingle
		err := json.Unmarshal(*v.Source, &mps)
		if err != nil {
			log.Info("failed to unmarshal elastic hit", logger.Attrs{"err": err})
			continue
		}
		mp, ok := multiparts[mps.Reference]
		if !ok {
			mp.Total = mps.Max + 1
		}
		if mps.Part < mp.Total {
			mp.Count++
			mp.Parts = append(mp.Parts, mps)
			multiparts[mps.Reference] = mp
		}

		if i == len(results.Hits.Hits)-1 && len(results.Hits.Hits) == searchSize {
			nextAfter = v.Sort
		}
	}

	return nextAfter, nil
}

func tryCompleteMultipart(mp Multipart) {
	if mp.Total > 1000 {
		log.Info("multipart has too many parts", logger.Attrs{"txid": mp.Parts[0].Meta.Txid, "part": mp.Total})
		return
	}

	rebuild := make([]string, mp.Total)
	var part0 MultipartSingle
	for i := range mp.Parts {
		value := &mp.Parts[i]
		if value.Part == 0 {
			part0 = *value
		}
		if rebuild[value.Part] != "" {
			log.Info("duplicate multipart", logger.Attrs{"txid": value.Meta.Txid, "part": value.Part})
		}
		rebuild[value.Part] = value.Data
	}

	for _, v := range rebuild {
		if v == "" {
			return
		}
	}

	log.Info("completed mp ", logger.Attrs{"reference": mp.Parts[0].Reference})

	dataString := strings.Join(rebuild, "")

	newVal := map[string]interface{}{
		"meta": map[string]interface{}{
			"complete":  true,
			"assembled": dataString,
		},
	}

	for _, part := range mp.Parts {
		upd := elastic.NewBulkUpdateRequest().Index(datastore.Index(multipartIndex)).Type("_doc").Id(part.Meta.Txid).Doc(newVal)
		datastore.AutoBulk.Add(upd)
	}

	events.Publish("flod:floData", dataString, part0.Meta.Tx)

	log.Info("marked as completed", logger.Attrs{"reference": part0.Reference})
}

func onMultipartSingle(floData string, tx *datastore.TransactionData) {
	ms, err := multipartSingleFromString(floData)
	if err != nil {
		log.Error("multipartSingleFromString error", logger.Attrs{"err": err, "txid": tx.Transaction.Txid})
		return
	}

	if ms.Part == 0 {
		if len(tx.Transaction.Txid) > 10 {
			ms.Reference = tx.Transaction.Txid[0:10]
		} else {
			ms.Reference = tx.Transaction.Txid
		}
	} else {
		if len(ms.Reference) > 10 {
			ms.Reference = ms.Reference[0:10]
		}
	}

	ms.Meta = MSMeta{
		Block:     tx.Block,
		BlockHash: tx.BlockHash,
		Complete:  false,
		Time:      tx.Transaction.Time,
		Txid:      tx.Transaction.Txid,
	}

	if ms.Part == 0 {
		ms.Meta.Tx = tx
	}

	bir := elastic.NewBulkIndexRequest().Index(datastore.Index(multipartIndex)).Type("_doc").Doc(ms).Id(tx.Transaction.Txid)
	datastore.AutoBulk.Add(bir)
}

func multipartSingleFromString(s string) (MultipartSingle, error) {
	var ret MultipartSingle

	// trim prefix off
	s = strings.TrimPrefix(s, "alexandria-media-multipart(")
	s = strings.TrimPrefix(s, "oip-mp(")

	comChunks := strings.Split(s, "):")
	if len(comChunks) < 2 {
		return ret, errors.New("malformed multi-part")
	}

	metaString := comChunks[0]
	dataString := strings.Join(comChunks[1:], "):")

	meta := strings.Split(metaString, ",")
	lm := len(meta)
	// 4 if omitting reference, 5 with all fields, 6 if erroneous fluffy-enigma trailing comma
	if lm != 4 && lm != 5 && lm != 6 {
		return ret, errors.New("malformed multi-part meta")
	}

	// check part and max
	partS := meta[0]
	part, err := strconv.Atoi(partS)
	if err != nil {
		return ret, errors.New("cannot convert part to int")
	}
	maxS := meta[1]
	max, err2 := strconv.Atoi(maxS)
	if err2 != nil {
		return ret, errors.New("cannot convert max to int")
	}

	if max <= 0 {
		return ret, errors.New("max must be positive")
	}

	if part > max {
		return ret, errors.New("part must not exceed max")
	}

	// get and check address
	address := meta[2]
	if ok, err := flo.CheckAddress(address); !ok {
		return ret, errors.Wrap(err, "ErrInvalidAddress")
	}

	reference := meta[3]
	signature := meta[lm-1]
	if signature == "" {
		// fluffy-enigma for a while appended an erroneous trailing comma
		signature = meta[lm-2]
	}

	// signature pre-image is <part>-<max>-<address>-<txid>-<data>
	// in the case of multipart[0], txid is 64 zeros
	// in the case of multipart[n], where n != 0, txid is the reference txid (from multipart[0])
	preimage := partS + "-" + maxS + "-" + address + "-" + reference + "-" + dataString

	if ok, err := flo.CheckSignature(address, signature, preimage); !ok {
		if part != 0 {
			return ret, errors.Wrap(err, "ErrBadSignature")
		}
		preimage := partS + "-" + maxS + "-" + address + "-" + strings.Repeat("0", 64) + "-" + dataString
		if ok, err := flo.CheckSignature(address, signature, preimage); !ok {
			return ret, errors.Wrap(err, "ErrBadSignature")
		}
	}

	if max == 0 {
		panic(s)
	}

	ret = MultipartSingle{
		Part:      uint32(part),
		Max:       uint32(max),
		Reference: reference,
		Address:   address,
		Signature: signature,
		Data:      dataString,
	}

	return ret, nil
}

func markStale() {
	s := elastic.NewScript("ctx._source.meta.stale=true;").Type("inline").Lang("painless")

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.complete", false),
		elastic.NewTermQuery("meta.stale", false),
		elastic.NewRangeQuery("meta.time").Lte("now-1w"),
	)
	cuq := datastore.Client().UpdateByQuery(datastore.Index(multipartIndex)).Query(q).
		Type("_doc").Script(s).Refresh("true")

	res, err := cuq.Do(context.TODO())
	if err != nil {
		log.Error("unable to mark stale", logger.Attrs{"err": err})
		return
	}
	log.Info("mark stale complete", logger.Attrs{"total": res.Total, "took": res.Took, "updated": res.Updated})
}

func onMultipartProto(msg *pb_oip.SignedMessage, tx *datastore.TransactionData) {
	ms := MultipartSingle{}

	mpp := &pb_oip.MultiPart{}
	err := proto.Unmarshal(msg.SerializedMessage, mpp)
	if err != nil {
		log.Error("unable to unmarshal multipart", logger.Attrs{"txid": tx.Transaction.Txid, "err": err})
		return
	}

	if mpp.CountParts == 0 {
		log.Error("multipart count == 0", logger.Attrs{"txid": tx.Transaction.Txid})
		return
	}

	ms.Part = mpp.CurrentPart
	ms.Max = mpp.CountParts - 1

	ms.Data = string(mpp.RawData)
	ms.Reference = pb_oip.TxidPrefixToString(mpp.Reference)

	ms.Address = string(msg.PubKey)
	ms.Signature = base64.StdEncoding.EncodeToString(msg.Signature)

	if ms.Part == 0 {
		if len(tx.Transaction.Txid) > 16 {
			ms.Reference = tx.Transaction.Txid[0:16]
		} else {
			ms.Reference = tx.Transaction.Txid
		}
	}

	ms.Meta = MSMeta{
		Block:     tx.Block,
		BlockHash: tx.BlockHash,
		Complete:  false,
		Time:      tx.Transaction.Time,
		Tx:        tx,
		Txid:      tx.Transaction.Txid,
	}

	bir := elastic.NewBulkIndexRequest().Index(datastore.Index(multipartIndex)).Type("_doc").Doc(ms).Id(tx.Transaction.Txid)
	datastore.AutoBulk.Add(bir)
}

type MultipartSingle struct {
	Part      uint32 `json:"part"`
	Max       uint32 `json:"max"`
	Reference string `json:"reference"`
	Address   string `json:"address"`
	Signature string `json:"signature"`
	Data      string `json:"data"`
	Meta      MSMeta `json:"meta"`
}

type MSMeta struct {
	Block     int64                      `json:"block"`
	BlockHash string                     `json:"block_hash"`
	Complete  bool                       `json:"complete"`
	Stale     bool                       `json:"stale"`
	Time      int64                      `json:"time"`
	Tx        *datastore.TransactionData `json:"tx"`
	Txid      string                     `json:"txid"`
}

type Multipart struct {
	Parts []MultipartSingle
	Count uint32
	Total uint32
}
