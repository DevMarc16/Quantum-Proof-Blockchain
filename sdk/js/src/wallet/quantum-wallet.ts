/**
 * Quantum Wallet Implementation
 * Handles quantum-resistant key management and transaction signing
 */

import { EventEmitter } from 'events';
import { sha3_256, keccak256 } from 'js-sha3';
import { 
  QuantumKeyPair, 
  QuantumWalletConfig, 
  QuantumTransaction, 
  QuantumTransactionRequest,
  SignatureAlgorithm,
  QuantumSignature
} from '../types/quantum';
import { DilithiumSigner, DilithiumUtils } from '../crypto/dilithium';
import { QuantumProvider } from '../provider/quantum-provider';

export class QuantumWallet extends EventEmitter {
  private keyPair: QuantumKeyPair;
  private provider?: QuantumProvider;
  private address: string;
  private config: QuantumWalletConfig;

  constructor(config: QuantumWalletConfig, provider?: QuantumProvider) {
    super();
    this.config = config;
    this.provider = provider;
    this.keyPair = {} as QuantumKeyPair; // Will be initialized
    this.address = '';
  }

  /**
   * Initialize wallet from private key
   */
  static async fromPrivateKey(
    privateKey: string | Uint8Array, 
    algorithm: SignatureAlgorithm = SignatureAlgorithm.Dilithium,
    provider?: QuantumProvider
  ): Promise<QuantumWallet> {
    const config: QuantumWalletConfig = {
      provider: provider?.config.url || '',
      signatureAlgorithm: algorithm
    };

    const wallet = new QuantumWallet(config, provider);
    await wallet.importPrivateKey(privateKey, algorithm);
    return wallet;
  }

  /**
   * Initialize wallet from mnemonic phrase
   */
  static async fromMnemonic(
    mnemonic: string,
    derivationPath: string = "m/44'/8888'/0'/0/0",
    algorithm: SignatureAlgorithm = SignatureAlgorithm.Dilithium,
    provider?: QuantumProvider
  ): Promise<QuantumWallet> {
    const config: QuantumWalletConfig = {
      provider: provider?.config.url || '',
      mnemonic,
      derivationPath,
      signatureAlgorithm: algorithm
    };

    const wallet = new QuantumWallet(config, provider);
    await wallet.initializeFromMnemonic();
    return wallet;
  }

  /**
   * Generate new random wallet
   */
  static async random(
    algorithm: SignatureAlgorithm = SignatureAlgorithm.Dilithium,
    provider?: QuantumProvider
  ): Promise<QuantumWallet> {
    const config: QuantumWalletConfig = {
      provider: provider?.config.url || '',
      signatureAlgorithm: algorithm
    };

    const wallet = new QuantumWallet(config, provider);
    await wallet.generateNewKeyPair();
    return wallet;
  }

  /**
   * Get wallet address
   */
  getAddress(): string {
    return this.address;
  }

  /**
   * Get public key
   */
  getPublicKey(): Uint8Array {
    return this.keyPair.publicKey;
  }

  /**
   * Get signature algorithm
   */
  getSignatureAlgorithm(): SignatureAlgorithm {
    return this.keyPair.algorithm;
  }

  /**
   * Get account balance
   */
  async getBalance(): Promise<string> {
    if (!this.provider) {
      throw new Error('Provider not configured');
    }
    
    const account = await this.provider.getAccount(this.address);
    return account.balance;
  }

  /**
   * Get transaction count (nonce)
   */
  async getTransactionCount(): Promise<number> {
    if (!this.provider) {
      throw new Error('Provider not configured');
    }
    
    const account = await this.provider.getAccount(this.address);
    return parseInt(account.nonce, 16);
  }

