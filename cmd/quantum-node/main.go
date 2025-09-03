package main

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	"quantum-blockchain/chain/node"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "quantum-node",
	Short: "Quantum-resistant blockchain node",
	Long:  "A quantum-resistant blockchain node with EVM compatibility",
	Run:   runNode,
}

var (
	configFile    string
	port          int
	rpcPort       int
	dataDir       string
	genesisConfig string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file")
	rootCmd.PersistentFlags().IntVar(&port, "port", 30303, "P2P network port")
	rootCmd.PersistentFlags().IntVar(&rpcPort, "rpc-port", 8545, "JSON-RPC server port")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", "./data", "data directory")
	rootCmd.PersistentFlags().StringVar(&genesisConfig, "genesis", "./config/genesis.json", "genesis configuration file")

	viper.BindPFlags(rootCmd.PersistentFlags())
}

func runNode(cmd *cobra.Command, args []string) {
	fmt.Printf("🚀 Starting Quantum Blockchain Node v%s\n", Version)
	fmt.Printf("📊 Build: %s (commit: %s)\n", BuildTime, Commit)
	
	config := &node.Config{
		DataDir:       dataDir,
		NetworkID:     8888,
		ListenAddr:    fmt.Sprintf(":%d", port),
		HTTPPort:      rpcPort,
		WSPort:        rpcPort + 1,
		ValidatorKey:  "auto",  // Enable validator mode (key will be auto-generated)
		ValidatorAlg:  "dilithium",
		GenesisConfig: genesisConfig,
		Mining:        true,
		GasLimit:      15000000,
		GasPrice:      big.NewInt(1000000000), // 1 Gwei
	}

	// Create and start the node
	quantumNode, err := node.NewNode(config)
	if err != nil {
		log.Printf("❌ Failed to create node: %v", err)
		os.Exit(1)
	}
	
	// Start the node in a goroutine
	go func() {
		if err := quantumNode.Start(); err != nil {
			log.Printf("❌ Node failed to start: %v", err)
			os.Exit(1)
		}
	}()

	fmt.Printf("🌐 P2P listening on port %d\n", port)
	fmt.Printf("🔗 JSON-RPC server listening on port %d\n", rpcPort)
	fmt.Printf("💾 Data directory: %s\n", dataDir)
	fmt.Println("✅ Quantum blockchain is running!")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	<-c
	fmt.Println("\n🛑 Shutting down quantum blockchain node...")
	
	quantumNode.Stop()
	
	fmt.Println("👋 Quantum blockchain node stopped")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}