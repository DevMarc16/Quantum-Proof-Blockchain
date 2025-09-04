# 🚀 Enterprise Quantum Blockchain Deployment Guide

## Overview

This guide provides complete instructions for deploying the enterprise-grade quantum-resistant blockchain infrastructure. All critical production components have been implemented and tested.

## ✅ Implemented Components

### 1. Hardware Security Module (HSM) Integration
- **Location**: `/chain/security/hsm/`
- **Features**: 
  - FIPS 140-2 Level 3/4 compliant key management
  - AWS CloudHSM provider implementation
  - Automated key rotation and backup
  - Secure validator key storage
  - Emergency recovery procedures

### 2. Kubernetes Production Infrastructure
- **Location**: `/k8s/`
- **Features**:
  - Complete StatefulSet for multi-validator deployment
  - Auto-scaling configuration
  - High availability (99.9% uptime) setup
  - Resource quotas and security policies
  - Persistent volume claims for blockchain data

### 3. Comprehensive Monitoring & Alerting
- **Location**: `/k8s/monitoring/`
- **Features**:
  - Prometheus metrics collection
  - Grafana dashboards with quantum-specific metrics
  - AlertManager configuration
  - Custom quantum blockchain alerts
  - Performance anomaly detection

### 4. JavaScript/TypeScript SDK
- **Location**: `/sdk/js/`
- **Features**:
  - Complete quantum transaction support
  - CRYSTALS-Dilithium-II integration
  - Web3.js compatibility layer
  - TypeScript definitions
  - Comprehensive examples and documentation

### 5. MetaMask Integration
- **Location**: `/integrations/metamask/`
- **Features**:
  - MetaMask Snap for quantum signatures
  - Post-quantum key management
  - Quantum transaction signing UI
  - Account import/export functionality

## 🔧 Deployment Instructions

### Prerequisites

1. **Kubernetes Cluster** (v1.24+)
   - AWS EKS, GCP GKE, or Azure AKS recommended
   - Minimum 3 nodes with 4 CPU, 16GB RAM each
   - 500GB SSD storage per validator node

2. **HSM Setup** (Production only)
   - AWS CloudHSM cluster configured
   - HSM credentials and certificates
   - VPN access to HSM network

3. **Domain & SSL**
   - Domain names for RPC endpoints
   - Valid SSL certificates
   - Load balancer configuration

### Step 1: Infrastructure Deployment

```bash
# 1. Create namespace and base resources
kubectl apply -f k8s/base/namespace.yaml
kubectl apply -f k8s/base/configmap.yaml

# 2. Create secrets (replace with actual values)
kubectl create secret generic quantum-hsm-credentials \
  --namespace quantum-blockchain \
  --from-literal=aws-access-key-id=YOUR_ACCESS_KEY \
  --from-literal=aws-secret-access-key=YOUR_SECRET_KEY \
  --from-literal=aws-region=us-west-2

kubectl create secret generic quantum-hsm-certificates \
  --namespace quantum-blockchain \
  --from-file=ca.crt=path/to/ca.crt \
  --from-file=client.crt=path/to/client.crt \
  --from-file=client.key=path/to/client.key

kubectl create secret generic grafana-credentials \
  --namespace quantum-blockchain \
  --from-literal=admin-password=SECURE_ADMIN_PASSWORD

# 3. Deploy storage class (adjust for your cloud provider)
cat <<EOF | kubectl apply -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: quantum-ssd-retain
provisioner: kubernetes.io/aws-ebs  # or gce-pd, azure-disk
parameters:
  type: gp3  # or pd-ssd, Premium_LRS
  fsType: ext4
reclaimPolicy: Retain
allowVolumeExpansion: true
volumeBindingMode: WaitForFirstConsumer
EOF

# 4. Deploy validators
kubectl apply -f k8s/validators/validator-statefulset.yaml

# 5. Deploy monitoring stack
kubectl apply -f k8s/monitoring/prometheus.yaml
```

### Step 2: Monitoring Setup

```bash
# Create monitoring configurations
kubectl create configmap grafana-config \
  --namespace quantum-blockchain \
  --from-file=grafana-config.yaml

kubectl create configmap alertmanager-config \
  --namespace quantum-blockchain \
  --from-literal=alertmanager.yml="$(cat <<EOF
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alerts@quantum-blockchain.org'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
- name: 'web.hook'
  email_configs:
  - to: 'admin@quantum-blockchain.org'
    subject: 'Quantum Blockchain Alert: {{ .GroupLabels.alertname }}'
    body: |
      {{ range .Alerts }}
      Alert: {{ .Annotations.summary }}
      Description: {{ .Annotations.description }}
      {{ end }}

- name: 'slack'
  slack_configs:
  - api_url: 'YOUR_SLACK_WEBHOOK_URL'
    channel: '#quantum-alerts'
    title: 'Quantum Blockchain Alert'
    text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
EOF
)"
```

### Step 3: Network Configuration

```bash
# Configure ingress for external access
cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: quantum-blockchain-ingress
  namespace: quantum-blockchain
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - rpc.quantum-blockchain.org
    - monitoring.quantum-blockchain.org
    secretName: quantum-tls
  rules:
  - host: rpc.quantum-blockchain.org
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: quantum-validator-rpc
            port:
              number: 8545
  - host: monitoring.quantum-blockchain.org
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: grafana
            port:
              number: 3000
EOF
```

### Step 4: HSM Integration (Production)

```bash
# Initialize HSM manager
go run cmd/hsm-setup/main.go --config=hsm-config.yaml

# Create validator keys
go run cmd/create-validator-keys/main.go \
  --validators=validator-001,validator-002,validator-003 \
  --hsm-provider=aws-cloudhsm \
  --backup-location=s3://quantum-blockchain-backups/keys
```

