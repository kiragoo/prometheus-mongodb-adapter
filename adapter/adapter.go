package adapter

import (
	"crypto/tls"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/prometheus/prompb"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type timeSeries struct {
	Labels  []*label  `bson:"labels,omitempty"`
	Samples []*sample `bson:"samples,omitempty"`
}

type label struct {
	Name  string `bson:"name,omitempty"`
	Value string `bson:"value,omitempty"`
}

type sample struct {
	Timestamp int64   `bson:"timestamp"`
	Value     float64 `bson:"value"`
}

// MongoDBAdapter is an implemantation of prometheus remote stprage adapter for MongoDB
type MongoDBAdapter struct {
	session *mgo.Session
	c       *mgo.Collection
}

// New provides a MongoDBAdapter after initialization
func New(urlString, database, collection string) (*MongoDBAdapter, error) {

	u, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("url parse error: %s", err.Error())
	}
	query := u.Query()
	u.RawQuery = ""

	// DialInfo
	dialInfo, err := mgo.ParseURL(u.String())
	if err != nil {
		return nil, fmt.Errorf("mongo url parse error: %s", err.Error())
	}

	// SSL
	if strings.ToLower(query.Get("ssl")) == "true" {
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", addr.String(), &tls.Config{})
		}
	}
	//dialInfo.Timeout, _ = time.ParseDuration("10s")
	// Dial
	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return nil, fmt.Errorf("dial error: %s", err.Error())
	}

	// Database
	if dialInfo.Database == "" {
		dialInfo.Database = database
	}
	c := session.DB(dialInfo.Database).C(collection)

	return &MongoDBAdapter{
		session: session,
		c:       c,
	}, nil
}

// Close closes the connection with MongoDB
func (p *MongoDBAdapter) Close() {

	p.session.Close()
}

// Run serves with http listener
func (p *MongoDBAdapter) Run(address string) error {

	router := httprouter.New()
	router.POST("/write", p.handleWriteRequest)
	router.POST("/read", p.handleReadRequest)
	return http.ListenAndServe(address, handlers.RecoveryHandler()(handlers.LoggingHandler(os.Stdout, router)))
}

func (p *MongoDBAdapter) handleWriteRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !p.handleAuthRequest(w, r, params) {
		return
	}
	compressed, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reqBuf, err := snappy.Decode(nil, compressed)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req prompb.WriteRequest
	if err := proto.Unmarshal(reqBuf, &req); err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, ts := range req.Timeseries {
		mongoTS := &timeSeries{
			Labels:  []*label{},
			Samples: []*sample{},
		}
		for _, l := range ts.Labels {
			mongoTS.Labels = append(mongoTS.Labels, &label{
				Name:  l.Name,
				Value: l.Value,
			})
		}
		for _, s := range ts.Samples {
			mongoTS.Samples = append(mongoTS.Samples, &sample{
				Timestamp: s.Timestamp,
				Value:     s.Value,
			})
		}
		if err := p.c.Insert(mongoTS); err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func (p *MongoDBAdapter) handleReadRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !p.handleAuthRequest(w, r, params) {
		return
	}
	compressed, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reqBuf, err := snappy.Decode(nil, compressed)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req prompb.ReadRequest
	if err := proto.Unmarshal(reqBuf, &req); err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results := []*prompb.QueryResult{}
	for _, q := range req.Queries {

		query := map[string]interface{}{
			"samples": map[string]interface{}{
				"$elemMatch": map[string]interface{}{
					"timestamp": map[string]interface{}{
						"$gte": q.StartTimestampMs,
						"$lte": q.EndTimestampMs,
					},
				},
			},
		}
		if q.Matchers != nil && len(q.Matchers) > 0 {
			matcher := []map[string]interface{}{}
			for _, m := range q.Matchers {
				switch m.Type {
				case prompb.LabelMatcher_EQ:
					matcher = append(matcher, map[string]interface{}{
						"$elemMatch": map[string]interface{}{
							m.Name: m.Value,
						},
					})
				case prompb.LabelMatcher_NEQ:
					matcher = append(matcher, map[string]interface{}{
						"$elemMatch": map[string]interface{}{
							m.Name: map[string]interface{}{
								"$ne": m.Value,
							},
						},
					})
				case prompb.LabelMatcher_RE:
					matcher = append(matcher, map[string]interface{}{
						"$elemMatch": map[string]interface{}{
							m.Name: map[string]interface{}{
								"$regex": m.Value,
							},
						},
					})
				case prompb.LabelMatcher_NRE:
					matcher = append(matcher, map[string]interface{}{
						"$elemMatch": map[string]interface{}{
							m.Name: map[string]interface{}{
								"$not": map[string]interface{}{
									"$regex": m.Value,
								},
							},
						},
					})
				}
			}
			query["labels"] = map[string]interface{}{
				"$all": matcher,
			}
		}

		iter := p.c.Find(query).Sort("samples.timestamp").Iter()
		defer iter.Close()

		timeseries := []*prompb.TimeSeries{}
		var ts timeSeries
		for iter.Next(&ts) {
			timeseries = append(timeseries, &prompb.TimeSeries{})
		}

		results = append(results, &prompb.QueryResult{
			Timeseries: timeseries,
		})
	}
	resp := &prompb.ReadResponse{
		Results: results,
	}
	data, err := proto.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-protobuf")
	w.Header().Set("Content-Encoding", "snappy")
	compressed = snappy.Encode(nil, data)
	if _, err := w.Write(compressed); err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
