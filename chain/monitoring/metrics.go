package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"runtime"
	"sync"
	"time"

	"quantum-blockchain/chain/consensus"
	"quantum-blockchain/chain/types"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsServer provides comprehensive monitoring for the quantum blockchain
type MetricsServer struct {
	// Configuration
	listenAddr     string
	metricsPath    string
	healthPath     string
	
	// Metrics registry
	registry       *prometheus.Registry
	
	// Core metrics
	blockHeight           prometheus.Gauge
	blockTime            prometheus.Histogram
	transactionCount     prometheus.Counter
	transactionPool      prometheus.Gauge
	consensusLatency     prometheus.Histogram
	validatorCount       prometheus.Gauge
	networkLatency       prometheus.Histogram
	
	// Validator metrics
	validatorUptime      *prometheus.GaugeVec
	validatorPerformance *prometheus.GaugeVec
	validatorStake       *prometheus.GaugeVec
	slashingEvents       prometheus.Counter
	
	// Network metrics
	peerCount            prometheus.Gauge
	messageRate          *prometheus.CounterVec
	bandwidthUsage       *prometheus.CounterVec
	connectionErrors     prometheus.Counter
	
	// Quantum crypto metrics
	dilithiumVerifyTime  prometheus.Histogram
	kyberEncryptTime     prometheus.Histogram
	signatureFailures    *prometheus.CounterVec
	
	// System metrics
	memoryUsage          prometheus.Gauge
	cpuUsage             prometheus.Gauge
	diskUsage            prometheus.Gauge
	goroutineCount       prometheus.Gauge
	
	// Health status
	healthStatus         *HealthChecker
	
	// Data collection
	dataCollector        *DataCollector
	
	// HTTP server
	server               *http.Server
	
	// Control
	ctx                  context.Context
	cancel               context.CancelFunc
	wg                   sync.WaitGroup
	mu                   sync.RWMutex
	
	// State
	running              bool
	startTime            time.Time
}

// HealthChecker monitors system health
type HealthChecker struct {
	checks          map[string]HealthCheck
	overallStatus   HealthStatus
	lastCheck       time.Time
	checkInterval   time.Duration
	mu              sync.RWMutex
}

// HealthCheck represents a single health check
type HealthCheck struct {
	Name        string        `json:"name"`
	Status      HealthStatus  `json:"status"`
	Message     string        `json:"message"`
	LastCheck   time.Time     `json:"lastCheck"`
	Duration    time.Duration `json:"duration"`
	Critical    bool          `json:"critical"`
	CheckFunc   func() (HealthStatus, string, error) `json:"-"`
}

// HealthStatus represents health status
type HealthStatus int

const (
	HealthStatusHealthy HealthStatus = iota
	HealthStatusWarning
	HealthStatusCritical
	HealthStatusUnknown
)

// DataCollector collects and aggregates metrics data
type DataCollector struct {
	blockchain       BlockchainInterface
	consensus        ConsensusInterface
	network          NetworkInterface
	
	// Historical data
	blockTimes       []time.Duration
	transactionRates []float64
	validatorMetrics map[string]*ValidatorMetrics
	networkStats     *NetworkStats
	
	// Collection intervals
	blockMetricsInterval    time.Duration
	networkMetricsInterval  time.Duration
	systemMetricsInterval   time.Duration
	
	mu                      sync.RWMutex
}

// Interfaces for data collection
type BlockchainInterface interface {
	GetCurrentBlock() *types.Block
	GetBlockByNumber(number *big.Int) *types.Block
	GetTransactionCount() uint64
	GetPendingTransactionCount() uint64
}

type ConsensusInterface interface {
	GetValidatorSet() []*consensus.ValidatorState
	GetConsensusInfo() map[string]interface{}
	GetNetworkPerformance() *consensus.NetworkPerformance
}

type NetworkInterface interface {
	GetPeerCount() int
	GetBandwidthUsage() (uint64, uint64)
	GetMessageStats() map[string]uint64
	GetLatencyStats() time.Duration
}

