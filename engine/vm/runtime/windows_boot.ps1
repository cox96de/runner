$scriptPath = $MyInvocation.MyCommand.Path
$scriptDir = Split-Path -Parent $scriptPath
$executorPath = Join-Path $scriptDir "executor.exe"

Start-Process -FilePath $executorPath -RedirectStandardOutput "C:\executor.log" -RedirectStandardError "C:\executor_error.log"