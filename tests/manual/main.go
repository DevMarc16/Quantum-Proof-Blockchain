package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("🚀 Quantum Blockchain Test Suite")
		fmt.Println("================================")
		fmt.Println("\nUsage: go run . [test-name]")
		fmt.Println("\nAvailable tests:")
		fmt.Println("  deploy              - Deploy Quantum Token contract")
		fmt.Println("  contract            - Test contract deployment")
		fmt.Println("  transaction         - Test basic transaction")
		fmt.Println("  rpc-submit          - Test RPC transaction submission")
		fmt.Println("  query               - Query transaction by hash")
		fmt.Println("  balance             - Test simple balance check")
		fmt.Println("  funded              - Test funded transaction")
		fmt.Println("  funded-genesis      - Test funded genesis transaction")
		fmt.Println("  successful          - Test successful transaction")
		fmt.Println("\nExample: go run . deploy")
		return
	}

	switch os.Args[1] {
	case "deploy":
		fmt.Println("Running: Deploy Quantum Token")
		runDeployQuantumToken()
	case "contract":
		fmt.Println("Running: Test Contract Deployment")
		runTestContractDeployment()
	case "transaction":
		fmt.Println("Running: Test Transaction")
		runTestTransaction()
	case "rpc-submit":
		fmt.Println("Running: Test RPC Submit")
		runTestRPCSubmit()
	case "query":
		fmt.Println("Running: Test Query Transaction")
		runTestQueryTx()
	case "balance":
		fmt.Println("Running: Test Simple Balance")
		runTestSimpleBalance()
	case "funded":
		fmt.Println("Running: Test Funded Transaction")
		runTestFundedTx()
	case "funded-genesis":
		fmt.Println("Running: Test Funded Genesis Transaction")
		runTestFundedGenesisTx()
	case "successful":
		fmt.Println("Running: Test Successful Transaction")
		runTestSuccessfulTx()
	default:
		fmt.Printf("❌ Unknown test: %s\n", os.Args[1])
		fmt.Println("Run without arguments to see available tests")
	}
}