/**
 * Quantum Blockchain MetaMask Snap
 * Enables post-quantum cryptography in MetaMask
 */

import { OnRpcRequestHandler } from '@metamask/snaps-types';
import { panel, text, heading, divider } from '@metamask/snaps-ui';
import { QuantumWallet, SignatureAlgorithm, DilithiumUtils } from '@quantum-blockchain/sdk';

interface QuantumSnapState {
  accounts: Record<string, {
    address: string;
    publicKey: string;
    algorithm: SignatureAlgorithm;
    created: number;
  }>;
  currentAccount?: string;
}

/**
 * Handle incoming JSON-RPC requests from dApps
 */
export const onRpcRequest: OnRpcRequestHandler = async ({ origin, request }) => {
  console.log(`Quantum Snap RPC request from ${origin}:`, request.method);

  try {
    switch (request.method) {
      case 'quantum_getAccounts':
        return await getQuantumAccounts();
      
      case 'quantum_createAccount':
        return await createQuantumAccount();
      
      case 'quantum_getPublicKey':
        return await getQuantumPublicKey(request.params as { address?: string });
      
      case 'quantum_signMessage':
        return await signQuantumMessage(
          request.params as { address: string; message: string }
        );
      
      case 'quantum_signTransaction':
        return await signQuantumTransaction(
          request.params as { transaction: any }
        );
      
      case 'quantum_getAlgorithmInfo':
        return await getAlgorithmInfo();
      
      case 'quantum_exportAccount':
        return await exportQuantumAccount(
          request.params as { address: string; password: string }
        );
      
      case 'quantum_importAccount':
        return await importQuantumAccount(
          request.params as { privateKey: string; algorithm?: SignatureAlgorithm }
        );
      
      default:
        throw new Error(`Method "${request.method}" not supported`);
    }
  } catch (error: any) {
    console.error('Quantum Snap error:', error);
    throw new Error(`Quantum Snap error: ${error.message}`);
  }
};

/**
 * Get all quantum accounts
 */
async function getQuantumAccounts(): Promise<string[]> {
  const state = await getSnapState();
  return Object.keys(state.accounts);
}

/**
 * Create new quantum account
 */
async function createQuantumAccount(): Promise<{
  address: string;
  publicKey: string;
  algorithm: SignatureAlgorithm;
}> {
  // Show account creation dialog
  const confirmed = await snap.request({
    method: 'snap_dialog',
    params: {
      type: 'confirmation',
      content: panel([
        heading('Create Quantum Account'),
        text('This will create a new quantum-resistant account using CRYSTALS-Dilithium-II.'),
        text('⚠️ Make sure to backup your recovery phrase!'),
        divider(),
        text('Algorithm: CRYSTALS-Dilithium-II'),
        text('Security Level: 128-bit post-quantum'),
        text('Signature Size: 2420 bytes')
      ])
    }
  });

  if (!confirmed) {
    throw new Error('Account creation cancelled by user');
  }

  // Get entropy for key generation
  const entropy = await snap.request({
    method: 'snap_getBip44Entropy',
    params: {
      coinType: 8888 // Quantum Blockchain coin type
    }
  });

  // Generate deterministic private key from entropy
  const privateKey = DilithiumUtils.hexToKey(entropy.privateKey.slice(2, 66)); // Use first 32 bytes
  
  // Create quantum wallet
  const wallet = await QuantumWallet.fromPrivateKey(privateKey, SignatureAlgorithm.Dilithium);
  const address = wallet.getAddress();
  const publicKey = DilithiumUtils.keyToHex(wallet.getPublicKey());

  // Store account in snap state
  const state = await getSnapState();
  state.accounts[address] = {
    address,
    publicKey,
    algorithm: SignatureAlgorithm.Dilithium,
    created: Date.now()
  };
  
  if (!state.currentAccount) {
    state.currentAccount = address;
  }
  
  await setSnapState(state);

  // Show success dialog
  await snap.request({
    method: 'snap_dialog',
    params: {
      type: 'alert',
      content: panel([
        heading('✅ Quantum Account Created'),
        text(`Address: ${address}`),
        text(`Algorithm: CRYSTALS-Dilithium-II`),
        text(`Public Key Size: ${wallet.getPublicKey().length} bytes`),
        divider(),
        text('🔐 Your quantum-resistant account is ready!')
      ])
    }
  });

  return {
    address,
    publicKey,
    algorithm: SignatureAlgorithm.Dilithium
  };
}

/**
 * Get public key for account
 */
