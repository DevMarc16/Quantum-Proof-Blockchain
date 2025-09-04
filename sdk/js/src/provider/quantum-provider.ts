/**
 * Quantum Blockchain Provider
 * Extends standard Web3 provider with quantum-resistant functionality
 */

import { EventEmitter } from 'events';
import axios, { AxiosInstance } from 'axios';
import WebSocket from 'ws';
import { 
  ProviderConfig, 
  RPCRequest, 
  RPCResponse, 
  QuantumTransaction,
  QuantumBlock,
  QuantumAccount,
  QuantumTransactionReceipt,
  QuantumMetrics,
  Web3QuantumProvider
} from '../types/quantum';

export class QuantumProvider extends EventEmitter implements Web3QuantumProvider {
  private config: ProviderConfig;
  private http: AxiosInstance;
  private ws?: WebSocket;
  private requestId = 0;
  private pendingRequests = new Map<number, { resolve: Function; reject: Function }>();

  public readonly isQuantumEnabled = true;
  public readonly chainId: string;

  constructor(config: ProviderConfig) {
    super();
    this.config = {
      timeout: 30000,
      retries: 3,
      ...config
    };
    this.chainId = `0x${config.chainId.toString(16)}`;

    // Setup HTTP client
    this.http = axios.create({
      baseURL: config.url,
      timeout: config.timeout,
      headers: {
        'Content-Type': 'application/json',
        ...config.headers
      }
    });

    // Setup WebSocket connection if available
    const wsUrl = config.url.replace(/^http/, 'ws') + '/ws';
    this.initializeWebSocket(wsUrl);
  }

  /**
   * Make RPC request to quantum blockchain
   */
  async request(args: { method: string; params?: any[] }): Promise<any> {
    const request: RPCRequest = {
      jsonrpc: '2.0',
      id: ++this.requestId,
      method: args.method,
      params: args.params || []
    };

    try {
      const response = await this.http.post('/', request);
      const rpcResponse: RPCResponse = response.data;

      if (rpcResponse.error) {
        throw new Error(`RPC Error ${rpcResponse.error.code}: ${rpcResponse.error.message}`);
      }

      return rpcResponse.result;
    } catch (error: any) {
      if (error.response?.data?.error) {
        throw new Error(`RPC Error: ${error.response.data.error.message}`);
      }
      throw error;
    }
  }

  /**
   * Get current block number
   */
  async getBlockNumber(): Promise<number> {
    const result = await this.request({ method: 'eth_blockNumber' });
    return parseInt(result, 16);
  }

  /**
   * Get quantum account information
   */
  async getAccount(address: string): Promise<QuantumAccount> {
    const [balance, nonce] = await Promise.all([
      this.request({ method: 'eth_getBalance', params: [address, 'latest'] }),
      this.request({ method: 'eth_getTransactionCount', params: [address, 'latest'] })
    ]);

    // Try to get quantum-specific account info
    let quantumInfo;
    try {
      quantumInfo = await this.request({ 
        method: 'quantum_getAccountInfo', 
        params: [address] 
      });
    } catch {
      // Fallback if quantum methods not available
      quantumInfo = {};
    }

    return {
      address,
      balance: balance,
      nonce: nonce,
      quantumPublicKey: quantumInfo.quantumPublicKey,
      signatureAlgorithm: quantumInfo.signatureAlgorithm
    };
  }

  /**
   * Get quantum block information
   */
  async getBlock(blockNumber: string | number): Promise<QuantumBlock> {
    const blockParam = typeof blockNumber === 'number' 
      ? `0x${blockNumber.toString(16)}` 
      : blockNumber;
    
    const block = await this.request({ 
      method: 'eth_getBlockByNumber', 
      params: [blockParam, true] 
    });

    return {
      number: block.number,
      hash: block.hash,
      parentHash: block.parentHash,
      timestamp: block.timestamp,
      validator: block.miner || block.validator,
      quantumSignature: block.quantumSignature || '',
      transactions: block.transactions.map(this.mapToQuantumTransaction),
      gasUsed: block.gasUsed,
      gasLimit: block.gasLimit
    };
  }

  /**
   * Send quantum transaction
   */
  async sendQuantumTransaction(tx: QuantumTransaction): Promise<string> {
    const result = await this.request({ 
      method: 'quantum_sendRawTransaction', 
      params: [this.serializeQuantumTransaction(tx)] 
    });
    
    this.emit('transactionSent', { hash: result, transaction: tx });
    return result;
  }

  /**
   * Get quantum transaction receipt
   */
  async getTransactionReceipt(txHash: string): Promise<QuantumTransactionReceipt | null> {
    const receipt = await this.request({ 
      method: 'eth_getTransactionReceipt', 
      params: [txHash] 
    });

    if (!receipt) return null;

    return {
      transactionHash: receipt.transactionHash,
      transactionIndex: receipt.transactionIndex,
      blockNumber: receipt.blockNumber,
      blockHash: receipt.blockHash,
      from: receipt.from,
      to: receipt.to,
      gasUsed: receipt.gasUsed,
      effectiveGasPrice: receipt.effectiveGasPrice,
      status: receipt.status,
      logs: receipt.logs,
      contractAddress: receipt.contractAddress,
      quantumVerified: receipt.quantumVerified || false
    };
  }

  /**
   * Estimate gas for quantum transaction
   */
  async estimateGas(txRequest: Partial<QuantumTransaction>): Promise<string> {
    return await this.request({ 
      method: 'eth_estimateGas', 
      params: [txRequest] 
    });
  }

  /**
   * Get current gas price
   */
  async getGasPrice(): Promise<string> {
    return await this.request({ method: 'eth_gasPrice' });
  }

