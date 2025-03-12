package main

import (
	"fmt"
)

func main() {
	var maxSize int64 = 1 << 40 // 1 Go de données (ajuster si besoin)

	s := make([]byte, maxSize) // Crée un slice de 1 Go
	for i := range s {
		s[i] = 'a' // Remplit avec des 'a'
	}

	str := string(s) // Convertit en string
	fmt.Println("Taille du string:", len(str))
}
