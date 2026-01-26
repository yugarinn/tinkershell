package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const VERSION = "v0.0.3"

func main() {
	envName := flag.String("e", "", "The environment to run against (e.g., production, staging)")
	filePath := flag.String("f", "", "The path to the PHP file to execute")
	showVersion := flag.Bool("version", false, "Show the current version of tinkershell")

	flag.Parse()

	if *showVersion {
		fmt.Printf("tinkershell version %s\n", VERSION)
		os.Exit(0)
	}

	if *envName == "" || *filePath == "" {
		fmt.Println("Usage: tinkershell -e <environment> -f <file.php>")
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

	hostElements := []string{connectionConfig["ssh_username"], connectionConfig["ip_address"]}
	host := strings.Join(hostElements, "@")

	executionID := generateExecutionID()
	run(executionID, host, connectionConfig["project_path"], *filePath)
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

func prepare(executionID string, userCode string, laravelProjectPath string) string {
	return fmt.Sprintf(`<?php

if (!defined('STDOUT')) define('STDOUT', fopen('php://stdout', 'w'));
if (!defined('STDERR')) define('STDERR', fopen('php://stderr', 'w'));
if (!defined('STDIN')) define('STDIN', fopen('php://stdin', 'w'));

$executionId = '%s';

if (!class_exists('Tinkershell')) {
    final class Tinkershell
    {
        public static function log(mixed $log = null): void
        {
            $log = (string) $log;

            echo $log . "\n";
            Log::info($log);
        }
    }
}

require '%s/vendor/autoload.php';

$app = require_once '%s/bootstrap/app.php';
$kernel = $app->make(Illuminate\Contracts\Console\Kernel::class);
$kernel->bootstrap();

\Symfony\Component\VarDumper\VarDumper::setHandler(function ($var) {
    $dumper = new \Symfony\Component\VarDumper\Dumper\CliDumper(STDOUT, null, \Symfony\Component\VarDumper\Dumper\CliDumper::DUMP_LIGHT_ARRAY);

    $dumper->setDisplayOptions(['display_source' => false]);
    $dumper->dump((new \Symfony\Component\VarDumper\Cloner\VarCloner())->cloneVar($var));
});

$config = new \Psy\Configuration([
    'updateCheck' => 'never',
    'usePcntl'    => false,
    'configFile'  => null,
]);

$output = new \Psy\Output\ShellOutput();
$shell = new \Psy\Shell($config);

$shell->setScopeVariables(['app' => $app]);
$shell->setOutput($output);

\Illuminate\Foundation\AliasLoader::getInstance($app->make('config')->get('app.aliases'))->register();

if (class_exists(\Laravel\Tinker\ClassAliasAutoloader::class)) {
    $classMapPath = '%s/vendor/composer/autoload_classmap.php';

    if (file_exists($classMapPath)) {
        \Laravel\Tinker\ClassAliasAutoloader::register($shell, $classMapPath);
    }
}

$pid = getmypid();
Tinkershell::log("[Tinkershell INFO] running process '{$executionId}' [PID: {$pid}]...");

try {
    $shell->execute(<<<'TINKERSHELL'
%s
TINKERSHELL
    );
} catch (\Throwable $e) {
    fwrite(STDERR, "Execution Error: " . $e->getMessage() . PHP_EOL);
	Tinkershell::log("[Tinkershell INFO] running process '{$executionId}' [PID: {$pid}]... done");

    exit(1);
}

Tinkershell::log("[Tinkershell INFO] running process '{$executionId}' [PID: {$pid}]... done");

`, executionID, laravelProjectPath, laravelProjectPath, laravelProjectPath, userCode)
}

func run(executionID string, host string, projectPath string, localFile string) {
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
	payload := prepare(executionID, string(userCode), projectPath)
	cmd := exec.Command("ssh", "-t", "-q", host, "php")

	writer := io.MultiWriter(os.Stdout, logFile)

	cmd.Stdin = strings.NewReader(payload)
	cmd.Stdout = writer
	cmd.Stderr = writer

	if err := cmd.Run(); err != nil {
		fmt.Printf("Remote execution failed: %v\n", err)
	}
}

func generateLogFilename(executionID string) string {
	timestamp := time.Now().Format("20060102-150405")

	return fmt.Sprintf("%s-%s.log", timestamp, executionID)
}
