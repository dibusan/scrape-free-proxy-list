# scrape-free-proxy-list

* Using `dep` for satisfying vendor code: `$ dep init`

### To Deploy in Production (Ubuntu)
- https://jonathanmh.com/deploying-go-apps-systemd-10-minutes-without-docker/
1. Clone repo to Ubuntu server `$ git clone https://github.com/dibusan/scrape-free-proxy-list.git`
2. Update `/lib/systemd/system/scrapefreeproxylist.service` line `ExecStart=/home/erieljr1/scrape-free-proxy-list` to reflect the correct path for the downloaded application
3. Run installer `$ bash scrape-free-proxy-list/daemon/install.sh`
