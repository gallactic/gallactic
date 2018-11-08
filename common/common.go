package common

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

func RandomHex(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func WriteFile(filename string, data []byte) error {
	// create directory
	if err := Mkdir(filepath.Dir(filename)); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filename, data, 0777); err != nil {
		return fmt.Errorf("Failed to write to %s: %v", filename, err)
	}
	return nil
}

func Mkdir(dir string) error {
	// create the directory
	if err := os.MkdirAll(dir, 0777); err != nil {
		return fmt.Errorf("Could not create directory %s", dir)
	}
	return nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func GallacticHomeDir() string {
	home := ""
	usr, err := user.Current()
	if err == nil {
		home = usr.HomeDir + "/gallactic/"
	}
	return home
}

func GallacticKeystoreDir() string {
	return GallacticHomeDir() + "/keystore/"
}
