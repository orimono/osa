package api

import (
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
)

func NodesHandler(nc *nats.Conn, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		msg, err := nc.Request("orimono.loom.nodes.list", nil, timeout)
		if err != nil {
			http.Error(w, "upstream unavailable", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(msg.Data)
	}
}
