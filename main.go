package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const VERSION = "v0.0.7"

//go:embed tinkershell.php
var tinkershellTemplate string

func main() {
	envName := flag.String("e", "", "The environment to run against (e.g., production, staging)")
	filePath := flag.String("f", "", "The path to the PHP file to execute")
	silentMode := flag.Bool("s", false, "Disables all output and logging")
	showVersion := flag.Bool("version", false, "Show the current version of tinkershell")

	flag.Parse()

	if *showVersion {
		fmt.Printf("tinkershell version %s\n", VERSION)
		os.Exit(0)
	}

	if *envName == "" || *filePath == "" {
		fmt.Println("Usage: tinkershell [-s] -e <environment> -f <file.php>")
		flag.PrintDefaults()

		os.Exit(1)
	}

	config := loadConfig()
	connectionConfig := config[*envName]

	requiredFields := []string{"ip_address", "ssh_username", "ssh_public_key", "project_path"}
	for _, requiredField := range requiredFields {
		if connectionConfig[requiredField] == "" {
			panic(fmt.Sprintf("missing required field in config: '%s'", requiredField))
		}
	}

	host := fmt.Sprintf("%s@%s", connectionConfig["ssh_username"], connectionConfig["ip_address"])
	executionID := generateExecutionID()

	run(executionID, host, silentMode, connectionConfig["ssh_public_key"], connectionConfig["project_path"], *filePath)
}

func generateExecutionID() string {
	milli := time.Now().UnixMilli()
	milliStr := strconv.FormatInt(milli, 10)
	executionID := base64.StdEncoding.EncodeToString([]byte(milliStr))

	return executionID
}

func loadConfig() map[string]map[string]string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	configPath := filepath.Join(home, ".config", "tinkershell", "tinkershell.toml")
	configFile, err := os.Open(configPath)
	if err != nil {
		panic(err)
	}
	defer configFile.Close()

	config := make(map[string]map[string]string)
	currentSection := "root"
	config[currentSection] = make(map[string]string)

	scanner := bufio.NewScanner(configFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			if _, exists := config[currentSection]; !exists {
				config[currentSection] = make(map[string]string)
			}

			continue
		}

		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			value = strings.Trim(value, "\"")
			config[currentSection][key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return config
}

func prepare(executionID string, silentMode *bool, userCode string, laravelProjectPath string) (string, error) {
	tmpl, err := template.New("php").Parse(tinkershellTemplate)
	if err != nil {
		return "", err
	}

	variables := struct {
		ExecutionID string
		SilentMode  *bool
		UserCode    string
		LaravelPath string
	}{
		ExecutionID: executionID,
		SilentMode:  silentMode,
		UserCode:    userCode,
		LaravelPath: laravelProjectPath,
	}

	var phpScript bytes.Buffer
	if err := tmpl.Execute(&phpScript, variables); err != nil {
		return "", err
	}

	return phpScript.String(), nil
}

func run(executionID string, host string, silentMode *bool, publicKeyPath string, projectPath string, localFile string) {
	home, _ := os.UserHomeDir()
	logDir := filepath.Join(home, ".config", "tinkershell", "logs")

	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Warning: Could not create log directory: %v\n", err)
	}

	logFilename := generateLogFilename(executionID)
	logFilePath := filepath.Join(logDir, logFilename)
	logFile, _ := os.Create(logFilePath)
	defer logFile.Close()

	userCode, _ := os.ReadFile(localFile)
	payload, err := prepare(executionID, silentMode, stripPHPOpeningTag(string(userCode)), projectPath)
	if err != nil {
		panic(fmt.Sprintf("error while preparing script for execution: '%s'", err.Error()))
	}

	cmd := exec.Command("ssh", "-t", "-q", host, "-i", publicKeyPath, "php")

	cmd.Stdin = strings.NewReader(payload)

	if !*silentMode {
		writer := io.MultiWriter(os.Stdout, logFile)
		cmd.Stdout = writer
		cmd.Stderr = writer
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Remote execution failed: %v\n", err)
	}
}

func generateLogFilename(executionID string) string {
	timestamp := time.Now().Format("20060102-150405")

	return fmt.Sprintf("%s-%s.log", timestamp, executionID)
}

func stripPHPOpeningTag(code string) string {
	code = strings.TrimSpace(code)
	code = strings.TrimPrefix(code, "<?php")
	return strings.TrimSpace(code)
}
