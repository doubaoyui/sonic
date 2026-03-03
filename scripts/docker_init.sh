#!/bin/sh

check_and_copy() {
    if [ ! -e /sonic/$1 ]; then
        mkdir -p /sonic/$1
        cp -Rf /app/$1/* /sonic/$1/
    fi
}

# Data-only directories (db/logs/uploads) live in the volume at /sonic.
mkdir -p /data/logs /data/upload
# Remove legacy persisted UI/theme assets (keeps host volume clean; assets come from the image).
# Only remove the path we previously created by mistake.
if [ -d /data/resources/admin ] || [ -d /data/resources/template ]; then
    rm -rf /data/resources
fi

# Backward compatibility: older deployments may still point to /sonic/conf/config.yaml
check_and_copy 'conf'

