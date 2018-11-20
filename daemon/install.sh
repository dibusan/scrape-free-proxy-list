#! /bin/bash
set -x

# Copy demon to systemd folder
cp $(pwd)/daemon/scrapefreeproxylist.service /lib/systemd/system/.

# Start service
service scrapefreeproxylist start

# Enable service on bootup
service scrapefreeproxylist enable

# Check status of service
service scrapefreeproxylist status
