/**
 * Quantum Smart Contract Interaction
 * Provides quantum-safe contract deployment and interaction
 */

import { EventEmitter } from 'events';
import { 
  QuantumContractABI, 
  QuantumContract as IQuantumContract,
  QuantumTransactionRequest,
  QuantumEventLog
} from '../types/quantum';
import { QuantumProvider } from '../provider/quantum-provider';
import { QuantumWallet } from '../wallet/quantum-wallet';

export class QuantumContract extends EventEmitter implements IQuantumContract {
  public readonly address: string;
  public readonly abi: QuantumContractABI[];
  public readonly methods: Record<string, any> = {};
  public readonly events: Record<string, any> = {};

  private provider: QuantumProvider;
  private wallet?: QuantumWallet;

  constructor(
    address: string,
    abi: QuantumContractABI[],
    provider: QuantumProvider,
    wallet?: QuantumWallet
  ) {
    super();
    this.address = address;
    this.abi = abi;
    this.provider = provider;
    this.wallet = wallet;

    this.initializeMethods();
    this.initializeEvents();
  }

  /**
   * Deploy new quantum contract
   */
  static async deploy(
    abi: QuantumContractABI[],
    bytecode: string,
    constructorArgs: any[],
    wallet: QuantumWallet,
    options: {
      gasLimit?: string;
      gasPrice?: string;
      value?: string;
    } = {}
  ): Promise<QuantumContract> {
    const provider = wallet['provider'] as QuantumProvider;
    if (!provider) {
      throw new Error('Wallet must have provider configured');
    }

    // Encode constructor arguments
    const constructorData = QuantumContractUtils.encodeConstructor(abi, constructorArgs);
    const deploymentData = bytecode + constructorData;

    // Create deployment transaction
    const deployTx: QuantumTransactionRequest = {
      data: deploymentData,
      gasLimit: options.gasLimit || '0x1e8480', // 2M gas
      gasPrice: options.gasPrice,
      value: options.value || '0x0'
    };

    // Send deployment transaction
    const txHash = await wallet.sendTransaction(deployTx);
    
    // Wait for deployment confirmation
    const receipt = await provider.getTransactionReceipt(txHash);
    if (!receipt?.contractAddress) {
      throw new Error('Contract deployment failed');
    }

    return new QuantumContract(receipt.contractAddress, abi, provider, wallet);
  }

  /**
   * Call contract method (read-only)
   */
  async call(methodName: string, args: any[] = [], options: {
    from?: string;
    blockTag?: string;
  } = {}): Promise<any> {
    const method = this.abi.find(item => item.name === methodName && item.type === 'function');
    if (!method) {
      throw new Error(`Method ${methodName} not found in ABI`);
    }

    if (method.stateMutability !== 'pure' && method.stateMutability !== 'view') {
      throw new Error(`Method ${methodName} is not a read-only method`);
    }

    const callData = QuantumContractUtils.encodeFunctionCall(method, args);
    
    const result = await this.provider.request({
      method: 'eth_call',
      params: [{
        to: this.address,
        data: callData,
        from: options.from || (this.wallet ? this.wallet.getAddress() : undefined)
      }, options.blockTag || 'latest']
    });

    return QuantumContractUtils.decodeFunctionResult(method, result);
  }

  /**
   * Send transaction to contract method
   */
  async send(methodName: string, args: any[] = [], options: {
    gasLimit?: string;
    gasPrice?: string;
    value?: string;
  } = {}): Promise<string> {
    if (!this.wallet) {
      throw new Error('Wallet required for sending transactions');
    }

    const method = this.abi.find(item => item.name === methodName && item.type === 'function');
    if (!method) {
      throw new Error(`Method ${methodName} not found in ABI`);
    }

    if (method.stateMutability === 'pure' || method.stateMutability === 'view') {
      throw new Error(`Method ${methodName} is read-only, use call() instead`);
    }

    const callData = QuantumContractUtils.encodeFunctionCall(method, args);
    
    const txRequest: QuantumTransactionRequest = {
      to: this.address,
      data: callData,
      gasLimit: options.gasLimit,
      gasPrice: options.gasPrice,
      value: options.value || '0x0'
    };

    return await this.wallet.sendTransaction(txRequest);
  }

  /**
   * Estimate gas for contract method call
   */
  async estimateGas(methodName: string, args: any[] = [], options: {
    from?: string;
    value?: string;
  } = {}): Promise<string> {
    const method = this.abi.find(item => item.name === methodName && item.type === 'function');
    if (!method) {
      throw new Error(`Method ${methodName} not found in ABI`);
    }

    const callData = QuantumContractUtils.encodeFunctionCall(method, args);
    
    return await this.provider.estimateGas({
      to: this.address,
      data: callData,
      from: options.from || (this.wallet ? this.wallet.getAddress() : undefined),
      value: options.value || '0x0'
    } as any);
  }

