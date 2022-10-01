FROM scratch

ADD buildfs/rootfs.tar.gz /

ENTRYPOINT ["/app"]

HEALTHCHECK --interval=16s --timeout=2s \
    CMD ["/usr/bin/xh", "--quiet", "get", "http://127.0.0.1/healthcheck"]
