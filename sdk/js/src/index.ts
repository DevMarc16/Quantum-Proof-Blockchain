/**
 * Quantum Blockchain JavaScript/TypeScript SDK
 * Complete toolkit for building applications on quantum-resistant blockchain
 */

// Core Types
export * from './types/quantum';

// Cryptographic Modules
export { DilithiumSigner, DilithiumUtils } from './crypto/dilithium';

// Provider and Network Layer
export { QuantumProvider, createQuantumProvider, QuantumNetworks } from './provider/quantum-provider';

// Wallet Management
export { QuantumWallet, QuantumWalletUtils } from './wallet/quantum-wallet';

// Contract Interaction
export { QuantumContract } from './contracts/quantum-contract';

// Utilities
export { QuantumUtils } from './utils/quantum-utils';

// Version
export const VERSION = '1.0.0';

/**
 * Quick start configuration for common use cases
 */
export const QuickStart = {
  /**
   * Connect to local development node
   */
  async connectLocal() {
    const { createQuantumProvider, QuantumNetworks } = await import('./provider/quantum-provider');
    return createQuantumProvider({
      url: QuantumNetworks.localhost.rpcUrl,
      chainId: QuantumNetworks.localhost.chainId
    });
  },

  /**
   * Connect to testnet
   */
  async connectTestnet() {
    const { createQuantumProvider, QuantumNetworks } = await import('./provider/quantum-provider');
    return createQuantumProvider({
      url: QuantumNetworks.testnet.rpcUrl,
      chainId: QuantumNetworks.testnet.chainId
    });
  },

  /**
   * Connect to mainnet
   */
  async connectMainnet() {
    const { createQuantumProvider, QuantumNetworks } = await import('./provider/quantum-provider');
    return createQuantumProvider({
      url: QuantumNetworks.mainnet.rpcUrl,
      chainId: QuantumNetworks.mainnet.chainId
    });
  },

  /**
   * Create new random wallet
   */
  async createWallet(provider?: any) {
    const { QuantumWallet, SignatureAlgorithm } = await Promise.all([
      import('./wallet/quantum-wallet'),
      import('./types/quantum')
    ]);
    
    return await QuantumWallet.QuantumWallet.random(
      SignatureAlgorithm.SignatureAlgorithm.Dilithium,
      provider
    );
  },

  /**
   * Import wallet from private key
   */
  async importWallet(privateKey: string, provider?: any) {
    const { QuantumWallet, SignatureAlgorithm } = await Promise.all([
      import('./wallet/quantum-wallet'),
      import('./types/quantum')
    ]);
    
    return await QuantumWallet.QuantumWallet.fromPrivateKey(
      privateKey,
      SignatureAlgorithm.SignatureAlgorithm.Dilithium,
      provider
    );
  }
};

/**
 * Default export for convenience
 */
export default {
  VERSION,
  QuickStart,
  // Re-export commonly used classes
  QuantumProvider,
  QuantumWallet,
  DilithiumSigner,
  QuantumNetworks,
  createQuantumProvider
};