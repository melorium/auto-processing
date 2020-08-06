package powershell

import "fmt"

func (c *Client) autoProcessing(archive, path, cfg string) string {
	return fmt.Sprintf(`
	$ComputerName = "%s";
	$Archive = "%s";
	$Path = "%s";
	$Config = "%s";
	$ExitCode = 0;
	
	$Password = ConvertTo-SecureString -String %s -AsPlainText -Force;
	$Cred = New-Object -TypeName 'System.Management.Automation.PSCredential' -ArgumentList %s, $Password;

	$session = New-PSSession -ComputerName $ComputerName -Credential $Cred;
	If (-not($session)) { $ExitCode = 40; exit $ExitCode;};
	
	$session_tmp = Invoke-Command -Session $session -Script { echo $env:temp };
	If (-not($session_tmp)) { $ExitCode = 45; exit $ExitCode;};
	
	Copy-Item -Path $Path -Destination $session_tmp -ToSession $session;
	
	Invoke-Command -Session $session -ScriptBlock { Set-Location $Using:session_tmp };
	Invoke-Command -Session $session -ScriptBlock { Expand-Archive -Path $Using:Archive -DestinationPath .\unzipped-$Using:Archive };
	Invoke-Command -Session $session -Command { Remove-Item $Using:Archive };
	Invoke-Command -Session $session -ScriptBlock { cd unzipped-$Using:Archive };
	Invoke-Command -Session $session -ScriptBlock { .\auto-processing.exe --cfg=$Using:Config };

	$ExitCode = Invoke-Command -Session $session -ScriptBlock { echo $LastExitCode };

	Invoke-Command -Session $session -ScriptBlock { Set-Location $Using:session_tmp };
	Invoke-Command -Session $session -Command { Remove-Item .\unzipped-$Using:Archive -Recurse };
	
	Disconnect-PSSession -Session $session;
	
	exit $ExitCode;
	`, c.Host, archive, path, cfg, c.Password, c.Username)
}

func (c *Client) testConnection(path string) string {
	return fmt.Sprintf(`
	$ComputerName = "%s";
	$Path = "%s";
	$ExitCode = 0;
	
	$Password = ConvertTo-SecureString -String %s -AsPlainText -Force;
	$Cred = New-Object -TypeName 'System.Management.Automation.PSCredential' -ArgumentList %s, $Password;

	$session = New-PSSession -ComputerName $ComputerName -Credential $Cred;
	If (-not($session)) { $ExitCode = 40; exit $ExitCode;};

	$pathCheck = Invoke-Command -Session $s -ScriptBlock {Test-Path -path $Using:Path}
	if (-not($pathCheck)) { $ExitCode = 50 };

	Disconnect-PSSession -Session $session | Out-Null;
	exit $ExitCode;
	`, c.Host, path, c.Password, c.Username)
}