  /**
   * Sign quantum transaction
   */
  async signTransaction(txRequest: QuantumTransactionRequest): Promise<QuantumTransaction> {
    // Fill in missing transaction fields
    const tx: QuantumTransaction = {
      nonce: txRequest.nonce?.toString(16) || await this.getNonceHex(),
      gasPrice: txRequest.gasPrice || await this.getGasPriceHex(),
      gasLimit: txRequest.gasLimit || '0x5208', // 21000 gas default
      to: txRequest.to || '',
      value: txRequest.value || '0x0',
      data: txRequest.data || '0x',
      chainId: this.provider?.config.chainId || 8888,
      signatureAlgorithm: this.keyPair.algorithm,
      publicKey: '',
      signature: ''
    };

    // Create signing data (RLP-like encoding)
    const signingData = this.createSigningData(tx);
    
    // Sign with quantum algorithm
    const quantumSig = await this.signData(signingData);
    
    // Add signature and public key to transaction
    tx.publicKey = DilithiumUtils.keyToHex(quantumSig.publicKey);
    tx.signature = DilithiumUtils.keyToHex(quantumSig.signature);

    return tx;
  }

  /**
   * Send transaction
   */
  async sendTransaction(txRequest: QuantumTransactionRequest): Promise<string> {
    if (!this.provider) {
      throw new Error('Provider not configured');
    }

    const signedTx = await this.signTransaction(txRequest);
    const txHash = await this.provider.sendQuantumTransaction(signedTx);
    
    this.emit('transactionSent', { hash: txHash, transaction: signedTx });
    return txHash;
  }

  /**
   * Sign arbitrary data
   */
  async signMessage(message: string | Uint8Array): Promise<QuantumSignature> {
    const messageBytes = typeof message === 'string' 
      ? new TextEncoder().encode(message)
      : message;
    
    return await this.signData(messageBytes);
  }

  /**
   * Verify signature
   */
  async verifySignature(
    message: string | Uint8Array, 
    signature: QuantumSignature
  ): Promise<boolean> {
    const messageBytes = typeof message === 'string' 
      ? new TextEncoder().encode(message)
      : message;

    if (signature.algorithm === SignatureAlgorithm.Dilithium) {
      const dilithium = DilithiumSigner.getInstance();
      return await dilithium.verify(messageBytes, signature);
    }
    
    throw new Error(`Unsupported signature algorithm: ${signature.algorithm}`);
  }

  /**
   * Export private key
   */
  exportPrivateKey(): string {
    return DilithiumUtils.keyToHex(this.keyPair.privateKey);
  }

  /**
   * Export wallet to JSON
   */
  exportWallet(password: string): string {
    // In production: encrypt with password using AES-256-GCM
    const walletData = {
      version: 1,
      address: this.address,
      algorithm: this.keyPair.algorithm,
      publicKey: DilithiumUtils.keyToHex(this.keyPair.publicKey),
      privateKey: DilithiumUtils.keyToHex(this.keyPair.privateKey), // Should be encrypted
      createdAt: new Date().toISOString()
    };
    
    return JSON.stringify(walletData, null, 2);
  }

  /**
   * Import wallet from JSON
   */
  static async fromJSON(
    jsonWallet: string, 
    password: string,
    provider?: QuantumProvider
  ): Promise<QuantumWallet> {
    const walletData = JSON.parse(jsonWallet);
    
    // In production: decrypt privateKey with password
    const privateKey = DilithiumUtils.hexToKey(walletData.privateKey);
    
    return await QuantumWallet.fromPrivateKey(
      privateKey, 
      walletData.algorithm, 
      provider
    );
  }

  // Private helper methods

  private async generateNewKeyPair(): Promise<void> {
    if (this.config.signatureAlgorithm === SignatureAlgorithm.Dilithium) {
      const dilithium = DilithiumSigner.getInstance();
      this.keyPair = await dilithium.generateKeyPair();
    } else {
      throw new Error(`Unsupported signature algorithm: ${this.config.signatureAlgorithm}`);
    }
    
    this.address = this.deriveAddress(this.keyPair.publicKey);
  }

  private async importPrivateKey(
    privateKey: string | Uint8Array, 
    algorithm: SignatureAlgorithm
  ): Promise<void> {
    const privateKeyBytes = typeof privateKey === 'string' 
      ? DilithiumUtils.hexToKey(privateKey)
      : privateKey;

    if (algorithm === SignatureAlgorithm.Dilithium) {
      const dilithium = DilithiumSigner.getInstance();
      const publicKey = dilithium.getPublicKey(privateKeyBytes);
      
      this.keyPair = {
        publicKey,
        privateKey: privateKeyBytes,
        algorithm
      };
    } else {
      throw new Error(`Unsupported signature algorithm: ${algorithm}`);
    }
    
    this.address = this.deriveAddress(this.keyPair.publicKey);
  }

