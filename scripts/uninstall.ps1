$ErrorActionPreference = "Stop"

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

function Confirm-Action {
    param([string]$Prompt)
    $choices = "&Yes", "&No"
    $decision = $Host.UI.PromptForChoice("", $Prompt, $choices, 1)
    return $decision -eq 0
}

function Main {
    $InstallPath = Join-Path $env:USERPROFILE ".local\bin\bamboo.exe"

    if (-not (Test-Path $InstallPath)) {
        $found = Get-Command bamboo -ErrorAction SilentlyContinue
        if (-not $found) {
            Write-Info "未检测到已安装的 bamboo"
            return
        }
        $InstallPath = $found.Source
    }

    Write-Info "将删除: $InstallPath"
    if (-not (Confirm-Action "确认卸载 bamboo 吗？")) {
        Write-Info "已取消"
        return
    }

    Remove-Item -Path $InstallPath -Force -ErrorAction Stop
    Write-Success "卸载完成"

    $stillInPath = Get-Command bamboo -ErrorAction SilentlyContinue
    if ($stillInPath) {
        Write-Warn "PATH 中仍存在 bamboo: $($stillInPath.Source)"
    }
}

Main
