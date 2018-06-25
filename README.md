# High-level overview
1. Use `bbl` to install BOSH
2. Download stemcell and Concourse releases
3. Deploy Concourse
4. Setup socks5 proxy to jumpbox (e.g., ssh -i id_rsa -D 9999 -q -N jumpbox@35.227.150.58 -f)
  * In Crostini and in ChromeOS WiFi settings
