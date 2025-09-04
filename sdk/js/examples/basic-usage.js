/**
 * Basic Quantum Blockchain SDK Usage Examples
 * Demonstrates core functionality for developers
 */

const { 
  QuickStart, 
  QuantumWallet, 
  QuantumContract, 
  SignatureAlgorithm 
} = require('@quantum-blockchain/sdk');

async function basicUsageDemo() {
  console.log('🚀 Quantum Blockchain SDK Demo');
  
  try {
    // 1. Connect to local quantum blockchain
    console.log('\n📡 Connecting to local quantum node...');
    const provider = await QuickStart.connectLocal();
    
    // Check quantum compatibility
    const isQuantumCompatible = await provider.isQuantumCompatible();
    console.log(`✅ Quantum compatibility: ${isQuantumCompatible}`);
    
    // Get network info
    const blockNumber = await provider.getBlockNumber();
    console.log(`📊 Current block: ${blockNumber}`);
    
    // 2. Create quantum wallet
    console.log('\n🔑 Creating quantum wallet...');
    const wallet = await QuantumWallet.random(SignatureAlgorithm.Dilithium, provider);
    
    console.log(`Address: ${wallet.getAddress()}`);
    console.log(`Algorithm: ${wallet.getSignatureAlgorithm()}`);
    console.log(`Public Key: ${wallet.getPublicKey().length} bytes`);
    
    // 3. Check balance
    console.log('\n💰 Checking wallet balance...');
    const balance = await wallet.getBalance();
    console.log(`Balance: ${balance} wei`);
    
    // 4. Sign a message
    console.log('\n✍️ Signing quantum message...');
    const message = "Hello Quantum Blockchain!";
    const signature = await wallet.signMessage(message);
    
    console.log(`Message: "${message}"`);
    console.log(`Signature size: ${signature.signature.length} bytes`);
    console.log(`Algorithm: ${signature.algorithm}`);
    
    // 5. Verify signature
    console.log('\n🔍 Verifying quantum signature...');
    const isValid = await wallet.verifySignature(message, signature);
    console.log(`Signature valid: ${isValid}`);
    
    // 6. Create and sign transaction
    if (balance !== '0') {
      console.log('\n📝 Creating quantum transaction...');
      
      const txRequest = {
        to: '0x742d35cc6269c4c2a4e8d8c6f0e1c8f8a2b4c6a8',
        value: '1000000000000000000', // 1 ETH in wei
        gasLimit: '0x5208', // 21000 gas
        gasPrice: '0x3b9aca00' // 1 Gwei
      };
      
      const signedTx = await wallet.signTransaction(txRequest);
      console.log(`Transaction signed with ${signedTx.signatureAlgorithm} algorithm`);
      console.log(`Signature: ${signedTx.signature.length} bytes`);
      
      // Note: Don't actually send the transaction in demo
      console.log('📤 Transaction ready to send (not sent in demo)');
    }
    
    // 7. Get quantum metrics
    console.log('\n📈 Quantum network metrics...');
    const metrics = await provider.getQuantumMetrics();
    console.log(`Block time: ${metrics.blockTime}s`);
    console.log(`Validators: ${metrics.validators}`);
    console.log(`Block height: ${metrics.blockHeight}`);
    
    // 8. Export wallet
    console.log('\n💾 Exporting wallet...');
    const walletJson = wallet.exportWallet('demo-password');
    console.log(`Wallet exported (${walletJson.length} chars)`);
    
    console.log('\n✅ Demo completed successfully!');
    
  } catch (error) {
    console.error('❌ Demo failed:', error.message);
  }
}

