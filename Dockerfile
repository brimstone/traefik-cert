FROM brimstone/golang-musl as builder

FROM scratch
ENV ADDRESS= \
    KEY=
EXPOSE 80
ENTRYPOINT ["/traefik-cert", "serve"]
COPY --from=builder /app /traefik-cert