  /**
   * Get past events
   */
  async getPastEvents(eventName: string, options: {
    fromBlock?: string | number;
    toBlock?: string | number;
    filter?: Record<string, any>;
  } = {}): Promise<QuantumEventLog[]> {
    const event = this.abi.find(item => item.name === eventName && item.type === 'event');
    if (!event) {
      throw new Error(`Event ${eventName} not found in ABI`);
    }

    const topics = [QuantumContractUtils.getEventTopic(event)];
    
    // Add filter topics
    if (options.filter) {
      const filterTopics = QuantumContractUtils.encodeEventFilter(event, options.filter);
      topics.push(...filterTopics);
    }

    const logs = await this.provider.request({
      method: 'eth_getLogs',
      params: [{
        address: this.address,
        topics,
        fromBlock: typeof options.fromBlock === 'number' 
          ? `0x${options.fromBlock.toString(16)}` 
          : options.fromBlock || '0x0',
        toBlock: typeof options.toBlock === 'number'
          ? `0x${options.toBlock.toString(16)}`
          : options.toBlock || 'latest'
      }]
    });

    return logs.map((log: any) => ({
      ...log,
      decodedData: QuantumContractUtils.decodeEventLog(event, log.data, log.topics)
    }));
  }

  /**
   * Subscribe to contract events
   */
  subscribeToEvents(eventName: string, callback: (event: any) => void): void {
    const event = this.abi.find(item => item.name === eventName && item.type === 'event');
    if (!event) {
      throw new Error(`Event ${eventName} not found in ABI`);
    }

    // Subscribe to logs for this contract and event
    this.provider.subscribe('logs', (log: QuantumEventLog) => {
      if (log.address.toLowerCase() === this.address.toLowerCase()) {
        const eventTopic = QuantumContractUtils.getEventTopic(event);
        if (log.topics[0] === eventTopic) {
          const decodedEvent = QuantumContractUtils.decodeEventLog(event, log.data, log.topics);
          callback({ ...log, decodedData: decodedEvent });
        }
      }
    });
  }

  /**
   * Get contract code
   */
  async getCode(): Promise<string> {
    return await this.provider.request({
      method: 'eth_getCode',
      params: [this.address, 'latest']
    });
  }

  /**
   * Check if contract is quantum-safe
   */
  async isQuantumSafe(): Promise<boolean> {
    // Check if contract implements quantum-safe interface
    const code = await this.getCode();
    if (code === '0x') return false;

    // Look for quantum-safe methods in ABI
    return this.abi.some(item => item.quantumSafe === true);
  }

  // Private initialization methods

  private initializeMethods(): void {
    this.abi
      .filter(item => item.type === 'function')
      .forEach(method => {
        const isReadOnly = method.stateMutability === 'pure' || method.stateMutability === 'view';
        
        this.methods[method.name] = (...args: any[]) => {
          const lastArg = args[args.length - 1];
          const options = (typeof lastArg === 'object' && !Array.isArray(lastArg)) ? args.pop() : {};
          
          return isReadOnly 
            ? this.call(method.name, args, options)
            : this.send(method.name, args, options);
        };
      });
  }

  private initializeEvents(): void {
    this.abi
      .filter(item => item.type === 'event')
      .forEach(event => {
        this.events[event.name] = {
          getPastEvents: (options: any) => this.getPastEvents(event.name, options),
          subscribe: (callback: Function) => this.subscribeToEvents(event.name, callback)
        };
      });
  }
}

/**
 * Quantum Contract Utilities
 */
