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

		limit := 120
		if l := r.URL.Query().Get("limit"); l != "" {
			if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 1440 {
				limit = v
			}
		}

		payload, _ := json.Marshal(map[string]any{
			"node_id": nodeID,
			"type":    typ,
			"limit":   limit,
		})

		msg, err := nc.Request("orimono.loom.history", payload, timeout)
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
