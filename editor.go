package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Editor interface {
	save() (*string, error)
}

type VSCode struct {
	Dir string
}

// mdファイルを保存し、ファイルの内容を返す
func (v VSCode) save() (*string, error) {
	now := time.Now().Format(time.RFC3339)
	contentFile, err := os.Create(filepath.Join(v.Dir, fmt.Sprintf(`%s.md`, now)))
	if err != nil {
		return nil, err
	}
	defer contentFile.Close()

	cmd := exec.Command("code", "--wait", contentFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile(contentFile.Name())
	if err != nil {
		return nil, err
	}
	content := string(bytes)
	return &content, nil
}
