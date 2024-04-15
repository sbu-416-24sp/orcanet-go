package hash

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	Price     float64 `json:"price"`
	Timestamp string  `json:"timestamp"`
	Uuid      string  `json:"uuid"`
}

func GeneratePriceBytes(price float64) []byte {
	timestamp := time.Now()
	timestampStr := timestamp.Format(time.RFC3339)
	uuidObj, err := uuid.NewRandom()
	if err != nil {
		fmt.Println("Error generating UUID:", err)
		return nil
	}

	// Convert the UUID to a string
	uuidStr := uuidObj.String()
	jsonData := Transaction{
		Price:     price,
		Timestamp: timestampStr,
		Uuid:      uuidStr,
	}
	// Marshal JSON object to byte array
	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return nil
	}
	return jsonBytes
}
func OpenTransactionFile(message []byte, pubKey *rsa.PublicKey) {

}
func GenerateKeyPair() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 4096)
}
func SignFile(file []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	hashed := sha256.Sum256(file)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, err
	}
	return signature, nil
}
func VerifySignature(file []byte, signature []byte, publicKey *rsa.PublicKey) error {
	if publicKey == nil {
		return errors.New("null private key")
	}
	hashed := sha256.Sum256(file)
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature)
}

func ExportRsaPrivateKeyAsPemStr(privkey *rsa.PrivateKey) []byte {
	privkey_bytes := x509.MarshalPKCS1PrivateKey(privkey)
	privkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privkey_bytes,
		},
	)
	return privkey_pem
}

func ParseRsaPrivateKeyFromPemStr(privPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

func ExportRsaPublicKeyAsPemStr(pubkey *rsa.PublicKey) ([]byte, error) {
	pubkey_bytes, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		return nil, err
	}
	pubkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubkey_bytes,
		},
	)

	return pubkey_pem, nil
}

func ParseRsaPublicKeyFromPemStr(pubPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		break // fall through
	}
	return nil, errors.New("key type is not RSA")
}

func LoadInKeys() (*rsa.PublicKey, *rsa.PrivateKey) {
	fileContent, err := os.ReadFile("install.sh")
	var privateKey *rsa.PrivateKey
	var publicKey *rsa.PublicKey
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}
	_, err1 := os.Stat("./config/key.pub")
	_, err2 := os.Stat("./config/key.priv")
	if err1 == nil && err2 == nil {
		fmt.Printf("Loading in public/private key locally...\n")
		privateKeyContent, err := os.ReadFile("./config/key.priv")
		if err != nil {
			fmt.Println("Error loading in key file:", err)
			os.Exit(1)
		}
		publicKeyContent, err := os.ReadFile("./config/key.pub")
		if err != nil {
			fmt.Println("Error loading in key file:", err)
			os.Exit(1)
		}
		privateKey, err = ParseRsaPrivateKeyFromPemStr(string(privateKeyContent))
		if err != nil {
			fmt.Println("Error loading in key file:", err)
			os.Exit(1)
		}
		publicKey, err = ParseRsaPublicKeyFromPemStr(string(publicKeyContent))
		if err != nil {
			fmt.Println("Error loading in key file:", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Public/Private key does not exist, generating...\n")
		privateKey, err = GenerateKeyPair()
		if err != nil {
			fmt.Println("Error generating key pair:", err)
			os.Exit(1)
		}

		publicKey = &privateKey.PublicKey
		pubBytes, err := ExportRsaPublicKeyAsPemStr(publicKey)
		if err != nil {
			fmt.Println("Error generating public key as PEM str:", err)
			os.Exit(1)
		}
		os.WriteFile("./config/key.pub", pubBytes, 0644)

		privBytes := ExportRsaPrivateKeyAsPemStr(privateKey)
		if err != nil {
			fmt.Println("Error generating public key as PEM str:", err)
			os.Exit(1)
		}
		os.WriteFile("./config/key.priv", privBytes, 0644)
	}

	// Sign file
	signedFile, err := SignFile(fileContent, privateKey)
	if err != nil {
		fmt.Println("Error signing file:", err)
		os.Exit(1)
	}

	// Verify signature
	err = VerifySignature(fileContent, signedFile, publicKey)

	if err != nil {
		fmt.Println("Unable to verify key pair, please try again:", err)
		os.Exit(1)
	} else {
		fmt.Println("Public/Private key verified.")
	}
	return publicKey, privateKey
}