export const QuantumContractUtils = {
  /**
   * Encode function call data
   */
  encodeFunctionCall(method: QuantumContractABI, args: any[]): string {
    // Simplified encoding - in production use proper ABI encoding
    const selector = this.getFunctionSelector(method);
    const encodedArgs = this.encodeParameters(method.inputs, args);
    return selector + encodedArgs.slice(2); // Remove 0x prefix from args
  },

  /**
   * Get function selector (first 4 bytes of function signature hash)
   */
  getFunctionSelector(method: QuantumContractABI): string {
    const signature = this.getFunctionSignature(method);
    const { keccak256 } = require('js-sha3');
    const hash = keccak256(signature);
    return '0x' + hash.slice(0, 8); // First 4 bytes as hex
  },

  /**
   * Get function signature string
   */
  getFunctionSignature(method: QuantumContractABI): string {
    const inputs = method.inputs.map(input => input.type).join(',');
    return `${method.name}(${inputs})`;
  },

  /**
   * Encode constructor arguments
   */
  encodeConstructor(abi: QuantumContractABI[], args: any[]): string {
    const constructor = abi.find(item => item.type === 'constructor');
    if (!constructor) return '';
    
    return this.encodeParameters(constructor.inputs, args);
  },

  /**
   * Encode parameters (simplified)
   */
  encodeParameters(inputs: any[], values: any[]): string {
    if (inputs.length !== values.length) {
      throw new Error('Parameter count mismatch');
    }
    
    // Simplified encoding - in production use ethers.js ABI coder
    let encoded = '0x';
    for (let i = 0; i < values.length; i++) {
      const type = inputs[i].type;
      const value = values[i];
      
      if (type === 'uint256' || type === 'uint') {
        const num = BigInt(value);
        encoded += num.toString(16).padStart(64, '0');
      } else if (type === 'address') {
        const addr = value.startsWith('0x') ? value.slice(2) : value;
        encoded += addr.padStart(64, '0');
      } else if (type === 'string') {
        const utf8 = new TextEncoder().encode(value);
        const length = utf8.length.toString(16).padStart(64, '0');
        const data = Array.from(utf8).map(b => b.toString(16).padStart(2, '0')).join('');
        encoded += length + data.padEnd(Math.ceil(data.length / 64) * 64, '0');
      } else {
        // Default: assume it's hex data
        encoded += value.toString().padStart(64, '0');
      }
    }
    
    return encoded;
  },

  /**
   * Decode function result
   */
  decodeFunctionResult(method: QuantumContractABI, result: string): any {
    if (!method.outputs || method.outputs.length === 0) return null;
    if (method.outputs.length === 1) {
      return this.decodeParameter(method.outputs[0].type, result);
    }
    
    // Multiple outputs - return array
    const values = [];
    let offset = 2; // Skip 0x
    
    for (const output of method.outputs) {
      const value = this.decodeParameter(output.type, '0x' + result.slice(offset, offset + 64));
      values.push(value);
      offset += 64;
    }
    
    return values;
  },

  /**
   * Decode single parameter
   */
  decodeParameter(type: string, data: string): any {
    if (type === 'uint256' || type === 'uint') {
      return BigInt('0x' + data.slice(2)).toString();
    } else if (type === 'address') {
      return '0x' + data.slice(-40);
    } else if (type === 'bool') {
      return data.slice(-1) === '1';
    } else {
      return data; // Return raw data
    }
  },

  /**
   * Get event topic (event signature hash)
   */
  getEventTopic(event: QuantumContractABI): string {
    const signature = this.getEventSignature(event);
    const { keccak256 } = require('js-sha3');
    return '0x' + keccak256(signature);
  },

  /**
   * Get event signature string
   */
  getEventSignature(event: QuantumContractABI): string {
    const inputs = event.inputs.map(input => input.type).join(',');
    return `${event.name}(${inputs})`;
  },

  /**
   * Encode event filter topics
   */
  encodeEventFilter(event: QuantumContractABI, filter: Record<string, any>): string[] {
    const topics = [];
    
    for (const input of event.inputs) {
      if (input.indexed && filter[input.name] !== undefined) {
        const value = filter[input.name];
        if (input.type === 'address') {
          topics.push('0x' + value.slice(2).padStart(64, '0'));
        } else if (input.type === 'uint256' || input.type === 'uint') {
          topics.push('0x' + BigInt(value).toString(16).padStart(64, '0'));
        } else {
          topics.push(value);
        }
      } else {
        topics.push(null); // Wildcard
      }
    }
    
    return topics;
  },

  /**
   * Decode event log data
   */
  decodeEventLog(event: QuantumContractABI, data: string, topics: string[]): Record<string, any> {
    const result: Record<string, any> = {};
    let topicIndex = 1; // Skip event signature topic
    let dataOffset = 2; // Skip 0x
    
    for (const input of event.inputs) {
      if (input.indexed) {
        // Indexed parameters are in topics
        result[input.name] = this.decodeParameter(input.type, topics[topicIndex]);
        topicIndex++;
      } else {
        // Non-indexed parameters are in data
        const paramData = '0x' + data.slice(dataOffset, dataOffset + 64);
        result[input.name] = this.decodeParameter(input.type, paramData);
        dataOffset += 64;
      }
    }
    
    return result;
  }
};