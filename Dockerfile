ARG PACKAGE=github.com/brimstone/traefik-cert
FROM brimstone/golang:1.12 as builder

FROM scratch
ENV ADDRESS= \
    KEY=
EXPOSE 80
ENTRYPOINT ["/traefik-cert", "serve"]
COPY --from=builder /app /traefik-cert