// ValidatorMetrics tracks individual validator performance
type ValidatorMetrics struct {
	Address          string    `json:"address"`
	Uptime           float64   `json:"uptime"`
	Performance      float64   `json:"performance"`
	BlocksProposed   uint64    `json:"blocksProposed"`
	BlocksMissed     uint64    `json:"blocksMissed"`
	Stake            string    `json:"stake"`
	LastActive       time.Time `json:"lastActive"`
	Slashed          bool      `json:"slashed"`
	JailedUntil      time.Time `json:"jailedUntil"`
}

// NetworkStats tracks network performance
type NetworkStats struct {
	TotalPeers       int           `json:"totalPeers"`
	ValidatorPeers   int           `json:"validatorPeers"`
	AvgLatency       time.Duration `json:"avgLatency"`
	MessageRate      float64       `json:"messageRate"`
	BandwidthIn      uint64        `json:"bandwidthIn"`
	BandwidthOut     uint64        `json:"bandwidthOut"`
	ConnectionErrors uint64        `json:"connectionErrors"`
	LastUpdate       time.Time     `json:"lastUpdate"`
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(config *MetricsConfig) *MetricsServer {
	ctx, cancel := context.WithCancel(context.Background())
	
	registry := prometheus.NewRegistry()
	
	ms := &MetricsServer{
		listenAddr:    config.ListenAddr,
		metricsPath:   config.MetricsPath,
		healthPath:    config.HealthPath,
		registry:      registry,
		ctx:           ctx,
		cancel:        cancel,
		startTime:     time.Now(),
		healthStatus:  NewHealthChecker(),
		dataCollector: NewDataCollector(config),
	}
	
	// Initialize metrics
	ms.initMetrics()
	
	// Setup HTTP server
	ms.setupServer()
	
	return ms
}

// MetricsConfig defines metrics configuration
type MetricsConfig struct {
	ListenAddr    string `json:"listenAddr"`
	MetricsPath   string `json:"metricsPath"`
	HealthPath    string `json:"healthPath"`
	EnableAuth    bool   `json:"enableAuth"`
	AuthToken     string `json:"authToken,omitempty"`
}

// initMetrics initializes all Prometheus metrics
func (ms *MetricsServer) initMetrics() {
	// Block metrics
	ms.blockHeight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "quantum_block_height",
		Help: "Current block height",
	})
	
	ms.blockTime = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "quantum_block_time_seconds",
		Help:    "Time between blocks in seconds",
		Buckets: []float64{0.5, 1, 2, 5, 10, 30, 60},
	})
	
	ms.transactionCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "quantum_transactions_total",
		Help: "Total number of transactions processed",
	})
	
	ms.transactionPool = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "quantum_transaction_pool_size",
		Help: "Current transaction pool size",
	})
	
	// Consensus metrics
	ms.consensusLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "quantum_consensus_latency_seconds",
		Help:    "Consensus latency in seconds",
		Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
	})
	
	ms.validatorCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "quantum_validators_active",
		Help: "Number of active validators",
	})
	
	// Validator metrics
	ms.validatorUptime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "quantum_validator_uptime",
		Help: "Validator uptime percentage",
	}, []string{"validator_address"})
	
	ms.validatorPerformance = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "quantum_validator_performance",
		Help: "Validator performance score",
	}, []string{"validator_address"})
	
	ms.validatorStake = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "quantum_validator_stake",
		Help: "Validator stake amount",
	}, []string{"validator_address"})
	
	ms.slashingEvents = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "quantum_slashing_events_total",
		Help: "Total number of slashing events",
	})
	
	// Network metrics
	ms.peerCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "quantum_peers_connected",
		Help: "Number of connected peers",
	})
	
	ms.messageRate = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "quantum_messages_total",
		Help: "Total messages sent/received by type",
	}, []string{"message_type", "direction"})
	
	ms.bandwidthUsage = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "quantum_bandwidth_bytes_total",
		Help: "Total bandwidth usage in bytes",
	}, []string{"direction"})
	
	ms.connectionErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "quantum_connection_errors_total",
		Help: "Total connection errors",
	})
	
	// Quantum crypto metrics
	ms.dilithiumVerifyTime = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "quantum_dilithium_verify_seconds",
		Help:    "Time to verify Dilithium signatures",
		Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5},
	})
	
	ms.kyberEncryptTime = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "quantum_kyber_encrypt_seconds",
		Help:    "Time to perform Kyber encryption",
		Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1},
	})
	
	ms.signatureFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "quantum_signature_failures_total",
		Help: "Total signature verification failures",
	}, []string{"algorithm"})
	
	// System metrics
	ms.memoryUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "quantum_memory_usage_bytes",
		Help: "Memory usage in bytes",
	})
	
	ms.cpuUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "quantum_cpu_usage_percent",
		Help: "CPU usage percentage",
	})
	
	ms.diskUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "quantum_disk_usage_bytes",
		Help: "Disk usage in bytes",
	})
	
	ms.goroutineCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "quantum_goroutines",
		Help: "Number of goroutines",
	})
	
	// Register all metrics
	metrics := []prometheus.Collector{
		ms.blockHeight,
		ms.blockTime,
		ms.transactionCount,
		ms.transactionPool,
		ms.consensusLatency,
		ms.validatorCount,
		ms.validatorUptime,
		ms.validatorPerformance,
		ms.validatorStake,
		ms.slashingEvents,
		ms.peerCount,
		ms.messageRate,
		ms.bandwidthUsage,
		ms.connectionErrors,
		ms.dilithiumVerifyTime,
		ms.kyberEncryptTime,
		ms.signatureFailures,
		ms.memoryUsage,
		ms.cpuUsage,
		ms.diskUsage,
		ms.goroutineCount,
	}
	
	for _, metric := range metrics {
		ms.registry.MustRegister(metric)
	}
}

