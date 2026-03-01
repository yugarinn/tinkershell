<?php

if (!defined('STDOUT')) define('STDOUT', fopen('php://stdout', 'w'));
if (!defined('STDERR')) define('STDERR', fopen('php://stderr', 'w'));
if (!defined('STDIN')) define('STDIN', fopen('php://stdin', 'w'));

$executionId = '{{.ExecutionID}}';
$silent = filter_var('{{.SilentMode}}', FILTER_VALIDATE_BOOLEAN);

if (! class_exists('Tinkershell')) {
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

require '{{.LaravelPath}}/vendor/autoload.php';

$app = require_once '{{.LaravelPath}}/bootstrap/app.php';
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
    $classMapPath = '{{.LaravelPath}}/vendor/composer/autoload_classmap.php';

    if (file_exists($classMapPath)) {
        \Laravel\Tinker\ClassAliasAutoloader::register($shell, $classMapPath);
    }
}

$pid = getmypid();
if (! $silent) {
    Tinkershell::log("[Tinkershell INFO] running process '{$executionId}' [PID: {$pid}]...");
}

try {
    $shell->execute(<<<'TINKERSHELL'
{{.UserCode}}
TINKERSHELL
    );
} catch (\Throwable $e) {
    fwrite(STDERR, "Execution Error: " . $e->getMessage() . PHP_EOL);
    if (! $silent) {
        Tinkershell::log("[Tinkershell INFO] running process '{$executionId}' [PID: {$pid}]... done");
    }

    exit(1);
}

if (! $silent) {
    Tinkershell::log("[Tinkershell INFO] running process '{$executionId}' [PID: {$pid}]... done");
}
