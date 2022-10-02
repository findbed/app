#!/usr/bin/env bash

cp /usr/share/zoneinfo/Europe/Moscow $ROOTFS/etc/localtime
echo "Europe/Moscow" > $ROOTFS/etc/timezone

cd $CWD
CGO_ENABLED=0 go build -v -ldflags="-s -w" -o $ROOTFS/app . || exit 1

ARCH=$(uname -m)
if [ "$ARCH" == "x86_64" ]; then
    echo apk add --no-cache upx@community
    echo upx -v --brute $ROOTFS/app
fi
