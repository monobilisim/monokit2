#!/bin/sh
if [ -d /artifacts ]; then
    echo "Extracting artifacts to /artifacts..."
    cp /var/lib/monokit2/monokit2.db /artifacts/ 2>/dev/null
    cp /var/log/monokit2.log /artifacts/ 2>/dev/null

    if [ -n "$HOST_UID" ] && [ -n "$HOST_GID" ]; then
        chown "$HOST_UID:$HOST_GID" /artifacts/* 2>/dev/null
    fi
else
    echo "/artifacts directory not found, skipping artifact extraction."
fi
