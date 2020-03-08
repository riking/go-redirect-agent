## go-redirect-agent

Small go agent binary for running on a system to redirect http://go/ links to a secure URL. Compatible with [golinks.io](https://www.golinks.io/).

By default, the agent will listen on `127.0.103.111`. It will not listen on the IPv6 loopback interface as you might want to run other programs on port 80. (103 and 111 are the decimal ASCII codes for 'g' 'o'.)

## Install

### Windows

Download a release and unzip it into Downloads/go-redirect-agent or another location as you prefer.

Open an elevated PowerShell prompt.

[image-powershell-runas]

```ps
PS C:\Windows\system32>  cd (gci env:USERPROFILE).Value
PS C:\Users\You> cd Downloads\go-redirect-agent

# Choose one of the following:

PS C:\Users\You\Downloads\go-redirect-agent> .\go-redirect-agent.exe --golinks install # Uses https://www.golinks.io/

PS C:\Users\You\Downloads\go-redirect-agent> .\go-redirect-agent.exe -d https://goto.contoso.com/ install

PS C:\Users\You\Downloads\go-redirect-agent> .\go-redirect-agent.exe -d https://subdomain.site.com/go/ install

# ... Then, install the hosts file entry:

PS C:\Users\You\Downloads\go-redirect-agent> .\go-redirect-agent.exe start
PS C:\Users\You\Downloads\go-redirect-agent> .\install_windows.bat
```

### Unix (systemd)

Use `go get` to compile the binary: `go get -v github.com/riking/go-redirect-agent` or download a compiled version and set `AGENT_BIN` to its path.

Set `AGENT_HOST` to the secure server path hosting the go link redirection service, e.g. `AGENT_HOST=https://subdomain.site.com/go/` , or leave blank to use golinks.io.

Run the following commands:

```bash
go get -v github.com/riking/go-redirect-agent
export AGENT_BIN="$(go env GOBIN)/go-redirect-agent"
export AGENT_HOST="${AGENT_HOST:-https://www.golinks.io/}"
export AGENT_IP="${AGENT_IP:-127.0.103.111}"
export AGENT_LISTEN="${AGENT_IP}:80"

sudo --preserve-env=AGENT_BIN,AGENT_HOST,AGENT_IP,AGENT_LISTEN bash -e

# Copy to /usr/local/bin
install "${AGENT_BIN}" /usr/local/bin/go-redirect-agent

# check & install hosts file
grep -q "${AGENT_IP}" /etc/hosts || ( echo "Adding to /etc/hosts..." ; echo "${AGENT_IP} go" | tee /etc/hosts )

cat >/etc/systemd/system/go-redirect-agent.service <<EOF
[Unit]
Name=go-redirect-agent
Description=Small HTTP server to redirect http://go/ links. https://github.com/riking/go-redirect-agent/

[Service]
Type=simple
ExecStart=/usr/local/bin/go-redirect-agent -d ${AGENT_HOST} -l ${AGENT_LISTEN}
Restart=always
User=root

[Install]
WantedBy=multi-user.target

EOF

systemctl enable go-redirect-agent
systemctl start go-redirect-agent

exit # the sudo'ed shell
```