async function contractInteractionDemo() {
  console.log('\n🔗 Contract Interaction Demo');
  
  try {
    const provider = await QuickStart.connectLocal();
    const wallet = await QuantumWallet.random(SignatureAlgorithm.Dilithium, provider);
    
    // Sample quantum-safe contract ABI
    const quantumTokenABI = [
      {
        name: 'transfer',
        type: 'function',
        inputs: [
          { name: 'to', type: 'address' },
          { name: 'amount', type: 'uint256' }
        ],
        outputs: [{ name: '', type: 'bool' }],
        stateMutability: 'nonpayable',
        quantumSafe: true
      },
      {
        name: 'balanceOf',
        type: 'function',
        inputs: [{ name: 'owner', type: 'address' }],
        outputs: [{ name: '', type: 'uint256' }],
        stateMutability: 'view',
        quantumSafe: true
      },
      {
        name: 'Transfer',
        type: 'event',
        inputs: [
          { name: 'from', type: 'address', indexed: true },
          { name: 'to', type: 'address', indexed: true },
          { name: 'value', type: 'uint256', indexed: false }
        ]
      }
    ];
    
    // Create contract instance (using mock address)
    const contractAddress = '0x1234567890123456789012345678901234567890';
    const contract = new QuantumContract(contractAddress, quantumTokenABI, provider, wallet);
    
    console.log(`📋 Contract Address: ${contract.address}`);
    console.log(`🔒 Quantum Safe: ${await contract.isQuantumSafe()}`);
    
    // Demonstrate method calls (would require actual deployed contract)
    console.log('📞 Contract methods available:', Object.keys(contract.methods));
    console.log('📡 Contract events available:', Object.keys(contract.events));
    
    // Estimate gas for quantum transaction
    try {
      const gasEstimate = await contract.estimateGas('transfer', [
        '0x742d35cc6269c4c2a4e8d8c6f0e1c8f8a2b4c6a8',
        '1000000000000000000'
      ]);
      console.log(`⛽ Gas estimate: ${gasEstimate}`);
    } catch (error) {
      console.log('⚠️ Gas estimation failed (contract not deployed)');
    }
    
    console.log('✅ Contract interaction demo completed');
    
  } catch (error) {
    console.error('❌ Contract demo failed:', error.message);
  }
}

async function performanceBenchmark() {
  console.log('\n⚡ Performance Benchmark');
  
  try {
    const provider = await QuickStart.connectLocal();
    
    // Benchmark key generation
    console.log('\n🔑 Benchmarking key generation...');
    const keyGenStart = Date.now();
    const wallet = await QuantumWallet.random(SignatureAlgorithm.Dilithium, provider);
    const keyGenTime = Date.now() - keyGenStart;
    console.log(`Key generation: ${keyGenTime}ms`);
    
    // Benchmark signing
    console.log('\n✍️ Benchmarking signing...');
    const message = 'Benchmark message for quantum signing performance test';
    const signingTimes = [];
    
    for (let i = 0; i < 5; i++) {
      const start = Date.now();
      await wallet.signMessage(message + i);
      signingTimes.push(Date.now() - start);
    }
    
    const avgSigningTime = signingTimes.reduce((a, b) => a + b) / signingTimes.length;
    console.log(`Average signing time: ${avgSigningTime.toFixed(2)}ms`);
    console.log(`Signing times: ${signingTimes.join(', ')}ms`);
    
    // Benchmark verification
    console.log('\n🔍 Benchmarking verification...');
    const signature = await wallet.signMessage(message);
    const verifyTimes = [];
    
    for (let i = 0; i < 5; i++) {
      const start = Date.now();
      await wallet.verifySignature(message, signature);
      verifyTimes.push(Date.now() - start);
    }
    
    const avgVerifyTime = verifyTimes.reduce((a, b) => a + b) / verifyTimes.length;
    console.log(`Average verification time: ${avgVerifyTime.toFixed(2)}ms`);
    console.log(`Verification times: ${verifyTimes.join(', ')}ms`);
    
    // Calculate throughput
    const signaturesPerSecond = 1000 / avgSigningTime;
    const verificationsPerSecond = 1000 / avgVerifyTime;
    
    console.log(`\n📊 Performance Summary:`);
    console.log(`Signatures/second: ${signaturesPerSecond.toFixed(2)}`);
    console.log(`Verifications/second: ${verificationsPerSecond.toFixed(2)}`);
    
  } catch (error) {
    console.error('❌ Benchmark failed:', error.message);
  }
}

// Run demos
async function runAllDemos() {
  await basicUsageDemo();
  await contractInteractionDemo(); 
  await performanceBenchmark();
  
  console.log('\n🎉 All demos completed!');
  process.exit(0);
}

// Run if called directly
if (require.main === module) {
  runAllDemos().catch(console.error);
}

module.exports = {
  basicUsageDemo,
  contractInteractionDemo,
  performanceBenchmark
};