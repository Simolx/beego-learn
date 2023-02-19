package utils

import (
	"log"
	"os"
)

func WriteFile(path string, content []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("create or open file %q failed, error %v", path, err)
		return err
	}
	defer f.Close()
	if _, err := f.Write(content); err != nil {
		log.Printf("write file %q failed, error %v", path, err)
		return err
	}
	return nil
}
