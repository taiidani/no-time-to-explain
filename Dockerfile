FROM alpine:3.19

# Dependencies
RUN apk add --no-cache ca-certificates

# Add pre-built application
COPY bin/app /app
RUN chmod +x /app
ENTRYPOINT [ "/app" ]

LABEL org.opencontainers.image.source https://github.com/taiidani/no-time-to-explain
