package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("🚀 Quantum Blockchain Performance Test Suite")
		fmt.Println("===========================================")
		fmt.Println("\nUsage: go run . [test-name]")
		fmt.Println("\nAvailable performance tests:")
		fmt.Println("  fast                - Fast performance test")
		fmt.Println("  live                - Live blockchain test")
		fmt.Println("\nExample: go run . fast")
		return
	}

	switch os.Args[1] {
	case "fast":
		fmt.Println("Running: Fast Performance Test")
		runFastPerformanceTest()
	case "live":
		fmt.Println("Running: Live Blockchain Test")
		runLiveBlockchainTest()
	default:
		fmt.Printf("❌ Unknown test: %s\n", os.Args[1])
		fmt.Println("Run without arguments to see available tests")
	}
}