&nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp;![alt text](healthcheck.gif "AVM")
# AWS Vitals Monitor
AVM - AWS vitals Monitor

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

AWS Vitals Monitor is a tool designed to streamline the process of analyzing and understanding the usage patterns and health of various AWS services within your account. This tool provides a user-friendly interface from the terminal for obtaining key insights and metrics associated with your deployed AWS resources. By simplifying the identification of essential service characteristics, AWS Vitals Monitor enables users to optimize their infrastructure, improve performance, and reduce costs. AWS Vitals Monitor is a solution for gaining a holistic view of your AWS environment and ensuring the smooth operation of your cloud-based services.

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

Pre-compiled binaries are available for Windows, macOS, and Linux systems in the [releases](https://github.com/bit-cloner/aws-vitals-monitor/releases) section. Download the appropriate binary for your platform.

### Oneliner for Mac and Linux systems
```
curl -sSL https://raw.githubusercontent.com/bit-cloner/aws-vitals-monitor/main/install.sh | bash
```
### Oneliner for windows systems
```
Invoke-Expression ((New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/bit-cloner/aws-vitals-monitor/main/install.ps1'))
```
### macOS

1. Download the `avm-darwin-amd64` binary from the [releases](https://github.com/bit-cloner/aws-vitals-monitor/releases) page.
```
curl -L -o avm-darwin-amd64.tar.gz https://github.com/bit-cloner/aws-vitals-monitor/releases/download/0.2/avm-0.2-darwin-amd64.tar.gz
```
2. Untar
```
tar -xvzf avm-0.2-darwin-amd64.tar.gz
```
2. Make the binary executable: 
```
chmod +x avm-0.2-darwin-amd64
```
3. Move the binary to a directory included in your `PATH` environment variable 
```
mv avm-0.2-darwin-amd64 /usr/local/bin/avm
```




### Linux

1. Download the `avm-linux-amd64` binary from the [releases](https://github.com/bit-cloner/aws-vitals-monitor/releases) page.
```
curl -L -o avm-linux-amd64.tar.gz https://github.com/bit-cloner/aws-vitals-monitor/releases/download/0.2/avm-0.2-linux-amd64.tar.gz
```
2. Untar
```
tar -xvzf avm-0.2-linux-amd64.tar.gz
```
3. Make the binary executable: 
```
chmod +x avm-0.2-linux-amd64
```
4. Move the binary to a directory included in your `PATH` environment variable 
```
sudo mv avm-0.2-linux-amd64 /usr/local/bin/avm
```

Now you can run the `avm` command from any directory on your Linux system.

### Windows

1. Download the `avm-windows-amd64.zip` binary from the [releases](https://github.com/bit-cloner/aws-vitals-monitor/releases) page.
2. Unzip the downloaded file to a directory. You can use the Expand-Archive cmdlet in PowerShell to unzip the file: 
  Rename the binary to `avm.exe` for ease of use.
  ```
  Expand-Archive -Path .\avm-windows-amd64.zip -DestinationPath C:\path\to\avm\directory
  ```
3. Move the binary to a directory included in your `PATH` environment variable or add the binary's location to your `PATH`.
```powershell
Move-Item .\avm.exe C:\path\to\directory\in\PATH

$env:Path += ";C:\path\to\avm\directory"
```

## Usage
```
./avm
```
or
```
avm.exe
```
Make sure AWS credentials are available in the current terminal . for ex: 
```
export AWS_ACCESS_KEY_ID="your_aws_access_key" && export AWS_SECRET_ACCESS_KEY="your_aws_secret_key" && export AWS_SESSION_TOKEN="your_aws_session_token"

```