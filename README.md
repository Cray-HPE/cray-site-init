# Cray Site Initializer (`csi`)

[![Continuous Integration](https://github.com/Cray-HPE/cray-site-init/actions/workflows/ci.yml/badge.svg)](https://github.com/Cray-HPE/cray-site-init/actions/workflows/ci.yml)

`csi` is a tool for facilitating the installation of an HPCaaS cluster.

> **`NOTE`** **This deprecates CrayCTL** (`crayctl`) from Shasta V1.4.0 and higher as the primary orchestrator tool.

# Getting Started

See https://cray-hpe.github.io/cray-site-init/ and follow the Site Survey directions.

# Usage

See https://cray-hpe.github.io/cray-site-init/commands for details on each command.

# Developing and contributing

## Build from source

> Note: You will need to add CRAY to the [GOPRIVATE lib][1] for a clean run:
> ```bash
> export GOPRIVATE="stash.us.cray.com,github.com/Cray-HPE/*"
> ```

1. Using the `makefile`
    ```bash
    $> make
    $> ./bin/csi --help
    ```
2. Calling Go
    ```bash
    $> go build -o bin/csi ./main.go
    $> ./bin/csi --help
    ```

## Contributing

Please create a pull request, and we will review it.

[1]: https://golang.org/cmd/go/#hdr-Module_configuration_for_non_public_modules