async function getQuantumPublicKey(params: { address?: string }): Promise<{
  publicKey: string;
  algorithm: SignatureAlgorithm;
}> {
  const state = await getSnapState();
  const address = params.address || state.currentAccount;
  
  if (!address || !state.accounts[address]) {
    throw new Error('Account not found');
  }

  const account = state.accounts[address];
  return {
    publicKey: account.publicKey,
    algorithm: account.algorithm
  };
}

/**
 * Sign message with quantum cryptography
 */
async function signQuantumMessage(params: { address: string; message: string }): Promise<{
  signature: string;
  publicKey: string;
  algorithm: SignatureAlgorithm;
}> {
  const state = await getSnapState();
  const account = state.accounts[params.address];
  
  if (!account) {
    throw new Error('Account not found');
  }

  // Show signing confirmation dialog
  const confirmed = await snap.request({
    method: 'snap_dialog',
    params: {
      type: 'confirmation',
      content: panel([
        heading('🔐 Sign Quantum Message'),
        text(`Account: ${params.address}`),
        text(`Algorithm: ${getAlgorithmName(account.algorithm)}`),
        divider(),
        text('Message:'),
        text(params.message),
        divider(),
        text('⚠️ This will create a quantum-resistant signature')
      ])
    }
  });

  if (!confirmed) {
    throw new Error('Message signing cancelled by user');
  }

  // Get entropy and recreate wallet
  const entropy = await snap.request({
    method: 'snap_getBip44Entropy',
    params: { coinType: 8888 }
  });

  const privateKey = DilithiumUtils.hexToKey(entropy.privateKey.slice(2, 66));
  const wallet = await QuantumWallet.fromPrivateKey(privateKey, account.algorithm);

  // Sign message
  const quantumSignature = await wallet.signMessage(params.message);

  return {
    signature: DilithiumUtils.keyToHex(quantumSignature.signature),
    publicKey: DilithiumUtils.keyToHex(quantumSignature.publicKey),
    algorithm: quantumSignature.algorithm
  };
}

/**
 * Sign transaction with quantum cryptography
 */
async function signQuantumTransaction(params: { transaction: any }): Promise<{
  signature: string;
  publicKey: string;
  algorithm: SignatureAlgorithm;
  signedTransaction: any;
}> {
  const state = await getSnapState();
  const fromAddress = params.transaction.from;
  const account = state.accounts[fromAddress];
  
  if (!account) {
    throw new Error('Account not found');
  }

  // Show transaction confirmation dialog
  const confirmed = await snap.request({
    method: 'snap_dialog',
    params: {
      type: 'confirmation',
      content: panel([
        heading('📝 Sign Quantum Transaction'),
        text(`From: ${fromAddress}`),
        text(`To: ${params.transaction.to || 'Contract Creation'}`),
        text(`Value: ${params.transaction.value || '0'} wei`),
        text(`Gas: ${params.transaction.gasLimit || 'auto'}`),
        divider(),
        text(`Algorithm: ${getAlgorithmName(account.algorithm)}`),
        text('Signature Size: 2420 bytes'),
        divider(),
        text('⚠️ This transaction will use post-quantum cryptography')
      ])
    }
  });

  if (!confirmed) {
    throw new Error('Transaction signing cancelled by user');
  }

  // Get entropy and recreate wallet
  const entropy = await snap.request({
    method: 'snap_getBip44Entropy',
    params: { coinType: 8888 }
  });

  const privateKey = DilithiumUtils.hexToKey(entropy.privateKey.slice(2, 66));
  const wallet = await QuantumWallet.fromPrivateKey(privateKey, account.algorithm);

  // Sign transaction
  const signedTx = await wallet.signTransaction(params.transaction);

  return {
    signature: signedTx.signature,
    publicKey: signedTx.publicKey,
    algorithm: signedTx.signatureAlgorithm,
    signedTransaction: signedTx
  };
}

/**
 * Get quantum algorithm information
 */
async function getAlgorithmInfo(): Promise<{
  supported: Array<{
    algorithm: SignatureAlgorithm;
    name: string;
    signatureSize: number;
    publicKeySize: number;
    securityLevel: number;
  }>;
}> {
  return {
    supported: [
      {
        algorithm: SignatureAlgorithm.Dilithium,
        name: 'CRYSTALS-Dilithium-II',
        signatureSize: 2420,
        publicKeySize: 1312,
        securityLevel: 128
      },
      {
        algorithm: SignatureAlgorithm.Falcon,
        name: 'FALCON-512',
        signatureSize: 690,
        publicKeySize: 897,
        securityLevel: 103
      }
    ]
  };
}

