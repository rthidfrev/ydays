package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"os"
)

// func main() {
// 	decryptFile(os.Args[1])
// }

func decryptFile(key string) {
	enc, err := os.ReadFile("./assets/crypt.txt")
	if err != nil {
		fmt.Println("Erreur lors de la lecture du fichier chiffré:", err)
		return
	}

	// Décoder la clé (premières 64 caractères hexadécimaux)
	decodedKey, err := hex.DecodeString(key) // La clé fait 64 caractères hexadécimaux
	if err != nil {
		fmt.Println("Erreur lors du décodage de la clé:", err)
		return
	}

	// Décoder le texte chiffré (reste des caractères hexadécimaux)
	decodedCiphertext, err := hex.DecodeString(string(enc))
	if err != nil {
		fmt.Println("Erreur lors du décodage du texte chiffré:", err)
		return
	}

	// Créer le bloc AES avec la clé décodée
	block, err := aes.NewCipher(decodedKey)
	if err != nil {
		fmt.Println("Erreur lors de la création du bloc AES avec la clé:", err)
		return
	}

	// Initialisation de GCM avec le bloc AES
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Println("Erreur lors de l'initialisation du mode GCM pour le déchiffrement:", err)
		return
	}

	// Extraction du nonce : nous devons savoir où commence le nonce
	nonceSize := gcm.NonceSize()
	if len(decodedCiphertext) < nonceSize {
		fmt.Println("Erreur : les données chiffrées sont trop petites pour contenir un nonce valide.")
		return
	}
	nonce, ciphertext := decodedCiphertext[:nonceSize], decodedCiphertext[nonceSize:]

	// Déchiffrement
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		fmt.Println("Erreur lors du déchiffrement des données:", err)
		return
	}

	fmt.Println("Données déchiffrées:", string(plaintext))
}

// Fonction pour vérifier si une chaîne est valide en hexadécimal
func isHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}
