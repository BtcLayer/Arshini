package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/aarzilli/golua/lua"
	"github.com/fiatjaf/etleneum/data"
)

func contractFromRedis(ctid string) (ct *data.Contract, err error) {
	var jct []byte
	ct = &data.Contract{}

	jct, err = rds.Get("contract:" + ctid).Bytes()
	if err != nil {
		return
	}

	err = json.Unmarshal(jct, ct)
	if err != nil {
		return
	}

	return
}

func checkContractCode(code string) (ok bool) {
	if strings.Index(code, "function __init__") == -1 {
		return false
	}

	L := lua.NewState()
	defer L.Close()

	lerr := L.LoadString(code)
	if lerr != 0 {
		return false
	}

	return true
}

func getContractCost(ct data.Contract) int64 {
	words := int64(len(wordMatcher.FindAllString(ct.Code, -1)))
	return 1000*s.InitialContractCostSatoshis + 1000*words
}

func saveContractOnRedis(ct data.Contract) (jct []byte, err error) {
	jct, err = json.Marshal(ct)
	if err != nil {
		return
	}

	err = rds.Set("contract:"+ct.Id, jct, time.Hour*20).Err()
	if err != nil {
		return
	}

	return
}
