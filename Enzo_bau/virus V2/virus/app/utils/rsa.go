package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"log"
)

// Fonction pour chiffrer un fichier
func EncryptWithPublicKey(plaintext string, publicKeyPath string) (string, error) {
	// Lire la clé publique depuis le fichier
	publicKeyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return "", err
	}

	// Décoder la clé publique
	publicKeyBlock, _ := pem.Decode(publicKeyBytes)
	if publicKeyBlock == nil || publicKeyBlock.Type != "PUBLIC KEY" {
		return "", errors.New("clé publique invalide")
	}

	publicKeyInterface, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return "", err
	}

	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("clé publique invalide")
	}

	// Chiffrer le texte
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, []byte(plaintext), nil)
	if err != nil {
		return "", err
	}

	// Encoder le texte chiffré en base64
	encodedCiphertext := base64.StdEncoding.EncodeToString(ciphertext)

	log.Println("encrypt")
	return encodedCiphertext, nil
}
