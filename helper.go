package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const askCode = 0
const askText = "buy"
const bidText = "sell"

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
	err := json.Unmarshal(pairJsonList, &pairListParsed)
	if err != nil {
		return nil, err
	}
	pairList, ok := pairListParsed[pairKey]
	if !ok {
		return nil, fmt.Errorf("unknown pair key: %s", pairKey)
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
		return askText
	} else {
		return bidText
	}
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