  /**
   * Get quantum network metrics
   */
  async getQuantumMetrics(): Promise<QuantumMetrics> {
    try {
      return await this.request({ method: 'quantum_getMetrics' });
    } catch {
      // Fallback to standard methods
      const [blockNumber, gasPrice] = await Promise.all([
        this.getBlockNumber(),
        this.getGasPrice()
      ]);

      return {
        blockHeight: blockNumber,
        blockTime: 2, // 2 second blocks
        tps: 0,
        pendingTransactions: 0,
        validators: 3,
        quantumSignatures: 0,
        networkHashRate: '0',
        difficulty: '0x1'
      };
    }
  }

  /**
   * Subscribe to quantum events
   */
  subscribe(event: string, callback: (data: any) => void): void {
    this.on(event, callback);
    
    // Subscribe via WebSocket if available
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const subscription = {
        jsonrpc: '2.0',
        id: ++this.requestId,
        method: 'eth_subscribe',
        params: [event]
      };
      this.ws.send(JSON.stringify(subscription));
    }
  }

  /**
   * Unsubscribe from events
   */
  unsubscribe(event: string, callback: (data: any) => void): void {
    this.removeListener(event, callback);
  }

  /**
   * Check if node supports quantum features
   */
  async isQuantumCompatible(): Promise<boolean> {
    try {
      await this.request({ method: 'quantum_getVersion' });
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Get supported quantum algorithms
   */
  async getSupportedAlgorithms(): Promise<string[]> {
    try {
      return await this.request({ method: 'quantum_getSupportedAlgorithms' });
    } catch {
      return ['dilithium-ii', 'falcon-512']; // Default supported
    }
  }

  /**
   * Close connections
   */
  close(): void {
    if (this.ws) {
      this.ws.close();
    }
    this.removeAllListeners();
  }

  // Private helper methods

  private initializeWebSocket(wsUrl: string): void {
    try {
      // Only initialize WebSocket in Node.js environment
      if (typeof window === 'undefined') {
        this.ws = new WebSocket(wsUrl);
        
        this.ws.on('open', () => {
          this.emit('connected');
        });

        this.ws.on('message', (data: string) => {
          try {
            const message = JSON.parse(data);
            this.handleWebSocketMessage(message);
          } catch (error) {
            this.emit('error', error);
          }
        });

        this.ws.on('error', (error: Error) => {
          this.emit('error', error);
        });

        this.ws.on('close', () => {
          this.emit('disconnected');
          // Attempt reconnection after 5 seconds
          setTimeout(() => this.initializeWebSocket(wsUrl), 5000);
        });
      }
    } catch (error) {
      // WebSocket not available or failed to connect
      console.warn('WebSocket connection failed:', error);
    }
  }

  private handleWebSocketMessage(message: any): void {
    if (message.method === 'eth_subscription') {
      const { subscription, result } = message.params;
      this.emit(`subscription_${subscription}`, result);
    } else if (message.id && this.pendingRequests.has(message.id)) {
      const { resolve, reject } = this.pendingRequests.get(message.id)!;
      this.pendingRequests.delete(message.id);
      
      if (message.error) {
        reject(new Error(message.error.message));
      } else {
        resolve(message.result);
      }
    }
  }

  private mapToQuantumTransaction(tx: any): QuantumTransaction {
    return {
      nonce: tx.nonce,
      gasPrice: tx.gasPrice,
      gasLimit: tx.gas,
      to: tx.to,
      value: tx.value,
      data: tx.input || tx.data,
      chainId: this.config.chainId,
      signatureAlgorithm: tx.signatureAlgorithm || 1,
      publicKey: tx.publicKey || '',
      signature: tx.signature || ''
    };
  }

  private serializeQuantumTransaction(tx: QuantumTransaction): string {
    // In production, use proper RLP encoding
    return JSON.stringify(tx);
  }
}

/**
 * Create quantum provider instance
 */
export function createQuantumProvider(config: ProviderConfig): QuantumProvider {
  return new QuantumProvider(config);
}

/**
 * Default quantum networks
 */
export const QuantumNetworks = {
  mainnet: {
    chainId: 8888,
    name: 'Quantum Mainnet',
    rpcUrl: 'https://rpc.quantum-blockchain.org',
    wsUrl: 'wss://ws.quantum-blockchain.org',
    explorerUrl: 'https://explorer.quantum-blockchain.org',
    currency: {
      name: 'Quantum Token',
      symbol: 'QTM',
      decimals: 18
    },
    quantumFeatures: {
      postQuantumCrypto: true,
      supportedAlgorithms: [1, 2], // Dilithium, Falcon
      kemAlgorithms: [1, 2, 3],   // Kyber variants
      blockTime: 2
    }
  },
  testnet: {
    chainId: 8889,
    name: 'Quantum Testnet',
    rpcUrl: 'https://testnet-rpc.quantum-blockchain.org',
    wsUrl: 'wss://testnet-ws.quantum-blockchain.org',
    explorerUrl: 'https://testnet-explorer.quantum-blockchain.org',
    currency: {
      name: 'Test Quantum Token',
      symbol: 'tQTM',
      decimals: 18
    },
    quantumFeatures: {
      postQuantumCrypto: true,
      supportedAlgorithms: [1, 2],
      kemAlgorithms: [1, 2, 3],
      blockTime: 2
    }
  },
  localhost: {
    chainId: 8888,
    name: 'Quantum Local',
    rpcUrl: 'http://localhost:8545',
    wsUrl: 'ws://localhost:8546',
    explorerUrl: 'http://localhost:3000',
    currency: {
      name: 'Quantum Token',
      symbol: 'QTM',
      decimals: 18
    },
    quantumFeatures: {
      postQuantumCrypto: true,
      supportedAlgorithms: [1, 2],
      kemAlgorithms: [1, 2, 3],
      blockTime: 2
    }
  }
};