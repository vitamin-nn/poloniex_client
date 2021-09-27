package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const askCode = 0
const bidCode = 1
const askText = "buy"
const bidText = "sell"

var pairCodeDict =  map[int]string{
	121: "USDT_BTC",
	265: "USDT_TRX",
	149: "USDT_ETH",
}

func getPolonexPairList(poloniexParListJson []byte) ([]string, error) {
	pairList, err := parseInputPairList(poloniexParListJson, "poloniex")
	if err != nil {
		return nil, err
	}

	pairListReq := make([]string, len(pairList))
	var rPair string
	for i, pair := range pairList {
		rPair, err = getRevertedPair(pair)
		if err != nil {
			return nil, err
		}
		pairListReq[i] = rPair
	}

	return pairListReq, nil
}

func parseInputPairList(pairJsonList []byte, pairKey string) ([]string, error) {
	var pairListParsed map[string][]string
	json.Unmarshal(pairJsonList, &pairListParsed)
	pairList, ok := pairListParsed[pairKey]
	if !ok {
		return nil, errors.New(fmt.Sprintf("unknown pair key: %s", pairKey))
	}

	return pairList, nil
}

func getRevertedPair(pair string) (string, error) {
	cList := strings.Split(pair, "_")
	if len(cList) != 2 {
		return "", errors.New("unknown pair")
	}

	return strings.Join([]string{cList[1], cList[0]}, "_"), nil
}

func getSideText(sideCode float64) string {
	if (sideCode == askCode) {
		return "buy"
	} else {
		return "sell"
	}
}

func getPairByCode(pairCode int) (string, error) {
	pair, ok := pairCodeDict[pairCode]
	if !ok {
		return "", errors.New("unknown pair code")
	}

	return pair, nil
}

func getSubscribeCommand(pair string) ([]byte, error) {
	// {"command": "subscribe", "channel": "BTC_ETH"}
	type command struct{
		Command string `json:"command"`
		Channel string `json:"channel"`
	}
	cmd := command{
		Command: "subscribe",
		Channel: pair,
	}

	return json.Marshal(cmd)
}
