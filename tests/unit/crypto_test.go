package unit

import (
	"bytes"
	"testing"

	"quantum-blockchain/chain/crypto"
)

func TestDilithiumKeyGeneration(t *testing.T) {
	privKey, pubKey, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate Dilithium key pair: %v", err)
	}

	if len(privKey.Bytes()) != crypto.DilithiumPrivateKeySize {
		t.Errorf("Expected private key size %d, got %d", crypto.DilithiumPrivateKeySize, len(privKey.Bytes()))
	}

	if len(pubKey.Bytes()) != crypto.DilithiumPublicKeySize {
		t.Errorf("Expected public key size %d, got %d", crypto.DilithiumPublicKeySize, len(pubKey.Bytes()))
	}
}

func TestDilithiumSigningAndVerification(t *testing.T) {
	privKey, pubKey, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate Dilithium key pair: %v", err)
	}

	message := []byte("Hello, Quantum World!")

	// Sign the message
	signature, err := privKey.Sign(message)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	if len(signature) != crypto.DilithiumSignatureSize {
		t.Errorf("Expected signature size %d, got %d", crypto.DilithiumSignatureSize, len(signature))
	}

	// Verify the signature
	valid := pubKey.Verify(message, signature)
	if !valid {
		t.Error("Signature verification failed")
	}

	// Test with wrong message
	wrongMessage := []byte("Wrong message")
	validWrong := pubKey.Verify(wrongMessage, signature)
	if validWrong {
		t.Error("Signature verification should have failed for wrong message")
	}

	// Test with corrupted signature
	corruptedSignature := make([]byte, len(signature))
	copy(corruptedSignature, signature)
	corruptedSignature[0] ^= 0xFF
	validCorrupted := pubKey.Verify(message, corruptedSignature)
	if validCorrupted {
		t.Error("Signature verification should have failed for corrupted signature")
	}
}

func TestFalconKeyGeneration(t *testing.T) {
	privKey, pubKey, err := crypto.GenerateFalconKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate Falcon key pair: %v", err)
	}

	if len(privKey.Bytes()) != crypto.FalconPrivateKeySize {
		t.Errorf("Expected private key size %d, got %d", crypto.FalconPrivateKeySize, len(privKey.Bytes()))
	}

	if len(pubKey.Bytes()) != crypto.FalconPublicKeySize {
		t.Errorf("Expected public key size %d, got %d", crypto.FalconPublicKeySize, len(pubKey.Bytes()))
	}
}

func TestFalconSigningAndVerification(t *testing.T) {
	privKey, pubKey, err := crypto.GenerateFalconKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate Falcon key pair: %v", err)
	}

	message := []byte("Hello, Quantum World!")

	// Sign the message
	signature, err := privKey.Sign(message)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	if len(signature) > crypto.FalconSignatureSize {
		t.Errorf("Signature too large: expected max %d, got %d", crypto.FalconSignatureSize, len(signature))
	}

	// Verify the signature
	valid := pubKey.Verify(message, signature)
	if !valid {
		t.Error("Signature verification failed")
	}

	// Test with wrong message
	wrongMessage := []byte("Wrong message")
	validWrong := pubKey.Verify(wrongMessage, signature)
	if validWrong {
		t.Error("Signature verification should have failed for wrong message")
	}
}

func TestKyberKeyGeneration(t *testing.T) {
	privKey, pubKey, err := crypto.GenerateKyberKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate Kyber key pair: %v", err)
	}

	if len(privKey.Bytes()) != crypto.KyberPrivateKeySize {
		t.Errorf("Expected private key size %d, got %d", crypto.KyberPrivateKeySize, len(privKey.Bytes()))
	}

	if len(pubKey.Bytes()) != crypto.KyberPublicKeySize {
		t.Errorf("Expected public key size %d, got %d", crypto.KyberPublicKeySize, len(pubKey.Bytes()))
	}
}

func TestKyberEncapsulationAndDecapsulation(t *testing.T) {
	privKey, pubKey, err := crypto.GenerateKyberKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate Kyber key pair: %v", err)
	}

	// Encapsulate
	ciphertext, sharedSecret1, err := pubKey.Encapsulate()
	if err != nil {
		t.Fatalf("Failed to encapsulate: %v", err)
	}

	if len(ciphertext) != crypto.KyberCiphertextSize {
		t.Errorf("Expected ciphertext size %d, got %d", crypto.KyberCiphertextSize, len(ciphertext))
	}

	if len(sharedSecret1) != crypto.KyberSharedSecretSize {
		t.Errorf("Expected shared secret size %d, got %d", crypto.KyberSharedSecretSize, len(sharedSecret1))
	}

	// Decapsulate
	sharedSecret2, err := privKey.Decapsulate(ciphertext)
	if err != nil {
		t.Fatalf("Failed to decapsulate: %v", err)
	}

	// Verify shared secrets match
	if !bytes.Equal(sharedSecret1, sharedSecret2) {
		t.Error("Shared secrets don't match")
	}

	// Test with wrong ciphertext
	wrongCiphertext := make([]byte, len(ciphertext))
	copy(wrongCiphertext, ciphertext)
	wrongCiphertext[0] ^= 0xFF

	_, err = privKey.Decapsulate(wrongCiphertext)
	if err == nil {
		t.Log("Decapsulation should have failed with wrong ciphertext")
	}
}

