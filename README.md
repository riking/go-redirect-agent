## go-redirect-agent

Small go agent binary for running on a system to redirect http://go/ links to a secure URL.

By default, the agent will listen on `127.0.103.111`. It will not
listen on the IPv6 loopback interface as you might want to run other programs on port 80.

## Install

### Windows

Download a release and unzip it into Downloads/go-redirect-agent or another location as you prefer.

Open an elevated PowerShell prompt.

[image-powershell-runas]

```ps
PS C:\Windows\system32>  cd (gci env:USERPROFILE).Value
PS C:\Users\You> cd Downloads\go-redirect-agent

# Choose one of the following:

PS C:\Users\You\Downloads\go-redirect-agent> .\go-redirect-agent.exe --golinks install

PS C:\Users\You\Downloads\go-redirect-agent> .\go-redirect-agent.exe -d https://goto.contoso.com/ install

PS C:\Users\You\Downloads\go-redirect-agent> .\go-redirect-agent.exe -d https://subdomain.site.com/go/ install

# ... Then, install the hosts file entry:

PS C:\Users\You\Downloads\go-redirect-agent> .\install_windows.bat
```