  private async initializeFromMnemonic(): Promise<void> {
    if (!this.config.mnemonic) {
      throw new Error('Mnemonic not provided');
    }

    // Generate deterministic private key from mnemonic and derivation path
    const seed = this.config.mnemonic + (this.config.derivationPath || '');
    this.keyPair = await DilithiumUtils.generateFromSeed(seed);
    this.address = this.deriveAddress(this.keyPair.publicKey);
  }

  private deriveAddress(publicKey: Uint8Array): string {
    // Use Keccak-256 hash of public key for Ethereum compatibility
    const hash = keccak256.array(publicKey);
    // Take last 20 bytes as address
    const addressBytes = hash.slice(-20);
    return '0x' + Array.from(addressBytes)
      .map(b => b.toString(16).padStart(2, '0'))
      .join('');
  }

  private async signData(data: Uint8Array): Promise<QuantumSignature> {
    if (this.keyPair.algorithm === SignatureAlgorithm.Dilithium) {
      const dilithium = DilithiumSigner.getInstance();
      return await dilithium.sign(data, this.keyPair);
    }
    
    throw new Error(`Unsupported signature algorithm: ${this.keyPair.algorithm}`);
  }

  private createSigningData(tx: QuantumTransaction): Uint8Array {
    // Create deterministic signing data (simplified RLP encoding)
    const fields = [
      tx.nonce,
      tx.gasPrice,
      tx.gasLimit,
      tx.to,
      tx.value,
      tx.data,
      tx.chainId.toString()
    ];
    
    const concatenated = fields.join('|');
    return new TextEncoder().encode(concatenated);
  }

  private async getNonceHex(): Promise<string> {
    if (!this.provider) return '0x0';
    const nonce = await this.getTransactionCount();
    return '0x' + nonce.toString(16);
  }

  private async getGasPriceHex(): Promise<string> {
    if (!this.provider) return '0x3b9aca00'; // 1 Gwei
    const gasPrice = await this.provider.getGasPrice();
    return gasPrice;
  }
}

/**
 * Quantum Wallet Utilities
 */
export const QuantumWalletUtils = {
  /**
   * Validate quantum address
   */
  isValidAddress(address: string): boolean {
    return /^0x[a-fA-F0-9]{40}$/.test(address);
  },

  /**
   * Generate random mnemonic
   */
  generateMnemonic(): string {
    // In production: use proper BIP39 mnemonic generation
    const words = [
      'quantum', 'resistant', 'blockchain', 'cryptography', 'dilithium',
      'post', 'secure', 'future', 'protection', 'digital', 'signature', 'safe'
    ];
    
    const mnemonic = [];
    for (let i = 0; i < 12; i++) {
      mnemonic.push(words[Math.floor(Math.random() * words.length)]);
    }
    
    return mnemonic.join(' ');
  },

  /**
   * Validate mnemonic phrase
   */
  validateMnemonic(mnemonic: string): boolean {
    // Basic validation - in production use BIP39
    const words = mnemonic.trim().split(/\s+/);
    return words.length >= 12 && words.length <= 24;
  },

  /**
   * Format balance for display
   */
  formatBalance(balance: string, decimals: number = 18): string {
    const bn = BigInt(balance);
    const divisor = BigInt(10 ** decimals);
    const whole = bn / divisor;
    const fraction = bn % divisor;
    
    if (fraction === 0n) {
      return whole.toString();
    }
    
    const fractionStr = fraction.toString().padStart(decimals, '0');
    const trimmed = fractionStr.replace(/0+$/, '');
    
    return `${whole}.${trimmed}`;
  },

  /**
   * Parse balance from string
   */
  parseBalance(balance: string, decimals: number = 18): string {
    const [whole = '0', fraction = '0'] = balance.split('.');
    const fractionPadded = fraction.padEnd(decimals, '0').slice(0, decimals);
    const result = BigInt(whole) * BigInt(10 ** decimals) + BigInt(fractionPadded);
    return '0x' + result.toString(16);
  }
};