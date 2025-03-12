package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func GetMasterKey() string {

	// Get the master key
	// The master key is the key with which chrome encode the passwords but it has some suffixes and we need to work on it
	jsonFile, err := os.Open(AllData().LocalStatePath) // The rough key is stored in the Local State File which is a json file
	if err != nil {
		return "error key"
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return "error key"
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)
	roughKey := result["os_crypt"].(map[string]interface{})["encrypted_key"].(string) // Found parsing the json in it
	decodedKey, err := base64.StdEncoding.DecodeString(roughKey)                      // It's stored in Base64 so.. Let's decode it
	if err != nil {
		return "error key"
	}

	stringKey := string(decodedKey)
	stringKey = strings.Trim(stringKey, "DPAPI") // The key is encrypted using the windows DPAPI method and signed with it. the key looks like "DPAPI05546sdf879z456..." Let's Remove DPAPI.
	data := AllData()
	data.KeyG = stringKey

	// Decrypt the key using the dllcrypt32 dll.
	fmt.Println("Key : ", data.KeyG)

	return data.KeyG
}
