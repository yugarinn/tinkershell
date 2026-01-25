# Tinkershell
A command line utility to remotely execute code on deployed Laravel instances.

Inspired by [Tinkerwell](https://tinkerwell.app/) and motivated by Electron's clunkiness.

``` text
⚠️ This is in early alpha testing and should not be used in production environments, end user discretion is advised.
```

## Installation

## Usage
``` shell
tinkershell -e production -f processPendingPayments.php
```

## Configuration
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
