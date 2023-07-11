package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func writeDebug(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(v); err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}

	return nil
}
