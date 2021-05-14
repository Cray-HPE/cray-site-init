# CRAY Site init (csi)

CSI is a GoLang tool for facilitating the installation of an HPCaaS cluster.

> **`NOTE`** **This deprecates CrayCTL** (`crayctl`) from Shasta V1.4.0 and higher as the primary orchestrator tool.

# Usage

CSI can be installed into any local GoLang **`1.14`** environment.


> Note: You will need to add CRAY to the [GOPRIVATE lib][1] for a clean run: 
> ```bash
> export GOPRIVATE=*.us.cray.com go mod tidy`
> ```

### Build from source

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

CSI is also built for distributing through Linux package managers.

### OS Package Management

#### OpenSuSE 15.2 / SLES 15SP2 / SLE_HPC 15SP2

```bash
# Add repo.
repo=http://car.dev.cray.com/artifactory/csm/MTL/sle15_sp2_ncn/x86_64/dev/master/
zypper addrepo --no-gpgcheck --refresh "$repo" metal_x86-64

# Install.
# FIXME: --no-gpg-checks is not ideal.
zypper --plus-repo=http://car.dev.cray.com/artifactory/csm/MTL/sle15_sp2_ncn/x86_64/dev/master/ --no-gpg-checks -n in -y cray-site-init
```


[1]: https://golang.org/cmd/go/#hdr-Module_configuration_for_non_public_modules
