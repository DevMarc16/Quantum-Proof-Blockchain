/**
 * Quantum Blockchain Utilities
 * Common helper functions for quantum blockchain development
 */

import { sha3_256, keccak256 } from 'js-sha3';
import { SignatureAlgorithm } from '../types/quantum';

export class QuantumUtils {
  /**
   * Convert hex string to bytes
   */
  static hexToBytes(hex: string): Uint8Array {
    if (hex.startsWith('0x')) {
      hex = hex.slice(2);
    }
    
    if (hex.length % 2 !== 0) {
      hex = '0' + hex;
    }
    
    const bytes = new Uint8Array(hex.length / 2);
    for (let i = 0; i < bytes.length; i++) {
      bytes[i] = parseInt(hex.slice(i * 2, i * 2 + 2), 16);
    }
    
    return bytes;
  }

  /**
   * Convert bytes to hex string
   */
  static bytesToHex(bytes: Uint8Array): string {
    return '0x' + Array.from(bytes)
      .map(b => b.toString(16).padStart(2, '0'))
      .join('');
  }

  /**
   * Validate Ethereum-style address
   */
  static isValidAddress(address: string): boolean {
    return /^0x[a-fA-F0-9]{40}$/.test(address);
  }

  /**
   * Convert address to checksum format
   */
  static toChecksumAddress(address: string): string {
    if (!address.startsWith('0x')) {
      address = '0x' + address;
    }
    
    const addr = address.toLowerCase().slice(2);
    const hash = keccak256(addr);
    
    let checksum = '0x';
    for (let i = 0; i < addr.length; i++) {
      if (parseInt(hash[i], 16) >= 8) {
        checksum += addr[i].toUpperCase();
      } else {
        checksum += addr[i];
      }
    }
    
    return checksum;
  }

  /**
   * Generate random hex string
   */
  static randomHex(length: number): string {
    const bytes = new Uint8Array(length);
    
    if (typeof window !== 'undefined' && window.crypto) {
      window.crypto.getRandomValues(bytes);
    } else {
      const crypto = require('crypto');
      crypto.randomFillSync(bytes);
    }
    
    return this.bytesToHex(bytes);
  }

  /**
   * Hash data with SHA3-256
   */
  static sha3(data: string | Uint8Array): string {
    const bytes = typeof data === 'string' ? this.hexToBytes(data) : data;
    return '0x' + sha3_256(Array.from(bytes));
  }

  /**
   * Hash data with Keccak-256 (Ethereum style)
   */
  static keccak(data: string | Uint8Array): string {
    const bytes = typeof data === 'string' ? this.hexToBytes(data) : data;
    return '0x' + keccak256(Array.from(bytes));
  }

  /**
   * Convert Wei to Ether (18 decimals)
   */
  static fromWei(wei: string | bigint, unit: 'wei' | 'gwei' | 'ether' = 'ether'): string {
    const value = typeof wei === 'string' ? BigInt(wei) : wei;
    
    const decimals = {
      wei: 0,
      gwei: 9,
      ether: 18
    };
    
    const divisor = BigInt(10) ** BigInt(decimals[unit]);
    const quotient = value / divisor;
    const remainder = value % divisor;
    
    if (remainder === 0n) {
      return quotient.toString();
    }
    
    const decimalStr = remainder.toString().padStart(decimals[unit], '0');
    const trimmed = decimalStr.replace(/0+$/, '');
    
    return `${quotient}.${trimmed}`;
  }

  /**
   * Convert Ether to Wei
   */
  static toWei(value: string, unit: 'wei' | 'gwei' | 'ether' = 'ether'): string {
    const decimals = {
      wei: 0,
      gwei: 9,
      ether: 18
    };
    
    const [whole, fraction = '0'] = value.split('.');
    const fractionPadded = fraction.padEnd(decimals[unit], '0').slice(0, decimals[unit]);
    
    const result = BigInt(whole) * (BigInt(10) ** BigInt(decimals[unit])) + BigInt(fractionPadded);
    return result.toString();
  }

  /**
   * Generate deterministic address from public key
   */
  static publicKeyToAddress(publicKey: Uint8Array): string {
    const hash = keccak256(Array.from(publicKey));
    const addressBytes = hash.slice(-40); // Last 20 bytes
    return '0x' + addressBytes;
  }

  /**
   * Validate quantum signature algorithm
   */
  static isValidSignatureAlgorithm(algorithm: number): algorithm is SignatureAlgorithm {
    return Object.values(SignatureAlgorithm).includes(algorithm);
  }

  /**
   * Get signature algorithm name
   */
  static getSignatureAlgorithmName(algorithm: SignatureAlgorithm): string {
    const names = {
      [SignatureAlgorithm.Dilithium]: 'CRYSTALS-Dilithium-II',
      [SignatureAlgorithm.Falcon]: 'FALCON-512'
    };
    return names[algorithm] || 'Unknown';
  }

  /**
   * Estimate quantum signature size
   */
  static getSignatureSize(algorithm: SignatureAlgorithm): number {
    const sizes = {
      [SignatureAlgorithm.Dilithium]: 2420, // bytes
      [SignatureAlgorithm.Falcon]: 690     // bytes (approximate)
    };
    return sizes[algorithm] || 0;
  }

  /**
   * Estimate public key size
   */
  static getPublicKeySize(algorithm: SignatureAlgorithm): number {
    const sizes = {
      [SignatureAlgorithm.Dilithium]: 1312, // bytes
      [SignatureAlgorithm.Falcon]: 897     // bytes
    };
    return sizes[algorithm] || 0;
  }

