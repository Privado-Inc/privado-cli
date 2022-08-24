# Privado CLI

[![slack](https://img.shields.io/badge/slack-privado-brightgreen.svg?logo=slack)](https://join.slack.com/t/devprivops/shared_invite/zt-yk5zcxh3-gj8sS9w6SvL5lNYZLMbIpw)


## What is Privado CLI? <a href="#what-is-privado" id="what-is-privado"></a>

Privado is an open source static code analysis tool to discover data flows in the code. 

Privado CLI (previously, [Privado-Inc/privado-datasafety](https://github.com/Privado-Inc/privado-datasafety)) is the user-facing open-source interface responsible for interacting with the final bundle generated from [privado](https://github.com/Privado-Inc/privado) powered by the scan engine [privado-core](https://github.com/Privado-Inc/privado-core), which not only discovers data-elements, PIIs, and third-parties but help discover in-depth dataflows from code to external sinks such as Databases, Third Parties, APIs, and help find data leakages such as logs.

To read more about Privado and how it works, refer [this](https://github.com/Privado-Inc/privado) repository.

## Prerequisite - Docker
To start off, make sure `docker` is installed. To install docker, you can follow the steps stated in the [official documentation](https://docs.docker.com/engine/install/). Linux users should also follow docker [post installation steps](https://docs.docker.com/engine/install/linux-postinstall/#manage-docker-as-a-non-root-user) in order to run Privado CLI without root (`sudo`) privileges.

## Installation
You can install Privado CLI in multiple manners:  

- [Using `curl`](#install-using-curl)
- [Using `go`](#install-using-go)
- [Using releases](#install-release-manually)
- [Build locally](#build-privado-cli-locally)


### Install using `curl`:
The installation script will download and setup the latest stable release for you as per your OS and arch. Run: 

```
curl -o- https://raw.githubusercontent.com/Privado-Inc/privado-cli/main/install.sh | bash
```

To uninstall, simply delete `~/.privado/bin`.

### Install using Go
If you are a GoLang fan, you can use the `go install` command to install the Privado CLI:

```
go install github.com/Privado-Inc/privado-cli@latest
```

This will place the `privado` binary in your `GOPATH`'s bin directory. This directory must be added to the `$PATH` environment variable. You can learn more [here](https://www.digitalocean.com/community/tutorial_series/how-to-install-and-set-up-a-local-programming-environment-for-go). 

### Install Release Manually
We use [GitHub Releases](https://github.com/Privado-Inc/privado-cli/releases) to ship versioned `privado` releases for supported platforms. You can download a executable of Privado CLI for your platform.   

To know your architecture, you can run: 
```
$ uname -m
```
 
For detailed platform-specific instructions to setup `privado`, refer below:
<details>
<summary>MacOSX</summary>
 
#### ARM64 (M1 Chip)
To setup `privado` for macOS (arm64) i.e. Macbook with M1 chip, download `privado-darwin-arm64.tar.gz` from the [latest release](https://github.com/Privado-Inc/privado-cli/releases/latest). 

Navigate to the download directory and run:

```
$ tar -xf ~/.privado/privado-darwin-arm64.tar.gz
$ chmod +x privado
$ mv privado /usr/local/bin/
```

#### AMD64 (Intel Chip)
To setup `privado` for macOS (amd64), download `privado-darwin-amd64.tar.gz` from the [latest release](https://github.com/Privado-Inc/privado-cli/releases/latest). 

Navigate to the download directory and run:

```
$ tar -xf ~/.privado/privado-darwin-amd64.tar.gz
$ chmod +x privado
$ mv privado /usr/local/bin/
```
</details>
 
<details>
 <summary>Linux</summary>

To setup `privado` on your linux system, download the respective zip from [latest release](https://github.com/Privado-Inc/privado-cli/releases/latest) for your platform. Navigate to the download directory and run the following commands:

#### ARM64
```
$ tar -xf ~/.privado/privado-linux-arm64.tar.gz
$ chmod +x privado
$ mv privado /usr/bin/privado
```

#### AMD64
```
$ tar -xf ~/.privado/privado-linux-amd64.tar.gz
$ chmod +x privado
$ mv privado /usr/bin/privado
```

</details>

<details>
<summary>Windows</summary>

To setup `privado` on your windows system, download `privado-windows-amd64.zip` from [latest release](https://github.com/Privado-Inc/privado-cli/releases/latest). Navigate to the download directory and run the following `bash` commands:

```
$ mkdir -p $HOME/.privado/bin
$ unzip -o privado-windows-amd64.zip -d $HOME/.privado/bin
$ chmod +x $HOME/.privado/bin/privado
$ echo "export PATH=\$PATH:$HOME/.privado/bin" >> $HOME/.bashrc
```

Open a new session or source profile for effects to take place in the same session:
```
$ source $HOME/.bashrc
```

When using [WSL](https://docs.microsoft.com/en-us/windows/wsl/), we recommend moving the binary to `/usr/bin` instead for optimal experience across users. Refer to steps for [Linux](#install-release-manually) for more information.

</details>

### Build Privado CLI Locally
If you do not wish to use the pre-built binaries shipped in releases, you can choose to build Privado CLI locally. To do this, make sure that [GoLang](https://go.dev/doc/install) is installed and follow the following steps:

1. Clone the repository: `git clone https://github.com/Privado-Inc/privado-cli.git`    
2. Change directory: `cd privado`   
3. Skip this step if you intend to build the `main` branch.    
    To build a particular [release](https://github.com/Privado-Inc/privado-cli/releases/latest), checkout the intended tag: `git checkout <tag>`   
4. Build with Go: `go build -o privado`   
5. You can now run `./privado`.   

For convenience, we recommend moving `privado` to a `$PATH` directory. You can refer to manual installation steps for more details.

## Running a Scan
> Privado CLI works on the client-end and does not share any code files, or snippets during the scan process.


To scan a repository, simply run:
```
privado scan <path/to/repository>
```
Depending on repository size and system configuration, time to scan can vary. Post completion, you can choose to visualize the results on Privado Cloud.

Results are saved to the `<repository>/.privado` directory. We suggest keeping `.privado` folder a part of your repository to encourage privacy discovery & transparency.

## Command Reference 
The section contains detailed reference to `privado` commands.

### Privado CLI Global Flags
| Flag             	| Description                                                                                    	|
|--------------------------	|---------------------------------------------------------------------------------------------------------------------------	|
| `-h, --help` 	| Help about any command, or sub-command 	|


### Privado CLI Commands
| Command      | Description                                                            | Usage                          | Supported Flags                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| ------------ | ---------------------------------------------------------------------- | ------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `completion` | Generate the autocompletion script for privado for the specified shell | `privado completion [command]` | -                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `config`     | Set config for Privado CLI                                             | `privado config [metrics] [flags]`     | `--enable`, `--disable`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `help`       | Help about any command                                                 | `privado help [command]`       | -                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `scan`       | Scan a codebase or repository to identify dataflows and privacy issues | `privado scan  [flags]`        | `-c, --config <path-to-config>`: <br>Specifies the config (with rules) directory to be passed to privado-core for scanning. These external rules and configurations are merged with the default set that Privado defines<br><br> `--disable-deduplication`: <br>When specified, the engine does not remove duplicate and subset dataflows. This option is useful if you wish to review all flows (including duplicates) manually<br><br> `-o, --overwrite`: <br>If specified, the warning prompt for existing scan results is disabled and any existing results are overwritten<br><br> `-i, --ignore-default-rules `: <br>If specified, the default rules are ignored and only the specified rules (-c) are considered <br><br> `--skip-dependency-download `: <br>When specified, the engine skips downloading all locally unavailable dependencies. Skipping dependency download can yield incomplete results <br><br> `--debug`: <br>To enable process debug output for debugging purposes |
| `update`     | Updates Privado CLI to the latest version                              | `privado update`               | -                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `version`    | Prints the installed version of Privado CLI                            | `privado version`              | -                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |

## How Privado CLI handles your data? <a href="#how-privado-cli-handles-your-data" id="how-privado-cli-handles-your-data"></a>

Privado CLI was engineered with security in mind. Our tool runs the scan locally on your machine and your code never leaves your system.

## License
Privado OSS is distributed under the GNU Lesser GENERAL PUBLIC LICENSE(AGPL 3.0). This application may only be used in compliance with the License. In lieu of applicable law or written agreement, software distributed under the License is distributed "AS IS", VOID OF ALL WARRANTIES OR CONDITIONS. For specific details regarding permissions and restrictions, see [COPYING](/COPYING) and [COPYING.LESSER](/COPYING.LESSER).

<!-- 
### Source-code pre-text

This file is part of Privado OSS.

Privado is an open source static code analysis tool to discover data flows in the code.
Copyright (C) 2022 Privado, Inc.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>. 

For more information, contact support@privado.ai
-->
