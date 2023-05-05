$repo = "bit-cloner/aws-vitals-monitor"
$repoUrl = "https://github.com/$repo"
$apiUrl = "https://api.github.com/repos/$repo/releases/latest"
$binaryName = "avm"
$os = "windows"
$arch = "amd64"
$destDir = "C:\Program Files\$binaryName"

# Fetch the latest release version
$version = (Invoke-RestMethod -Uri $apiUrl).tag_name

# Download the appropriate binary for the system
$releaseUrl = "$repoUrl/releases/download/$version"
$downloadUrl = "$releaseUrl/$binaryName-$version-$os-$arch.zip"
Write-Host "Downloading $downloadUrl..."
Invoke-WebRequest -Uri $downloadUrl -OutFile "$binaryName.zip"

# Unzip the downloaded file
Write-Host "Extracting $binaryName.zip..."
Expand-Archive -Path "$binaryName.zip" -DestinationPath "."

# Create the destination directory if it does not exist
if (-not (Test-Path $destDir)) {
    New-Item -ItemType Directory -Path $destDir
}

# Move the binary to the destination directory
Write-Host "Installing $binaryName to $destDir..."
Move-Item -Path ".\$binaryName.exe" -Destination "$destDir\$binaryName.exe"

# Add the binary location to the PATH environment variable
$envPath = [System.Environment]::GetEnvironmentVariable("Path", [System.EnvironmentVariableTarget]::Machine)
if (-not ($envPath -split ';' -contains $destDir)) {
    [System.Environment]::SetEnvironmentVariable("Path", $envPath + ";$destDir", [System.EnvironmentVariableTarget]::Machine)
}

Write-Host "Installation completed. $binaryName is now available in $destDir."
