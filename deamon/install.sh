#! /bin/bash
set -x

# Copy demon to systemd folder
cp $(pwd)/scrapefreeproxylist.service /lib/systemd/system/.

# Start service
service wombatapp start

# Enable service on bootup
service wombatapp enable

# Check status of service
service wombatapp status
