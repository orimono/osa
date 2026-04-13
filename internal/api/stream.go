package api

import (
	"fmt"
	"net/http"

	"github.com/nats-io/nats.go"
)

func StreamHandler(nc *nats.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodeID := r.URL.Query().Get("node_id")
		typ := r.URL.Query().Get("type")
		if nodeID == "" || typ == "" {
			http.Error(w, "node_id and type required", http.StatusBadRequest)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		subject := "orimono.live." + nodeID + "." + typ
		sub, err := nc.SubscribeSync(subject)
		if err != nil {
			http.Error(w, "failed to subscribe", http.StatusInternalServerError)
			return
		}
		defer sub.Unsubscribe()

		for {
			select {
			case <-r.Context().Done():
				return
			default:
				msg, err := sub.NextMsgWithContext(r.Context())
				if err != nil {
					return
				}
				fmt.Fprintf(w, "data: %s\n\n", msg.Data)
				flusher.Flush()
			}
		}
	}
}
