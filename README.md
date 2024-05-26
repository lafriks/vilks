# Vilks - Red team simulation tool

## Installation

Clone repository and build with command:

```sh
go build -ldflags="-s -w" -o vilks cmd/runner/*.go
```

## Usage

```console
Usage:
   [flags]
   [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  exec        Execute
  help        Help about any command
  validate    Validate

Flags:
  -d, --debug     Enable debug mode
  -h, --help      help for this command
  -v, --version   version for this command
```

### Validate scenario

```console
Usage:
   vils validate [flags]

Flags:
  -h, --help              help for validate
  -r, --recipes string    Path to recipes directory
  -s, --scenario string   Path to scenario file
```

### Execute scenario

```console
Usage:
   vilks exec [flags]

Flags:
      --attack string     Attack name
  -a, --attacker string   Attacker IP address
  -e, --evidence string   Path to evidence directory
  -h, --help              help for exec
      --host string       Host name
  -r, --recipes string    Path to recipes directory
  -s, --scenario string   Path to scenario file
      --team string       Team name
```
