Param(
    [String]$ComputerName,
    [String]$ProgramPath,
    [String]$Config,
    [String]$Destination
)

$session = New-PSSession -ComputerName $ComputerName

Copy-Item -Path $Config -Destination $Destination -ToSession $session
$NewConfig = $Destination

Invoke-Command -Session $session -ScriptBlock {Set-Location $Using:ProgramPath}
Invoke-Command -Session $session -ScriptBlock { .\auto-processing.exe --cfg=$Using:NewConfig }
$exitcode = Invoke-Command -Session $session -ScriptBlock { echo $LastExitCode }

Invoke-Command -Session $session -Script {Remove-Item -Path $Using:NewConfig}
Disconnect-PSSession -Session $session

exit $ExitCode