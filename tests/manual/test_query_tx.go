package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type JSONRPCRequest_query struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

type JSONRPCResponse_query struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	ID      int         `json:"id"`
}

func runTestQueryTx() {
	// Query the transaction we just submitted
	txHash := "0xb20fb16a788e9a5e6aeb70cb6c45ad375f46a8a393d537711e64f43b896e01ba"
	
	req := JSONRPCRequest_query{
		JSONRPC: "2.0",
		Method:  "eth_getTransactionByHash",
		Params:  []string{txHash},
		ID:      1,
	}

	reqData, err := json.Marshal(req)
	if err != nil {
		log.Fatal("Request marshal error:", err)
	}

	fmt.Printf("Querying transaction %s...\n", txHash)
	resp, err := http.Post("http://localhost:8545", "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		log.Fatal("HTTP request error:", err)
	}
	defer resp.Body.Close()

	var rpcResp JSONRPCResponse_query
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	if err != nil {
		log.Fatal("Response decode error:", err)
	}

	if rpcResp.Error != nil {
		fmt.Printf("RPC Error: %v\n", rpcResp.Error)
	} else {
		fmt.Printf("Transaction found!\n")
		resultJSON, _ := json.MarshalIndent(rpcResp.Result, "", "  ")
		fmt.Printf("Result: %s\n", resultJSON)
	}
}