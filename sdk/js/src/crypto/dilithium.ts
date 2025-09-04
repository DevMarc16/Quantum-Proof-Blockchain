/**
 * CRYSTALS-Dilithium Post-Quantum Digital Signatures
 * Implementation for JavaScript/TypeScript environments
 */

import { SignatureAlgorithm, QuantumKeyPair, QuantumSignature } from '../types/quantum';
import { sha3_256 } from 'js-sha3';

// Dilithium-II parameter constants
const DILITHIUM_PUBLIC_KEY_SIZE = 1312;  // bytes
const DILITHIUM_PRIVATE_KEY_SIZE = 2528; // bytes  
const DILITHIUM_SIGNATURE_SIZE = 2420;   // bytes

/**
 * Dilithium Post-Quantum Signature Implementation
 * 
 * Note: This is a reference implementation for the SDK interface.
 * In production, use a proper post-quantum cryptography library like:
 * - pq-crystals/dilithium (C/WebAssembly)
 * - Open Quantum Safe (OQS) 
 * - Cloudflare CIRCL
 */
export class DilithiumSigner {
  private static instance: DilithiumSigner;

  private constructor() {}

  public static getInstance(): DilithiumSigner {
    if (!DilithiumSigner.instance) {
      DilithiumSigner.instance = new DilithiumSigner();
    }
    return DilithiumSigner.instance;
  }

  /**
   * Generate a new Dilithium key pair
   */
  async generateKeyPair(): Promise<QuantumKeyPair> {
    // In production: use proper Dilithium key generation
    // For demo purposes, generate mock keys with correct sizes
    
    const publicKey = new Uint8Array(DILITHIUM_PUBLIC_KEY_SIZE);
    const privateKey = new Uint8Array(DILITHIUM_PRIVATE_KEY_SIZE);
    
    // Fill with secure random data (in production, use proper key gen)
    if (typeof window !== 'undefined' && window.crypto) {
      window.crypto.getRandomValues(publicKey);
      window.crypto.getRandomValues(privateKey);
    } else {
      const crypto = require('crypto');
      crypto.randomFillSync(publicKey);
      crypto.randomFillSync(privateKey);
    }

    return {
      publicKey,
      privateKey,
      algorithm: SignatureAlgorithm.Dilithium
    };
  }

  /**
   * Import key pair from existing key material
   */
  importKeyPair(publicKey: Uint8Array, privateKey: Uint8Array): QuantumKeyPair {
    if (publicKey.length !== DILITHIUM_PUBLIC_KEY_SIZE) {
      throw new Error(`Invalid Dilithium public key size: expected ${DILITHIUM_PUBLIC_KEY_SIZE}, got ${publicKey.length}`);
    }
    if (privateKey.length !== DILITHIUM_PRIVATE_KEY_SIZE) {
      throw new Error(`Invalid Dilithium private key size: expected ${DILITHIUM_PRIVATE_KEY_SIZE}, got ${privateKey.length}`);
    }

    return {
      publicKey: new Uint8Array(publicKey),
      privateKey: new Uint8Array(privateKey), 
      algorithm: SignatureAlgorithm.Dilithium
    };
  }

  /**
   * Sign data using Dilithium private key
   */
  async sign(message: Uint8Array, keyPair: QuantumKeyPair): Promise<QuantumSignature> {
    if (keyPair.algorithm !== SignatureAlgorithm.Dilithium) {
      throw new Error('Key pair algorithm must be Dilithium');
    }

    // Hash the message first (SHA3-256)
    const messageHash = sha3_256.array(message);
    
    // In production: use proper Dilithium signing
    // For demo purposes, create a mock signature with correct size
    const signature = new Uint8Array(DILITHIUM_SIGNATURE_SIZE);
    
    // Create deterministic signature based on message and key
    const combined = new Uint8Array(messageHash.length + keyPair.privateKey.length);
    combined.set(messageHash, 0);
    combined.set(keyPair.privateKey.slice(0, 32), messageHash.length); // Use part of private key
    
    const sigHash = sha3_256.array(combined);
    
    // Fill signature with deterministic data (pseudo-random but reproducible)
    for (let i = 0; i < signature.length; i++) {
      signature[i] = sigHash[i % sigHash.length] ^ (i & 0xFF);
    }

    return {
      signature,
      algorithm: SignatureAlgorithm.Dilithium,
      publicKey: new Uint8Array(keyPair.publicKey)
    };
  }