/**
 * Export account (encrypted)
 */
async function exportQuantumAccount(params: { address: string; password: string }): Promise<{
  encryptedAccount: string;
  algorithm: SignatureAlgorithm;
}> {
  const state = await getSnapState();
  const account = state.accounts[params.address];
  
  if (!account) {
    throw new Error('Account not found');
  }

  // Show export warning
  const confirmed = await snap.request({
    method: 'snap_dialog',
    params: {
      type: 'confirmation',
      content: panel([
        heading('⚠️ Export Quantum Account'),
        text(`Account: ${params.address}`),
        text(`Algorithm: ${getAlgorithmName(account.algorithm)}`),
        divider(),
        text('🔐 This will export your private key encrypted with the provided password.'),
        text('⚠️ Keep your password safe! You will need it to import the account.'),
        text('🚨 Never share your exported account with anyone!')
      ])
    }
  });

  if (!confirmed) {
    throw new Error('Account export cancelled by user');
  }

  // Get entropy and recreate wallet
  const entropy = await snap.request({
    method: 'snap_getBip44Entropy',
    params: { coinType: 8888 }
  });

  const privateKey = DilithiumUtils.hexToKey(entropy.privateKey.slice(2, 66));
  const wallet = await QuantumWallet.fromPrivateKey(privateKey, account.algorithm);

  // Export wallet (encrypted)
  const encryptedAccount = wallet.exportWallet(params.password);

  return {
    encryptedAccount,
    algorithm: account.algorithm
  };
}

/**
 * Import account from private key
 */
async function importQuantumAccount(params: { 
  privateKey: string; 
  algorithm?: SignatureAlgorithm;
}): Promise<{
  address: string;
  publicKey: string;
  algorithm: SignatureAlgorithm;
}> {
  const algorithm = params.algorithm || SignatureAlgorithm.Dilithium;

  // Show import confirmation
  const confirmed = await snap.request({
    method: 'snap_dialog',
    params: {
      type: 'confirmation',
      content: panel([
        heading('📥 Import Quantum Account'),
        text(`Algorithm: ${getAlgorithmName(algorithm)}`),
        divider(),
        text('🔐 This will import a quantum-resistant account from a private key.'),
        text('⚠️ Make sure you trust the source of this private key.'),
        text('🗑️ Delete the private key from its original location after import.')
      ])
    }
  });

  if (!confirmed) {
    throw new Error('Account import cancelled by user');
  }

  try {
    // Create wallet from private key
    const privateKeyBytes = DilithiumUtils.hexToKey(params.privateKey);
    const wallet = await QuantumWallet.fromPrivateKey(privateKeyBytes, algorithm);
    
    const address = wallet.getAddress();
    const publicKey = DilithiumUtils.keyToHex(wallet.getPublicKey());

    // Store account in snap state
    const state = await getSnapState();
    state.accounts[address] = {
      address,
      publicKey,
      algorithm,
      created: Date.now()
    };
    
    await setSnapState(state);

    // Show success dialog
    await snap.request({
      method: 'snap_dialog',
      params: {
        type: 'alert',
        content: panel([
          heading('✅ Quantum Account Imported'),
          text(`Address: ${address}`),
          text(`Algorithm: ${getAlgorithmName(algorithm)}`),
          divider(),
          text('🎉 Your quantum account is ready to use!')
        ])
      }
    });

    return { address, publicKey, algorithm };
  } catch (error: any) {
    throw new Error(`Import failed: ${error.message}`);
  }
}

// Helper functions

async function getSnapState(): Promise<QuantumSnapState> {
  const state = await snap.request({
    method: 'snap_manageState',
    params: { operation: 'get' }
  });
  
  return state || { accounts: {} };
}

async function setSnapState(state: QuantumSnapState): Promise<void> {
  await snap.request({
    method: 'snap_manageState',
    params: { operation: 'update', newState: state }
  });
}

function getAlgorithmName(algorithm: SignatureAlgorithm): string {
  const names = {
    [SignatureAlgorithm.Dilithium]: 'CRYSTALS-Dilithium-II',
    [SignatureAlgorithm.Falcon]: 'FALCON-512'
  };
  return names[algorithm] || 'Unknown';
}

// Global error handler
process.on('unhandledRejection', (error) => {
  console.error('Unhandled promise rejection in Quantum Snap:', error);
});