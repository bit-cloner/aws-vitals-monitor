# Project Name

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

AWS Health checks is a tool designed to aid CREs in gaining a deeper understanding of an AWS account. It simplifies the task of identifying key characteristics of AWS services being used in an account.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
  - [Windows](#windows)
  - [macOS](#macos)
  - [Linux](#linux)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

## Features

- List the key features of your tool.

## Installation

Pre-compiled binaries are available for Windows, macOS, and Linux systems in the [releases](https://github.com/doitintl/aws-health-checks/releases) section. Download the appropriate binary for your platform.

### Windows

1. Download the `hca-windows-amd64.exe` binary from the [releases](https://github.com/yourusername/project-name/releases) page.
2. Rename the binary to `hca.exe` for ease of use.
3. Move the binary to a directory included in your `PATH` environment variable or add the binary's location to your `PATH`.
```powershell
Move-Item .\hca.exe C:\path\to\directory\in\PATH

$env:Path += ";C:\path\to\hca\directory"
```
### macOS

1. Download the `hca-darwin-amd64` binary from the [releases](https://github.com/yourusername/project-name/releases) page.
```
curl -L -o hca-darwin-amd64 https://github.com/yourusername/project-name/releases/download/v<version>/hca-darwin-amd64
```
2. Make the binary executable: 
```
chmod +x hca-darwin-amd64
```
3. Move the binary to a directory included in your `PATH` environment variable 
```
mv hca-darwin-amd64 /usr/local/bin/hca
```

### Linux

1. Download the `hca-linux-amd64` binary from the [releases](https://github.com/yourusername/project-name/releases) page.
```
curl -L -o hca-linux-amd64 https://github.com/yourusername/project-name/releases/download/v<version>/hca-linux-amd64
```
2. Make the binary executable: 
```
chmod +x hca-linux-amd64
```
3. Move the binary to a directory included in your `PATH` environment variable (e.g., `/usr/local/bin`): 
```
mv hca-linux-amd64 /usr/local/bin/hca
```

## Usage

Provide a brief example or instructions on how to use the tool. Include examples of command-line usage and any required input files or parameters.