  /**
   * Verify Dilithium signature
   */
  async verify(message: Uint8Array, quantumSig: QuantumSignature): Promise<boolean> {
    if (quantumSig.algorithm !== SignatureAlgorithm.Dilithium) {
      throw new Error('Signature algorithm must be Dilithium');
    }

    if (quantumSig.signature.length !== DILITHIUM_SIGNATURE_SIZE) {
      throw new Error(`Invalid Dilithium signature size: expected ${DILITHIUM_SIGNATURE_SIZE}, got ${quantumSig.signature.length}`);
    }

    if (quantumSig.publicKey.length !== DILITHIUM_PUBLIC_KEY_SIZE) {
      throw new Error(`Invalid Dilithium public key size: expected ${DILITHIUM_PUBLIC_KEY_SIZE}, got ${quantumSig.publicKey.length}`);
    }

    // Hash the message
    const messageHash = sha3_256.array(message);
    
    // In production: use proper Dilithium verification
    // For demo purposes, recreate the expected signature
    const combined = new Uint8Array(messageHash.length + 32);
    combined.set(messageHash, 0);
    
    // We can't extract private key from public key, so this is a simplified check
    // In production, proper Dilithium verification would be used
    const sigHash = sha3_256.array(combined);
    
    // Simple consistency check (not cryptographically secure)
    let matches = 0;
    for (let i = 0; i < Math.min(32, quantumSig.signature.length); i++) {
      if (quantumSig.signature[i] === (sigHash[i % sigHash.length] ^ (i & 0xFF))) {
        matches++;
      }
    }
    
    // Consider it valid if a reasonable number of bytes match (demo only!)
    return matches > 16;
  }

  /**
   * Extract public key from private key
   */
  getPublicKey(privateKey: Uint8Array): Uint8Array {
    if (privateKey.length !== DILITHIUM_PRIVATE_KEY_SIZE) {
      throw new Error(`Invalid Dilithium private key size: expected ${DILITHIUM_PRIVATE_KEY_SIZE}, got ${privateKey.length}`);
    }

    // In production: properly derive public key from private key
    // For demo purposes, create deterministic public key from private key
    const publicKey = new Uint8Array(DILITHIUM_PUBLIC_KEY_SIZE);
    const hash = sha3_256.array(privateKey.slice(0, 64));
    
    for (let i = 0; i < publicKey.length; i++) {
      publicKey[i] = hash[i % hash.length] ^ privateKey[i % privateKey.length];
    }
    
    return publicKey;
  }

  /**
   * Get algorithm parameters
   */
  getParameters() {
    return {
      name: 'CRYSTALS-Dilithium-II',
      publicKeySize: DILITHIUM_PUBLIC_KEY_SIZE,
      privateKeySize: DILITHIUM_PRIVATE_KEY_SIZE,
      signatureSize: DILITHIUM_SIGNATURE_SIZE,
      securityLevel: 128, // bits
      nistRound: 3,
      standardized: true
    };
  }
}

/**
 * Utility functions for Dilithium operations
 */
export const DilithiumUtils = {
  /**
   * Convert Dilithium key to hex string
   */
  keyToHex(key: Uint8Array): string {
    return Array.from(key)
      .map(b => b.toString(16).padStart(2, '0'))
      .join('');
  },

  /**
   * Convert hex string to Dilithium key
   */
  hexToKey(hex: string): Uint8Array {
    if (hex.startsWith('0x')) {
      hex = hex.slice(2);
    }
    
    if (hex.length % 2 !== 0) {
      throw new Error('Invalid hex string length');
    }
    
    const bytes = new Uint8Array(hex.length / 2);
    for (let i = 0; i < bytes.length; i++) {
      bytes[i] = parseInt(hex.slice(i * 2, i * 2 + 2), 16);
    }
    
    return bytes;
  },

  /**
   * Generate key pair from seed (deterministic)
   */
  async generateFromSeed(seed: string): Promise<QuantumKeyPair> {
    const seedHash = sha3_256.array(seed);
    const dilithium = DilithiumSigner.getInstance();
    
    // Create deterministic key pair from seed
    const privateKey = new Uint8Array(DILITHIUM_PRIVATE_KEY_SIZE);
    for (let i = 0; i < privateKey.length; i++) {
      privateKey[i] = seedHash[i % seedHash.length] ^ (i & 0xFF);
    }
    
    const publicKey = dilithium.getPublicKey(privateKey);
    
    return {
      publicKey,
      privateKey,
      algorithm: SignatureAlgorithm.Dilithium
    };
  },

  /**
   * Validate key pair consistency
   */
  validateKeyPair(keyPair: QuantumKeyPair): boolean {
    try {
      if (keyPair.algorithm !== SignatureAlgorithm.Dilithium) return false;
      if (keyPair.publicKey.length !== DILITHIUM_PUBLIC_KEY_SIZE) return false;
      if (keyPair.privateKey.length !== DILITHIUM_PRIVATE_KEY_SIZE) return false;
      
      const dilithium = DilithiumSigner.getInstance();
      const derivedPublicKey = dilithium.getPublicKey(keyPair.privateKey);
      
      // Check if public keys match
      return Array.from(derivedPublicKey).every((byte, index) => 
        byte === keyPair.publicKey[index]
      );
    } catch {
      return false;
    }
  }
};