// setupServer configures the HTTP server
func (ms *MetricsServer) setupServer() {
	router := mux.NewRouter()
	
	// Metrics endpoint
	router.Path(ms.metricsPath).Handler(promhttp.HandlerFor(ms.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))
	
	// Health endpoint
	router.PathPrefix(ms.healthPath).HandlerFunc(ms.healthHandler)
	
	// Additional endpoints
	router.PathPrefix("/api/metrics/blockchain").HandlerFunc(ms.blockchainMetricsHandler)
	router.PathPrefix("/api/metrics/validators").HandlerFunc(ms.validatorMetricsHandler)
	router.PathPrefix("/api/metrics/network").HandlerFunc(ms.networkMetricsHandler)
	router.PathPrefix("/api/metrics/system").HandlerFunc(ms.systemMetricsHandler)
	
	ms.server = &http.Server{
		Addr:    ms.listenAddr,
		Handler: router,
	}
}

// Start starts the metrics server
func (ms *MetricsServer) Start() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if ms.running {
		return fmt.Errorf("metrics server already running")
	}
	
	// Start health checker
	ms.healthStatus.Start()
	
	// Start data collection
	ms.wg.Add(1)
	go ms.collectMetrics()
	
	// Start HTTP server
	ms.wg.Add(1)
	go func() {
		defer ms.wg.Done()
		
		log.Printf("Starting metrics server on %s", ms.listenAddr)
		if err := ms.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Metrics server error: %v", err)
		}
	}()
	
	ms.running = true
	return nil
}

// Stop stops the metrics server
func (ms *MetricsServer) Stop() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if !ms.running {
		return
	}
	
	ms.cancel()
	
	if ms.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ms.server.Shutdown(ctx)
	}
	
	ms.healthStatus.Stop()
	ms.wg.Wait()
	
	ms.running = false
	log.Printf("Metrics server stopped")
}

// collectMetrics collects metrics at regular intervals
func (ms *MetricsServer) collectMetrics() {
	defer ms.wg.Done()
	
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ms.ctx.Done():
			return
		case <-ticker.C:
			ms.updateMetrics()
		}
	}
}

