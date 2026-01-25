# Tinkershell
A command line utility to remotely execute code on deployed Laravel instances through a SSH connection.
Inspired by [Tinkerwell](https://tinkerwell.app/) and motivated by Electron's clunkiness.

``` text
⚠️ This is in early alpha testing and should not be used in production
environments, end user discretion is advised.
```

## Installation

## Configuration
Tinkershell expects a [toml](https://toml.io/en/) config file to exists in `~/.config/tinkershell/tinkershell.toml`. The config file organizes connection environments under different sections. All connection environments must include four fields:

- `ip_address`: The ipv4 address of the remote server.
- `ssh_username`: The username to be used in the SSH connection.
- `ssh_public_key`: The path to the SSH public key to be used.
- `project_path`: The root path of the Laravel project in the remote server.

**Example**
``` toml
[production]
ip_address = "34.12.123.12"
ssh_username = "yugarinn"
ssh_public_key = "~/.ssh/id_rsa"
project_path = "/home/yugarinn/laravel-app/"

[staging]
ip_address = "129.11.98.32"
ssh_username = "yugarinn"
ssh_public_key = "~/.ssh/id_rsa"
project_path = "/home/yugarinn/laravel-app/"
```

## Usage
The executable expects two flags to be provided, `-e` one of the configured sections in the config file, and `-f` the path to the `.php` file to be executed.

``` shell
tinkershell -e production -f processPendingPayments.php
```

All `tinkershell` executions are assigned a pseudo-unique ID that is used in two files, the default remote server log file and a local log file under `~/.config/tinkershell/logs/`.
