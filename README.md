&nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp;![alt text](healthcheck.gif "HCA")
# AWS health checks
HCA- Health  checks for AWS

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

### macOS

1. Download the `hca-darwin-amd64` binary from the [releases](https://github.com/yourusername/project-name/releases) page.
```
curl -L -o hca-darwin-amd64.tar.gz https://github.com/doitintl/aws-health-checks/releases/download/0.2/hca-darwin-amd64.tar.gz
```
2. Untar
```
tar -xvzf hca-darwin-amd64.tar.gz
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
curl -L -o hca-linux-amd64.tar.gz https://github.com/doitintl/aws-health-checks/releases/download/0.2/hca-linux-amd64.tar.gz
```
2. Untar
```
tar -xvzf hca-linux-amd64.tar.gz
```
3. Make the binary executable: 
```
chmod +x hca-linux-amd64
```
4. Move the binary to a directory included in your `PATH` environment variable 
```
sudo mv hca-linux-amd64 /usr/local/bin/hca
```

Now you can run the `hca` command from any directory on your Linux system.

### Windows

1. Download the `hca-windows-amd64.exe` binary from the [releases](https://github.com/yourusername/project-name/releases) page.
2. Rename the binary to `hca.exe` for ease of use.
3. Move the binary to a directory included in your `PATH` environment variable or add the binary's location to your `PATH`.
```powershell
Move-Item .\hca.exe C:\path\to\directory\in\PATH

$env:Path += ";C:\path\to\hca\directory"
```

## Usage
```
./hca
```
Make sure AWS credentials are available in the current terminal . for ex: 
```
export AWS_ACCESS_KEY_ID="your_aws_access_key" && export AWS_SECRET_ACCESS_KEY="your_aws_secret_key" && export AWS_SESSION_TOKEN="your_aws_session_token"

```