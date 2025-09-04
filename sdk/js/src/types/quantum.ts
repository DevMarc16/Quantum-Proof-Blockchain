/**
 * TypeScript definitions for Quantum Blockchain SDK
 * Supports post-quantum cryptographic operations
 */

export enum SignatureAlgorithm {
  Dilithium = 1,
  Falcon = 2,
}

export enum KEMAlgorithm {
  Kyber512 = 1,
  Kyber768 = 2,
  Kyber1024 = 3,
}

export interface QuantumKeyPair {
  publicKey: Uint8Array;
  privateKey: Uint8Array;
  algorithm: SignatureAlgorithm;
}

export interface QuantumSignature {
  signature: Uint8Array;
  algorithm: SignatureAlgorithm;
  publicKey: Uint8Array;
}

export interface QuantumTransaction {
  nonce: string;
  gasPrice: string;
  gasLimit: string;
  to: string;
  value: string;
  data: string;
  chainId: number;
  signatureAlgorithm: SignatureAlgorithm;
  publicKey: string;
  signature: string;
}

export interface QuantumTransactionRequest {
  from?: string;
  to?: string;
  value?: string;
  gasLimit?: string;
  gasPrice?: string;
  data?: string;
  nonce?: number;
  signatureAlgorithm?: SignatureAlgorithm;
}

export interface QuantumBlock {
  number: string;
  hash: string;
  parentHash: string;
  timestamp: string;
  validator: string;
  quantumSignature: string;
  transactions: QuantumTransaction[];
  gasUsed: string;
  gasLimit: string;
}

export interface QuantumAccount {
  address: string;
  balance: string;
  nonce: string;
  quantumPublicKey?: string;
  signatureAlgorithm?: SignatureAlgorithm;
}

export interface RPCRequest {
  jsonrpc: '2.0';
  id: number;
  method: string;
  params: any[];
}

export interface RPCResponse<T = any> {
  jsonrpc: '2.0';
  id: number;
  result?: T;
  error?: {
    code: number;
    message: string;
    data?: any;
  };
}

export interface ProviderConfig {
  url: string;
  chainId: number;
  timeout?: number;
  retries?: number;
  headers?: Record<string, string>;
}

export interface Web3QuantumProvider {
  request(args: { method: string; params?: any[] }): Promise<any>;
  on(event: string, listener: (...args: any[]) => void): void;
  removeListener(event: string, listener: (...args: any[]) => void): void;
  isQuantumEnabled: boolean;
  chainId: string;
}

export interface QuantumContractABI {
  name: string;
  type: 'function' | 'constructor' | 'event' | 'fallback' | 'receive';
  inputs: Array<{
    name: string;
    type: string;
    indexed?: boolean;
  }>;
  outputs?: Array<{
    name: string;
    type: string;
  }>;
  stateMutability?: 'pure' | 'view' | 'nonpayable' | 'payable';
  quantumSafe?: boolean;
}

export interface QuantumContract {
  address: string;
  abi: QuantumContractABI[];
  methods: Record<string, (...args: any[]) => any>;
  events: Record<string, any>;
}

export interface QuantumWalletConfig {
  provider: string;
  mnemonic?: string;
  privateKey?: string;
  signatureAlgorithm: SignatureAlgorithm;
  derivationPath?: string;
}

export interface QuantumNetworkConfig {
  chainId: number;
  name: string;
  rpcUrl: string;
  wsUrl?: string;
  explorerUrl?: string;
  currency: {
    name: string;
    symbol: string;
    decimals: number;
  };
  quantumFeatures: {
    postQuantumCrypto: boolean;
    supportedAlgorithms: SignatureAlgorithm[];
    kemAlgorithms: KEMAlgorithm[];
    blockTime: number;
  };
}

export interface QuantumEventLog {
  address: string;
  topics: string[];
  data: string;
  blockNumber: string;
  blockHash: string;
  transactionHash: string;
  transactionIndex: string;
  logIndex: string;
  removed: boolean;
}

export interface QuantumTransactionReceipt {
  transactionHash: string;
  transactionIndex: string;
  blockNumber: string;
  blockHash: string;
  from: string;
  to: string;
  gasUsed: string;
  effectiveGasPrice: string;
  status: string;
  logs: QuantumEventLog[];
  contractAddress?: string;
  quantumVerified: boolean;
}

export interface QuantumMetrics {
  blockHeight: number;
  blockTime: number;
  tps: number;
  pendingTransactions: number;
  validators: number;
  quantumSignatures: number;
  networkHashRate: string;
  difficulty: string;
}