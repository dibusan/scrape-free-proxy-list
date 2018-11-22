# scrape-free-proxy-list

* Using `dep` for satisfying vendor code: `$ dep init`

### Setup in Production (Ubuntu)
- https://jonathanmh.com/deploying-go-apps-systemd-10-minutes-without-docker/
1. Install go `$ sudo apt-get install golang-go`
2. Setup $GOPATH 
  * `$ mkdir ~/go`
  * Add to `~/.bashrc`

    export GOPATH=$HOME/go
    export PATH=$PATH:$GOPATH/bin

  * Reload bashrc `$ source ~/.bashrc`

3. Download code using `$ go get github.com/dibusan/scrape-free-proxy-list`
4. Update `/lib/systemd/system/scrapefreeproxylist.service` line `ExecStart=/home/{username}/go/src/github.com/dibusan/scrape-free-proxy-list` to reflect the correct {username} 
5. Run installer `$ sudo bash scrape-free-proxy-list/daemon/install.sh`
