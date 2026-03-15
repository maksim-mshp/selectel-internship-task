$REPO_URL = "https://github.com/maksim-mshp/selectel-internship-task"
$MODULE_PATH = "github.com/maksim-mshp/selectel-internship-task"
$PLUGIN_DIR = "tools/loglint"

if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
	Write-Host "git not found"
	exit 1
}

if (-not (Get-Command golangci-lint -ErrorAction SilentlyContinue)) {
	Write-Host "golangci-lint not found"
	exit 1
}

if (-not (Test-Path (Join-Path $PLUGIN_DIR ".git"))) {
	New-Item -ItemType Directory -Force -Path $PLUGIN_DIR | Out-Null
	git clone $REPO_URL $PLUGIN_DIR
}

$ver_line = golangci-lint version 2>$null
if ($ver_line -match "version ([0-9A-Za-z\\.\\-]+)") {
	$gcl_ver = $Matches[1]
} else {
	Write-Host "Failed to detect golangci-lint version"
	exit 1
}

if ($gcl_ver -notmatch "^v") {
	$gcl_ver = "v$gcl_ver"
}

$custom_gcl = @"
version: $gcl_ver
plugins:
  - module: $MODULE_PATH
    import: $MODULE_PATH/gclplugin
    path: $PLUGIN_DIR

"@
[System.IO.File]::WriteAllText("$PWD\.custom-gcl.yml", $custom_gcl.Replace("`r`n", "`n"))

$golangci = @"
version: "2"

linters:
  default: none
  enable:
    - loglint
  settings:
    custom:
      loglint:
        type: module
        description: log message checks
        config: .loglint.yml

"@
[System.IO.File]::WriteAllText("$PWD\.golangci.yml", $golangci.Replace("`r`n", "`n"))

$golangci = @"
lowercase: true
english: true
special: true
sensitive: true
patterns:
  - '(?i)password'
  - '(?i)api_key'
  - '(?i)token'

"@
[System.IO.File]::WriteAllText("$PWD\.loglint.yml", $golangci.Replace("`r`n", "`n"))
