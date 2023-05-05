#!/bin/bash
set -e

# Variables
repo="bit-cloner/aws-vitals-monitor"
repo_url="https://github.com/${repo}"
api_url="https://api.github.com/repos/${repo}/releases/latest"
binary_name="avm"
os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
dest_dir="/usr/local/bin"

# Fetch the latest release version
version=$(curl -sSL "${api_url}" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

# Download the appropriate binary for the system
if [[ "${arch}" == "x86_64" ]]; then
    arch="amd64"
elif [[ "${arch}" == "arm64" ]]; then
    arch="arm64"
else
    echo "Unsupported architecture: ${arch}"
    exit 1
fi

release_url="${repo_url}/releases/download/${version}"
download_url="${release_url}/${binary_name}-${version}-${os}-${arch}.tar.gz"
echo "Downloading ${download_url}..."
curl -sSL -o "${binary_name}.tar.gz" "${download_url}"

# Untar the downloaded file
echo "Extracting ${binary_name}.tar.gz..."
tar -xzf "${binary_name}.tar.gz"

# Make the binary executable
chmod +x "${binary_name}"

# Move the binary to the destination directory
echo "Installing ${binary_name} to ${dest_dir}..."
sudo mv "${binary_name}" "${dest_dir}"

echo "Installation completed. ${binary_name} is now available in ${dest_dir}."