func TestQuantumSignatureInterface(t *testing.T) {
	// Test Dilithium
	privKey, _, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate Dilithium key pair: %v", err)
	}

	message := []byte("Test message")

	qrSig, err := crypto.SignMessage(message, crypto.SigAlgDilithium, privKey.Bytes())
	if err != nil {
		t.Fatalf("Failed to sign with Dilithium: %v", err)
	}

	if qrSig.Algorithm != crypto.SigAlgDilithium {
		t.Errorf("Expected algorithm %v, got %v", crypto.SigAlgDilithium, qrSig.Algorithm)
	}

	valid, err := crypto.VerifySignature(message, qrSig)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}

	if !valid {
		t.Error("Signature verification failed")
	}

	// Test Falcon
	privKeyFalcon, _, err := crypto.GenerateFalconKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate Falcon key pair: %v", err)
	}

	qrSigFalcon, err := crypto.SignMessage(message, crypto.SigAlgFalcon, privKeyFalcon.Bytes())
	if err != nil {
		t.Fatalf("Failed to sign with Falcon: %v", err)
	}

	if qrSigFalcon.Algorithm != crypto.SigAlgFalcon {
		t.Errorf("Expected algorithm %v, got %v", crypto.SigAlgFalcon, qrSigFalcon.Algorithm)
	}

	validFalcon, err := crypto.VerifySignature(message, qrSigFalcon)
	if err != nil {
		t.Fatalf("Failed to verify Falcon signature: %v", err)
	}

	if !validFalcon {
		t.Error("Falcon signature verification failed")
	}
}

func TestAlgorithmInfo(t *testing.T) {
	// Test Dilithium
	pubKeySize, err := crypto.GetPublicKeySize(crypto.SigAlgDilithium)
	if err != nil {
		t.Fatalf("Failed to get Dilithium public key size: %v", err)
	}
	if pubKeySize != crypto.DilithiumPublicKeySize {
		t.Errorf("Expected public key size %d, got %d", crypto.DilithiumPublicKeySize, pubKeySize)
	}

	sigSize, err := crypto.GetSignatureSize(crypto.SigAlgDilithium)
	if err != nil {
		t.Fatalf("Failed to get Dilithium signature size: %v", err)
	}
	if sigSize != crypto.DilithiumSignatureSize {
		t.Errorf("Expected signature size %d, got %d", crypto.DilithiumSignatureSize, sigSize)
	}

	privKeySize, err := crypto.GetPrivateKeySize(crypto.SigAlgDilithium)
	if err != nil {
		t.Fatalf("Failed to get Dilithium private key size: %v", err)
	}
	if privKeySize != crypto.DilithiumPrivateKeySize {
		t.Errorf("Expected private key size %d, got %d", crypto.DilithiumPrivateKeySize, privKeySize)
	}

	// Test Falcon
	pubKeySizeFalcon, err := crypto.GetPublicKeySize(crypto.SigAlgFalcon)
	if err != nil {
		t.Fatalf("Failed to get Falcon public key size: %v", err)
	}
	if pubKeySizeFalcon != crypto.FalconPublicKeySize {
		t.Errorf("Expected public key size %d, got %d", crypto.FalconPublicKeySize, pubKeySizeFalcon)
	}

	// Test unsupported algorithm
	_, err = crypto.GetPublicKeySize(crypto.SignatureAlgorithm(99))
	if err == nil {
		t.Error("Should have failed for unsupported algorithm")
	}
}

func TestAlgorithmStrings(t *testing.T) {
	if crypto.SigAlgDilithium.String() != "Dilithium" {
		t.Errorf("Expected 'Dilithium', got '%s'", crypto.SigAlgDilithium.String())
	}

	if crypto.SigAlgFalcon.String() != "Falcon" {
		t.Errorf("Expected 'Falcon', got '%s'", crypto.SigAlgFalcon.String())
	}

	if crypto.SigAlgSPHINCS.String() != "SPHINCS+" {
		t.Errorf("Expected 'SPHINCS+', got '%s'", crypto.SigAlgSPHINCS.String())
	}

	unknown := crypto.SignatureAlgorithm(99)
	if unknown.String() != "Unknown" {
		t.Errorf("Expected 'Unknown', got '%s'", unknown.String())
	}
}

func BenchmarkDilithiumKeyGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := crypto.GenerateDilithiumKeyPair()
		if err != nil {
			b.Fatalf("Key generation failed: %v", err)
		}
	}
}

func BenchmarkDilithiumSigning(b *testing.B) {
	privKey, _, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		b.Fatalf("Key generation failed: %v", err)
	}

	message := []byte("Benchmark message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := privKey.Sign(message)
		if err != nil {
			b.Fatalf("Signing failed: %v", err)
		}
	}
}

func BenchmarkDilithiumVerification(b *testing.B) {
	privKey, pubKey, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		b.Fatalf("Key generation failed: %v", err)
	}

	message := []byte("Benchmark message")
	signature, err := privKey.Sign(message)
	if err != nil {
		b.Fatalf("Signing failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		valid := pubKey.Verify(message, signature)
		if !valid {
			b.Fatal("Verification failed")
		}
	}
}
