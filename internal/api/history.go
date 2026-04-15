package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
)

type rpcError struct {
	Error string `json:"error"`
}

func HistoryHandler(nc *nats.Conn, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodeID := r.URL.Query().Get("node_id")
		typ := r.URL.Query().Get("type")
		if nodeID == "" || typ == "" {
			http.Error(w, "node_id and type required", http.StatusBadRequest)
			return
		}

		fromTs, err := strconv.ParseInt(r.URL.Query().Get("from_ts"), 10, 64)
		if err != nil || fromTs <= 0 {
			http.Error(w, "valid from_ts required (Unix seconds)", http.StatusBadRequest)
			return
		}
		toTs, err := strconv.ParseInt(r.URL.Query().Get("to_ts"), 10, 64)
		if err != nil || toTs <= fromTs {
			http.Error(w, "valid to_ts required (Unix seconds, must be > from_ts)", http.StatusBadRequest)
			return
		}

		maxPoints := 300
		if m := r.URL.Query().Get("max_points"); m != "" {
			if v, err := strconv.Atoi(m); err == nil && v > 0 && v <= 1000 {
				maxPoints = v
			}
		}

		payload, _ := json.Marshal(map[string]any{
			"node_id":    nodeID,
			"type":       typ,
			"from_ts":    fromTs,
			"to_ts":      toTs,
			"max_points": maxPoints,
		})

		msg, err := nc.Request("orimono.tsumu.history", payload, timeout)
		if err != nil {
			http.Error(w, "upstream unavailable", http.StatusServiceUnavailable)
			return
		}

		var rpcErr rpcError
		if json.Unmarshal(msg.Data, &rpcErr) == nil && rpcErr.Error != "" {
			http.Error(w, rpcErr.Error, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(msg.Data)
	}
}