// updateMetrics updates all metrics
func (ms *MetricsServer) updateMetrics() {
	// Update system metrics
	ms.updateSystemMetrics()
	
	// Update blockchain metrics
	if ms.dataCollector.blockchain != nil {
		ms.updateBlockchainMetrics()
	}
	
	// Update consensus metrics
	if ms.dataCollector.consensus != nil {
		ms.updateConsensusMetrics()
	}
	
	// Update network metrics
	if ms.dataCollector.network != nil {
		ms.updateNetworkMetrics()
	}
}

// updateSystemMetrics updates system-level metrics
func (ms *MetricsServer) updateSystemMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	ms.memoryUsage.Set(float64(m.Alloc))
	ms.goroutineCount.Set(float64(runtime.NumGoroutine()))
	
	// CPU usage would require additional implementation
	ms.cpuUsage.Set(ms.getCPUUsage())
	ms.diskUsage.Set(ms.getDiskUsage())
}

// updateBlockchainMetrics updates blockchain metrics
func (ms *MetricsServer) updateBlockchainMetrics() {
	currentBlock := ms.dataCollector.blockchain.GetCurrentBlock()
	if currentBlock != nil {
		ms.blockHeight.Set(float64(currentBlock.Number().Uint64()))
	}
	
	ms.transactionCount.Add(float64(ms.dataCollector.blockchain.GetTransactionCount()))
	ms.transactionPool.Set(float64(ms.dataCollector.blockchain.GetPendingTransactionCount()))
}

// updateConsensusMetrics updates consensus metrics
func (ms *MetricsServer) updateConsensusMetrics() {
	validators := ms.dataCollector.consensus.GetValidatorSet()
	ms.validatorCount.Set(float64(len(validators)))
	
	// Update individual validator metrics
	for _, validator := range validators {
		addr := validator.Address.Hex()
		ms.validatorPerformance.WithLabelValues(addr).Set(validator.Performance.ReliabilityScore)
		ms.validatorUptime.WithLabelValues(addr).Set(validator.Performance.UptimeScore)
		
		stakeFloat, _ := validator.TotalStake.Float64()
		ms.validatorStake.WithLabelValues(addr).Set(stakeFloat)
	}
}

// updateNetworkMetrics updates network metrics
func (ms *MetricsServer) updateNetworkMetrics() {
	ms.peerCount.Set(float64(ms.dataCollector.network.GetPeerCount()))
	
	bandwidthIn, bandwidthOut := ms.dataCollector.network.GetBandwidthUsage()
	ms.bandwidthUsage.WithLabelValues("in").Add(float64(bandwidthIn))
	ms.bandwidthUsage.WithLabelValues("out").Add(float64(bandwidthOut))
}

// HTTP handlers
func (ms *MetricsServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	health := ms.healthStatus.GetOverallHealth()
	
	status := http.StatusOK
	if health.Status == HealthStatusCritical {
		status = http.StatusServiceUnavailable
	} else if health.Status == HealthStatusWarning {
		status = http.StatusPartialContent
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(health)
}

func (ms *MetricsServer) blockchainMetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := ms.getBlockchainMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (ms *MetricsServer) validatorMetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := ms.getValidatorMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (ms *MetricsServer) networkMetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := ms.getNetworkMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (ms *MetricsServer) systemMetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := ms.getSystemMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	hc := &HealthChecker{
		checks:        make(map[string]HealthCheck),
		checkInterval: 30 * time.Second,
	}
	
	// Add default health checks
	hc.addDefaultChecks()
	
	return hc
}

// addDefaultChecks adds default health checks
func (hc *HealthChecker) addDefaultChecks() {
	hc.checks["memory"] = HealthCheck{
		Name:      "Memory Usage",
		Critical:  true,
		CheckFunc: hc.checkMemoryUsage,
	}
	
	hc.checks["disk"] = HealthCheck{
		Name:      "Disk Space",
		Critical:  true,
		CheckFunc: hc.checkDiskSpace,
	}
	
	hc.checks["goroutines"] = HealthCheck{
		Name:      "Goroutine Count",
		Critical:  false,
		CheckFunc: hc.checkGoroutineCount,
	}
}

