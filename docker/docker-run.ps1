# ---------------------------------------------------------------------------
# docker-run.ps1 — Quick-start script for Windows (PowerShell)
#
# Detects your Cursor auth and launches the proxy container.
# Usage: .\docker\docker-run.ps1 [-Mode proxy|both|test]
# ---------------------------------------------------------------------------
param(
    [string]$Mode = "proxy",
    [int]$Port = 32125,
    [int]$GatewayPort = 18789
)

$ErrorActionPreference = "Stop"
$Image = "openclaw-cursor:latest"
$Container = "openclaw-cursor-proxy"

# Detect Cursor auth directory (Windows paths)
$AuthDir = Join-Path $env:APPDATA "Cursor"
$AuthFile = Join-Path $AuthDir "auth.json"
if (-not (Test-Path $AuthFile)) {
    $AuthDir = Join-Path $env:LOCALAPPDATA "cursor"
    $AuthFile = Join-Path $AuthDir "auth.json"
}
if (-not (Test-Path $AuthFile)) {
    # WSL2 fallback
    $AuthDir = Join-Path $env:USERPROFILE ".config\cursor"
}

Write-Host "Platform:  Windows/$([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture)"
Write-Host "Auth dir:  $AuthDir"
Write-Host "Mode:      $Mode"
Write-Host "Port:      $Port"
Write-Host ""

# Build if image doesn't exist
$exists = docker image inspect $Image 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "Building Docker image..."
    $ScriptDir = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
    $Version = git -C $ScriptDir describe --tags --always 2>$null
    if (-not $Version) { $Version = "dev" }
    docker build -t $Image --build-arg "VERSION=$Version" $ScriptDir
}

# Stop existing container
docker rm -f $Container 2>$null | Out-Null

# Port arguments
$PortArgs = @("-p", "${Port}:32125")
if ($Mode -eq "both") {
    $PortArgs += @("-p", "${GatewayPort}:18789")
}

# Build docker run command
$RunArgs = @("run", "-d", "--name", $Container) + $PortArgs

# Auth mount (convert Windows path to Docker-compatible)
if (Test-Path $AuthDir) {
    $DockerAuthDir = $AuthDir -replace '\\', '/'
    $RunArgs += @("-v", "${DockerAuthDir}:/root/.config/cursor:ro")
    Write-Host "Mounting auth from: $AuthDir"
} else {
    Write-Host "WARNING: No Cursor auth directory found at $AuthDir"
    Write-Host "         Set CURSOR_ACCESS_TOKEN and CURSOR_REFRESH_TOKEN instead"
}

$RunArgs += @("-v", "openclaw-data:/root/.openclaw")

# Environment variables
if ($env:CURSOR_ACCESS_TOKEN) {
    $RunArgs += @("-e", "CURSOR_ACCESS_TOKEN=$($env:CURSOR_ACCESS_TOKEN)")
}
if ($env:CURSOR_REFRESH_TOKEN) {
    $RunArgs += @("-e", "CURSOR_REFRESH_TOKEN=$($env:CURSOR_REFRESH_TOKEN)")
}
if ($env:CURSOR_API_KEY) {
    $RunArgs += @("-e", "CURSOR_API_KEY=$($env:CURSOR_API_KEY)")
}

$RunArgs += @($Image, $Mode)

docker @RunArgs

Write-Host ""
Write-Host "Container started: $Container"
Write-Host "Health check: curl http://127.0.0.1:${Port}/health"

# Wait for health
Write-Host -NoNewline "Waiting for proxy..."
for ($i = 0; $i -lt 20; $i++) {
    try {
        $response = Invoke-RestMethod -Uri "http://127.0.0.1:${Port}/health" -TimeoutSec 3 -ErrorAction Stop
        Write-Host " ready!"
        $response | ConvertTo-Json -Depth 5
        exit 0
    } catch {
        Write-Host -NoNewline "."
        Start-Sleep -Seconds 2
    }
}
Write-Host " timeout (check: docker logs $Container)"
exit 1
