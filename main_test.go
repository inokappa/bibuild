package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestHelpFlag(t *testing.T) {
	cmd := exec.Command("./debug.sh", "-help")
	stdout := new(bytes.Buffer)
	cmd.Stderr = stdout
	msg := "ソースファイルの ZIP 圧縮を行う."

	_ = cmd.Run()

	if !strings.Contains(stdout.String(), msg) {
		t.Fatal("Failed Test")
	}
}

func TestVersionFlag(t *testing.T) {
	cmd := exec.Command("./debug.sh", "-version")
	stdout := new(bytes.Buffer)
	cmd.Stdout = stdout

	_ = cmd.Run()

	if !strings.Contains(stdout.String(), appVersion) {
		t.Fatal("Failed Test")
	}
}
