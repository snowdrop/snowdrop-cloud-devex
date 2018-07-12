#!/usr/bin/env bash

# Execute the Bootstrap application generating the supervisord's conf file
echo "Call the application generating the supervisord's conf file"
/opt/supervisord/bin/bootstrap

# Copy files to their target location
/usr/bin/cp -r /opt/supervisord /var/lib/
