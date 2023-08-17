package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/antage/eventsource.v1"
)

type ctevent struct {
	Id         string `json:"id"`
	ContractId string `json:"contract_id,omitempty"`
	Method     string `json:"method,omitempty"`
	Msatoshi   int64  `json:"msatoshi,omitempty"`
	Message    string `json:"message,omitempty"`
	Kind       string `json:"kind,omitempty"`
}

func dispatchContractEvent(contractId string, ev ctevent, typ string) {
	jpayload, _ := json.Marshal(ev)
	payload := string(jpayload)

	if ies, ok := contractstreams.Get(contractId); ok {
		ies.(eventsource.EventSource).SendEventMessage(payload, typ, "")
	}
}

func contractStream(w http.ResponseWriter, r *http.Request) {
	ctid := mux.Vars(r)["ctid"]

	var es eventsource.EventSource
	ies, ok := contractstreams.Get(ctid)

	if !ok {
		es = eventsource.New(
			&eventsource.Settings{
				Timeout:        5 * time.Second,
				CloseOnTimeout: true,
				IdleTimeout:    1 * time.Minute,
			},
			func(r *http.Request) [][]byte {
				return [][]byte{
					[]byte("X-Accel-Buffering: no"),
					[]byte("Cache-Control: no-cache"),
					[]byte("Content-Type: text/event-stream"),
					[]byte("Connection: keep-alive"),
					[]byte("Access-Control-Allow-Origin: *"),
				}
			},
		)
		go func() {
			for {
				time.Sleep(25 * time.Second)
				es.SendEventMessage("", "keepalive", "")
			}
		}()
		contractstreams.Set(ctid, es)
	} else {
		es = ies.(eventsource.EventSource)
	}

	go func() {
		time.Sleep(1 * time.Second)
		es.SendRetryMessage(3 * time.Second)
	}()

	es.ServeHTTP(w, r)
}

type callPrinter struct {
	ContractId string
	CallId     string
	Method     string
}

func (cp *callPrinter) Write(data []byte) (n int, err error) {
	dispatchContractEvent(cp.ContractId, ctevent{cp.CallId, cp.ContractId, cp.Method, 0, string(data), "print"}, "call-run-event")
	return len(data), nil
}
