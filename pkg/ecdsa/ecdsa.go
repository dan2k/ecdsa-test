package ecdsa  

import(
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"math/big"
	"log"
)

func Sign(data string) (string, string, string, error) {
	// Generate a new ECDSA key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
		return "", "", "", err
	}
	// Get the private key bytes.
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		log.Fatal(err)
		return "", "", "", err
	}
	// สามารถนำ privateKeyHex นี้เก็บลงฐานได้
	privateKeyHex := hex.EncodeToString(privateKeyBytes)
	/*
			fmt.Println("privatekeyHex:",privateKeyHex)

			// เป็นขั้นตอนการแปลง privateKey กลับมาให้ตามเดิม
			privateKeyBytes,err=hex.DecodeString(privateKeyHex)
			if err != nil {
				log.Fatal(err)
			}
			// Unmarshal the private key.
		    	privateKey, err = x509.ParseECPrivateKey(privateKeyBytes)
		    	if err != nil {
		        	log.Fatal(err)
		    	}
	*/

	// Sign the data
	hashed := sha256.Sum256([]byte(data))
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hashed[:])
	if err != nil {
		log.Fatal(err)
		return "", "", "", err
	}
	// Convert the signature to hex string
	signatureHex := hex.EncodeToString(append(r.Bytes(), s.Bytes()...))
	// Convert the public key to hex string
	publicKeyHex := hex.EncodeToString(append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...))
	return signatureHex, publicKeyHex, privateKeyHex, nil
}
func Verify(data, signatureHex, publicKeyHex string) (bool, error) {
	// Convert signature and public key from hex to byte slices
	signature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false, err
	}
	publicKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return false, err
	}
	// Create an ecdsa.PublicKey from the byte slices
	publicKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(publicKeyBytes[:32]),
		Y:     new(big.Int).SetBytes(publicKeyBytes[32:]),
	}
	// Verify the signature
	hashed := sha256.Sum256([]byte(data))
	valid := ecdsa.Verify(publicKey, hashed[:], new(big.Int).SetBytes(signature[:32]), new(big.Int).SetBytes(signature[32:]))
	return valid, nil
}