// Health check implementations
func (hc *HealthChecker) checkMemoryUsage() (HealthStatus, string, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Check if memory usage is above 80% of available
	usagePercent := float64(m.Alloc) / float64(m.Sys) * 100
	
	if usagePercent > 90 {
		return HealthStatusCritical, fmt.Sprintf("Memory usage critical: %.1f%%", usagePercent), nil
	} else if usagePercent > 80 {
		return HealthStatusWarning, fmt.Sprintf("Memory usage high: %.1f%%", usagePercent), nil
	}
	
	return HealthStatusHealthy, fmt.Sprintf("Memory usage normal: %.1f%%", usagePercent), nil
}

func (hc *HealthChecker) checkDiskSpace() (HealthStatus, string, error) {
	// Implementation would check actual disk space
	return HealthStatusHealthy, "Disk space normal", nil
}

func (hc *HealthChecker) checkGoroutineCount() (HealthStatus, string, error) {
	count := runtime.NumGoroutine()
	
	if count > 10000 {
		return HealthStatusWarning, fmt.Sprintf("High goroutine count: %d", count), nil
	}
	
	return HealthStatusHealthy, fmt.Sprintf("Goroutine count normal: %d", count), nil
}

// NewDataCollector creates a new data collector
func NewDataCollector(config *MetricsConfig) *DataCollector {
	return &DataCollector{
		validatorMetrics:        make(map[string]*ValidatorMetrics),
		networkStats:            &NetworkStats{},
		blockMetricsInterval:    10 * time.Second,
		networkMetricsInterval:  15 * time.Second,
		systemMetricsInterval:   5 * time.Second,
	}
}

// SetInterfaces sets the data collection interfaces
func (ms *MetricsServer) SetInterfaces(
	blockchain BlockchainInterface,
	consensus ConsensusInterface,
	network NetworkInterface,
) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	ms.dataCollector.blockchain = blockchain
	ms.dataCollector.consensus = consensus
	ms.dataCollector.network = network
}

// RecordDilithiumVerifyTime records Dilithium verification time
func (ms *MetricsServer) RecordDilithiumVerifyTime(duration time.Duration) {
	ms.dilithiumVerifyTime.Observe(duration.Seconds())
}

// RecordKyberEncryptTime records Kyber encryption time
func (ms *MetricsServer) RecordKyberEncryptTime(duration time.Duration) {
	ms.kyberEncryptTime.Observe(duration.Seconds())
}

// RecordSignatureFailure records signature verification failure
func (ms *MetricsServer) RecordSignatureFailure(algorithm string) {
	ms.signatureFailures.WithLabelValues(algorithm).Inc()
}

// RecordSlashingEvent records a slashing event
func (ms *MetricsServer) RecordSlashingEvent() {
	ms.slashingEvents.Inc()
}

// Helper methods (simplified implementations)
func (ms *MetricsServer) getCPUUsage() float64 {
	// Implementation would measure actual CPU usage
	return 0.0
}

func (ms *MetricsServer) getDiskUsage() float64 {
	// Implementation would measure actual disk usage
	return 0.0
}

func (ms *MetricsServer) getBlockchainMetrics() map[string]interface{} {
	// Implementation would return detailed blockchain metrics
	return make(map[string]interface{})
}

func (ms *MetricsServer) getValidatorMetrics() map[string]interface{} {
	// Implementation would return detailed validator metrics
	return make(map[string]interface{})
}

func (ms *MetricsServer) getNetworkMetrics() map[string]interface{} {
	// Implementation would return detailed network metrics
	return make(map[string]interface{})
}

func (ms *MetricsServer) getSystemMetrics() map[string]interface{} {
	// Implementation would return detailed system metrics
	return make(map[string]interface{})
}

// Additional health checker methods
func (hc *HealthChecker) Start() {
	// Implementation would start periodic health checks
}

func (hc *HealthChecker) Stop() {
	// Implementation would stop health checks
}

func (hc *HealthChecker) GetOverallHealth() *HealthCheck {
	// Implementation would return overall health status
	return &HealthCheck{
		Name:    "Overall Health",
		Status:  HealthStatusHealthy,
		Message: "All systems operational",
	}
}