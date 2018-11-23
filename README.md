# scrape-free-proxy-list

## Overview
This is the scraper part of a service to make the information in https://free-proxy-list.net available through a REST api. The scraper requests the page every 5 minutes, parses the list of 300 available proxies and updates a REST api at https://rest-free-proxy-list.herokuapp.com/api-docs/index.html . 

To retrieve the list of 300 proxies: GET https://rest-free-proxy-list.herokuapp.com/proxies

## Note
Using `dep` for satisfying vendor code: `$ dep init`

### Setup in Production (Ubuntu)
(https://jonathanmh.com/deploying-go-apps-systemd-10-minutes-without-docker/)
1. Install go `$ sudo apt-get install golang-go`
2. Setup $GOPATH `$ mkdir ~/go`
3. Add to `~/.bashrc`

        export GOPATH=$HOME/go
        export PATH=$PATH:$GOPATH/bin

4. Reload bashrc `$ source ~/.bashrc`            
5. Download code using `$ go get github.com/dibusan/scrape-free-proxy-list`
6. Copy Service file to Systemd `$ cp $GOPATH/src/github.com/dibusan/scrape-free-proxy-list/daemon/scrapefreeproxylist.service /lib/systemd/system/.`
7. Update `/lib/systemd/system/scrapefreeproxylist.service` line `ExecStart=/home/{username}/go/src/github.com/dibusan/scrape-free-proxy-list` to reflect the correct {username} 
5. Reload systemctl daemon `$ sudo systemctl daemon-reload`
5. Start the Service `$ service scrapefreeproxylist start`
6. Enable Service on startup `$ service scrapefreeproxylist enable`
7. Check status of service `$ service scrapefreeproxylist status`
8. Check the logs `$ tail -f /var/log/scrapefreeproxylist.log`
10. Troubleshooting
    
    -   **Error:** service scrapefreeproxylist start reports an Exit Code 203
    
        **Solution:** ensure systemctl is running `service scrapefreeproxylist start` with correct `user:group` for `/var/log`  
