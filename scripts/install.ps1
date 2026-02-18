param(
    [string]$Version = "latest"
)

$ErrorActionPreference = "Stop"

$RepoOwner = "bamboo-services"
$RepoName = "bamboo-base-go-cli"
$BinaryPrefix = "bamboo-base-cli"
$InstallName = "bamboo.exe"

function Write-Info {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Yellow
}

function Write-Err {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Red
}

function Get-Architecture {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default {
            Write-Err "不支持的架构: $env:PROCESSOR_ARCHITECTURE"
            exit 1
        }
    }
}

function Download-WithRetry {
    param(
        [string]$Url,
        [string]$OutputPath,
        [int]$MaxAttempts = 3
    )

    for ($attempt = 1; $attempt -le $MaxAttempts; $attempt++) {
        try {
            Invoke-WebRequest -Uri $Url -OutFile $OutputPath -UseBasicParsing
            return $true
        }
        catch {
            if ($attempt -lt $MaxAttempts) {
                Start-Sleep -Seconds 2
            }
        }
    }
    return $false
}

function Get-LatestVersion {
    try {
        $releaseUrl = "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
        $release = Invoke-RestMethod -Uri $releaseUrl -UseBasicParsing
        return ($release.tag_name -replace '^v', '')
    }
    catch {
        Write-Err "无法获取最新版本"
        exit 1
    }
}

function Main {
    $Arch = Get-Architecture
    Write-Success "检测到系统: windows-$Arch"

    if ($Version -eq "latest") {
        $Version = Get-LatestVersion
    }
    else {
        $Version = $Version -replace '^v', ''
    }

    $BinaryName = "$BinaryPrefix-windows-$Arch.exe"
    $DownloadUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/v$Version/$BinaryName"
    $ChecksumUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/v$Version/checksums.txt"

    $TmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
    New-Item -ItemType Directory -Path $TmpDir -Force | Out-Null

    try {
        $BinaryPath = Join-Path $TmpDir $BinaryName
        if (-not (Download-WithRetry -Url $DownloadUrl -OutputPath $BinaryPath)) {
            Write-Err "下载失败: $DownloadUrl"
            exit 1
        }
        Write-Success "二进制下载完成"

        $ChecksumPath = Join-Path $TmpDir "checksums.txt"
        if (Download-WithRetry -Url $ChecksumUrl -OutputPath $ChecksumPath) {
            $ExpectedLine = Get-Content $ChecksumPath | Where-Object { $_ -match $BinaryName } | Select-Object -First 1
            if ($ExpectedLine) {
                $ExpectedChecksum = ($ExpectedLine -split '\s+')[0].ToLower()
                $ActualChecksum = (Get-FileHash -Path $BinaryPath -Algorithm SHA256).Hash.ToLower()
                if ($ExpectedChecksum -ne $ActualChecksum) {
                    Write-Err "校验失败"
                    Write-Err "期望: $ExpectedChecksum"
                    Write-Err "实际: $ActualChecksum"
                    exit 1
                }
                Write-Success "校验通过"
            }
            else {
                Write-Warn "未找到对应校验值，跳过校验"
            }
        }
        else {
            Write-Warn "未下载到 checksums.txt，跳过校验"
        }

        $InstallDir = Join-Path $env:USERPROFILE ".local\bin"
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }

        $DestPath = Join-Path $InstallDir $InstallName
        Copy-Item -Path $BinaryPath -Destination $DestPath -Force
        Write-Success "安装完成: $DestPath"

        $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($UserPath -notlike "*$InstallDir*") {
            Write-Warn "$InstallDir 不在 PATH 中"
            Write-Info "请将 $InstallDir 加入用户 PATH 后重开终端"
        }
        else {
            Write-Success "现在可运行: bamboo --help"
        }
    }
    finally {
        if (Test-Path $TmpDir) {
            Remove-Item -Path $TmpDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}

Main