  /**
   * Calculate transaction size with quantum signature
   */
  static estimateTransactionSize(
    dataSize: number = 0,
    algorithm: SignatureAlgorithm = SignatureAlgorithm.Dilithium
  ): number {
    const baseSize = 100; // Base transaction fields
    const signatureSize = this.getSignatureSize(algorithm);
    const publicKeySize = this.getPublicKeySize(algorithm);
    
    return baseSize + dataSize + signatureSize + publicKeySize;
  }

  /**
   * Calculate quantum-safe gas requirements
   */
  static estimateQuantumGas(
    baseGas: number = 21000,
    algorithm: SignatureAlgorithm = SignatureAlgorithm.Dilithium
  ): number {
    // Quantum signature verification costs
    const quantumGasCosts = {
      [SignatureAlgorithm.Dilithium]: 800,  // Optimized cost
      [SignatureAlgorithm.Falcon]: 1200    // Estimated cost
    };
    
    return baseGas + (quantumGasCosts[algorithm] || 0);
  }

  /**
   * Format block time in human readable format
   */
  static formatBlockTime(timestamp: string | number): string {
    const time = typeof timestamp === 'string' 
      ? parseInt(timestamp, 16) * 1000 
      : timestamp * 1000;
      
    return new Date(time).toISOString();
  }

  /**
   * Calculate time until next block
   */
  static timeUntilNextBlock(
    lastBlockTime: string | number,
    blockInterval: number = 2000 // 2 seconds in ms
  ): number {
    const lastTime = typeof lastBlockTime === 'string' 
      ? parseInt(lastBlockTime, 16) * 1000 
      : lastBlockTime * 1000;
      
    const nextBlockTime = lastTime + blockInterval;
    return Math.max(0, nextBlockTime - Date.now());
  }

  /**
   * Validate transaction parameters
   */
  static validateTransactionParams(params: {
    to?: string;
    value?: string;
    gasLimit?: string;
    gasPrice?: string;
    data?: string;
  }): string[] {
    const errors: string[] = [];
    
    if (params.to && !this.isValidAddress(params.to)) {
      errors.push('Invalid "to" address');
    }
    
    if (params.value && !this.isValidHex(params.value)) {
      errors.push('Invalid "value" format');
    }
    
    if (params.gasLimit && !this.isValidHex(params.gasLimit)) {
      errors.push('Invalid "gasLimit" format');
    }
    
    if (params.gasPrice && !this.isValidHex(params.gasPrice)) {
      errors.push('Invalid "gasPrice" format');
    }
    
    if (params.data && !this.isValidHex(params.data)) {
      errors.push('Invalid "data" format');
    }
    
    return errors;
  }

  /**
   * Validate hex string format
   */
  static isValidHex(hex: string): boolean {
    return /^0x[a-fA-F0-9]*$/.test(hex);
  }

  /**
   * Pad hex string to specified length
   */
  static padHex(hex: string, length: number): string {
    if (hex.startsWith('0x')) {
      hex = hex.slice(2);
    }
    return '0x' + hex.padStart(length, '0');
  }

  /**
   * Compare two addresses (case insensitive)
   */
  static addressesEqual(addr1: string, addr2: string): boolean {
    return addr1.toLowerCase() === addr2.toLowerCase();
  }

  /**
   * Generate quantum-safe random bytes
   */
  static secureRandomBytes(length: number): Uint8Array {
    const bytes = new Uint8Array(length);
    
    if (typeof window !== 'undefined' && window.crypto) {
      // Browser environment
      window.crypto.getRandomValues(bytes);
    } else if (typeof global !== 'undefined') {
      // Node.js environment
      try {
        const crypto = require('crypto');
        crypto.randomFillSync(bytes);
      } catch {
        // Fallback to Math.random (not cryptographically secure!)
        for (let i = 0; i < length; i++) {
          bytes[i] = Math.floor(Math.random() * 256);
        }
      }
    }
    
    return bytes;
  }

  /**
   * Create quantum transaction ID
   */
  static createTransactionId(tx: {
    from: string;
    to: string;
    value: string;
    nonce: string;
    data?: string;
  }): string {
    const txData = [
      tx.from,
      tx.to,
      tx.value,
      tx.nonce,
      tx.data || '0x'
    ].join('|');
    
    return this.keccak(new TextEncoder().encode(txData));
  }

  /**
   * Batch multiple operations with retry logic
   */
  static async retry<T>(
    operation: () => Promise<T>,
    maxRetries: number = 3,
    delay: number = 1000
  ): Promise<T> {
    let lastError: Error;
    
    for (let i = 0; i <= maxRetries; i++) {
      try {
        return await operation();
      } catch (error) {
        lastError = error as Error;
        
        if (i < maxRetries) {
          await this.sleep(delay * Math.pow(2, i)); // Exponential backoff
        }
      }
    }
    
    throw lastError!;
  }

  /**
   * Sleep for specified duration
   */
  static sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * Deep clone object
   */
  static deepClone<T>(obj: T): T {
    return JSON.parse(JSON.stringify(obj));
  }

  /**
   * Check if running in browser environment
   */
  static isBrowser(): boolean {
    return typeof window !== 'undefined' && typeof window.document !== 'undefined';
  }

  /**
   * Check if running in Node.js environment
   */
  static isNode(): boolean {
    return typeof process !== 'undefined' && process.versions?.node !== undefined;
  }
}