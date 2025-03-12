// package assets

// import (
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"os"
// 	"strings"
// 	"time"
// )

// // Structure pour stocker les données reçues
// type Data struct {
// 	Name  string `json:"name"`
// 	Email string `json:"email"`
// 	Age   int    `json:"age"`
// }

// // Handler pour gérer les requêtes POST contenant un fichier et des données
// func uploadHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	err := r.ParseMultipartForm(10 << 20) // Limite de 10 Mo
// 	if err != nil {
// 		http.Error(w, "Erreur lors du parsing du formulaire", http.StatusBadRequest)
// 		return
// 	}

// 	// Récupérer le fichier
// 	file, header, err := r.FormFile("file")
// 	if err != nil {
// 		http.Error(w, "Aucun fichier reçu", http.StatusBadRequest)
// 		return
// 	}
// 	defer file.Close()

// 	// Générer un nom de fichier valide
// 	t := time.Now()
// 	name := t.Format("2006-01-02_15-04-05") + " - " + header.Filename
// 	name = strings.ReplaceAll(name, ":", "-") // Remplacer les caractères interdits

// 	// Créer le fichier dans le dossier uploads/
// 	dst, err := os.Create("./uploads/" + name)
// 	if err != nil {
// 		http.Error(w, "Erreur lors de la sauvegarde du fichier", http.StatusInternalServerError)
// 		return
// 	}
// 	defer dst.Close()

// 	// Copier le fichier reçu
// 	_, err = io.Copy(dst, file)
// 	if err != nil {
// 		http.Error(w, "Erreur lors de l'écriture du fichier", http.StatusInternalServerError)
// 		return
// 	}

// 	// Réponse JSON
// 	w.Header().Set("Content-Type", "application/json")
// 	fmt.Fprintf(w, `{"message": "Fichier enregistré", "file": "%s"}`, name)
// }

// func StartWeb() {
// 	http.HandleFunc("/upload", uploadHandler)

// 	fmt.Println("Serveur démarré sur : http://localhost:8080")
// 	if err := http.ListenAndServe(":8080", nil); err != nil {
// 		fmt.Println("Erreur lors du démarrage du serveur :", err)
// 	}
// }
