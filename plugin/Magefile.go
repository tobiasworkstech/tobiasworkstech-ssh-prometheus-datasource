//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Default target to run when none is specified
var Default = Build.All

type Build mg.Namespace

// All builds frontend and backend
func (Build) All() error {
	mg.Deps(Build.Frontend, Build.Backend)
	return nil
}

// Frontend builds the frontend
func (Build) Frontend() error {
	return sh.RunV("npm", "run", "build")
}

// Backend builds the backend for the current OS
func (Build) Backend() error {
	return buildBackend(runtime.GOOS, runtime.GOARCH)
}

// Linux builds the backend for Linux
func (Build) Linux() error {
	return buildBackend("linux", "amd64")
}

// LinuxARM64 builds the backend for Linux ARM64
func (Build) LinuxARM64() error {
	return buildBackend("linux", "arm64")
}

// Darwin builds the backend for macOS
func (Build) Darwin() error {
	return buildBackend("darwin", "amd64")
}

// DarwinARM64 builds the backend for macOS ARM64
func (Build) DarwinARM64() error {
	return buildBackend("darwin", "arm64")
}

// Windows builds the backend for Windows
func (Build) Windows() error {
	return buildBackend("windows", "amd64")
}

func buildBackend(goos, goarch string) error {
	exeName := "gpx_ssh_prometheus_datasource"
	if goos == "windows" {
		exeName += ".exe"
	}

	exeName = fmt.Sprintf("%s_%s_%s", exeName, goos, goarch)
	if goos == "windows" {
		exeName = fmt.Sprintf("gpx_ssh_prometheus_datasource_%s_%s.exe", goos, goarch)
	}

	outputPath := filepath.Join("dist", exeName)

	env := map[string]string{
		"GOOS":        goos,
		"GOARCH":      goarch,
		"CGO_ENABLED": "0",
	}

	ldflags := "-w -s"

	fmt.Printf("Building %s/%s...\n", goos, goarch)

	cmd := exec.Command("go", "build",
		"-ldflags", ldflags,
		"-o", outputPath,
		"./pkg",
	)
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

type Test mg.Namespace

// Backend runs Go tests
func (Test) Backend() error {
	return sh.RunV("go", "test", "-v", "./...")
}

// Frontend runs frontend tests
func (Test) Frontend() error {
	return sh.RunV("npm", "test")
}

type Dev mg.Namespace

// Watch runs frontend in watch mode
func (Dev) Watch() error {
	return sh.RunV("npm", "run", "dev")
}

// Clean removes build artifacts
func Clean() error {
	return os.RemoveAll("dist")
}
