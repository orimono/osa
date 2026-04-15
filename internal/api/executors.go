package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/orimono/ito"
)

func ExecutorsHandler(nc *nats.Conn, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodeID := r.PathValue("nodeId")
		if nodeID == "" {
			http.Error(w, "nodeId required", http.StatusBadRequest)
			return
		}

		payload, _ := json.Marshal(map[string]string{
			"node_id": nodeID,
		})

		msg, err := nc.Request("orimono.loom.executors", payload, timeout)
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

func RegisterExecutorHandler(nc *nats.Conn, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodeID := r.PathValue("nodeId")
		if nodeID == "" {
			http.Error(w, "nodeId required", http.StatusBadRequest)
			return
		}

		var reg ito.ExecutorRegistration
		if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
			http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
			return
		}
		if reg.Kind == "" || reg.Runtime == "" || reg.Script == "" {
			http.Error(w, "kind, runtime and script are required", http.StatusBadRequest)
			return
		}

		payload, _ := json.Marshal(map[string]any{
			"node_id":  nodeID,
			"executor": reg,
		})

		msg, err := nc.Request("orimono.loom.executor.register", payload, timeout)
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
		w.WriteHeader(http.StatusCreated)
		w.Write(msg.Data)
	}
}