### Step 5: Verification

```bash
# Check deployment status
kubectl get pods -n quantum-blockchain
kubectl get services -n quantum-blockchain
kubectl get ingress -n quantum-blockchain

# Test RPC endpoints
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  https://rpc.quantum-blockchain.org

# Check validator synchronization
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"quantum_getMetrics","params":[],"id":1}' \
  https://rpc.quantum-blockchain.org

# Access monitoring dashboard
open https://monitoring.quantum-blockchain.org
```

## 📊 Monitoring & Maintenance

### Key Metrics to Monitor

1. **Blockchain Health**
   - Block production rate (target: 2 seconds)
   - Transaction throughput (>500 TPS)
   - Network sync status

2. **Quantum Security**
   - Signature verification success rate (>99.99%)
   - HSM connectivity status
   - Key rotation schedule compliance

3. **Infrastructure**
   - Pod memory/CPU usage
   - Storage utilization
   - Network latency between validators

### Alert Thresholds

```yaml
# Critical Alerts (immediate response)
- Validator down >30 seconds
- Block production stopped >1 minute  
- HSM connectivity lost
- Memory usage >90%

# Warning Alerts (investigate within 1 hour)
- Block production slow (>3 seconds)
- High transaction queue (>1000 pending)
- Storage usage >80%
- Network partition detected
```

### Maintenance Procedures

1. **Daily**
   - Check validator sync status
   - Monitor alert notifications
   - Review transaction throughput

2. **Weekly**
   - Analyze performance trends
   - Check HSM audit logs
   - Update monitoring dashboards

3. **Monthly**
   - Review key rotation policies
   - Test disaster recovery procedures
   - Security audit review

## 🔒 Security Best Practices

### Network Security
- All traffic encrypted with TLS 1.3
- VPN access for validator management
- Firewall rules limiting P2P ports
- DDoS protection enabled

### Key Management  
- HSM keys never leave secure hardware
- Multi-party approval for key operations
- Regular key rotation (90 days)
- Secure backup procedures

### Access Control
- Role-based access control (RBAC)
- Multi-factor authentication required
- Audit logging for all operations
- Regular access review

## 🚀 SDK Usage Examples

### JavaScript/TypeScript Integration

```javascript
const { QuickStart, QuantumWallet, SignatureAlgorithm } = require('@quantum-blockchain/sdk');

// Connect to quantum blockchain
const provider = await QuickStart.connectMainnet();

// Create quantum wallet
const wallet = await QuantumWallet.random(
  SignatureAlgorithm.Dilithium, 
  provider
);

// Sign and send transaction
const txHash = await wallet.sendTransaction({
  to: '0x742d35cc6269c4c2a4e8d8c6f0e1c8f8a2b4c6a8',
  value: '1000000000000000000', // 1 QTM
  gasLimit: '0x5208'
});

console.log('Transaction sent:', txHash);
```

### MetaMask Snap Integration

```javascript
// Install quantum snap
await ethereum.request({
  method: 'wallet_requestSnaps',
  params: {
    '@quantum-blockchain/metamask-snap': {}
  }
});

// Create quantum account
const account = await ethereum.request({
  method: 'wallet_invokeSnap',
  params: {
    snapId: '@quantum-blockchain/metamask-snap',
    request: {
      method: 'quantum_createAccount'
    }
  }
});

// Sign quantum message
const signature = await ethereum.request({
  method: 'wallet_invokeSnap',
  params: {
    snapId: '@quantum-blockchain/metamask-snap',
    request: {
      method: 'quantum_signMessage',
      params: {
        address: account.address,
        message: 'Hello Quantum World!'
      }
    }
  }
});
```

## 📈 Performance Expectations

### Network Performance
- **Block Time**: 2 seconds consistent
- **TPS**: 500+ transactions per second  
- **Finality**: 12 seconds (6 blocks)
- **Uptime**: 99.9% availability

### Quantum Operations
- **Key Generation**: 10-50ms
- **Signature Creation**: 5-20ms
- **Signature Verification**: 2-10ms
- **Signature Size**: 2420 bytes (Dilithium)

### Resource Usage
- **CPU**: 2-4 cores per validator
- **Memory**: 4-8GB per validator
- **Storage**: 100GB initial, 10GB/month growth
- **Network**: 100Mbps per validator

## 🆘 Troubleshooting

### Common Issues

1. **Validator Not Syncing**
   ```bash
   kubectl logs quantum-validator-0 -n quantum-blockchain
   kubectl describe pod quantum-validator-0 -n quantum-blockchain
   ```

2. **HSM Connection Issues**
   ```bash
   kubectl exec -it quantum-validator-0 -n quantum-blockchain -- \
     /quantum-node hsm-test --provider=aws-cloudhsm
   ```

3. **High Memory Usage**
   ```bash
   kubectl top pods -n quantum-blockchain
   kubectl get events -n quantum-blockchain --sort-by='.lastTimestamp'
   ```

### Support Contacts

- **Technical Support**: support@quantum-blockchain.org
- **Security Issues**: security@quantum-blockchain.org  
- **Infrastructure**: infrastructure@quantum-blockchain.org
- **Emergency**: +1-555-QUANTUM (24/7)

---

## 🎯 Success Criteria

Your quantum blockchain deployment is considered successful when:

✅ All 3 validators producing blocks every 2 seconds  
✅ Quantum signatures verified on every transaction  
✅ HSM integration operational with key rotation  
✅ Monitoring dashboards showing green status  
✅ External RPC endpoints responding <100ms  
✅ SDK integration working in test applications  
✅ MetaMask snap creating quantum accounts  
✅ 99.9% uptime maintained for 30 days  

**Congratulations! You now have a production-ready enterprise quantum blockchain platform! 🎉**