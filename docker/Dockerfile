FROM golang:1.27 AS builder

COPY ../ /app
WORKDIR /app
RUN ls
# Toggle CGO on your app requirement
RUN CGO_ENABLED=0 go build -ldflags '-s -w -extldflags "-static"' -o /app/appbin *.go
# Use below if using vendor
# RUN CGO_ENABLED=0 go build -mod=vendor -ldflags '-extldflags "-static"' -o /app/appbin *.go

FROM gcr.io/distroless/static-debian13
LABEL org.opencontainers.image.author="Naughtygopher<github.com/naughtygopher>"


COPY --from=builder /app/cmd/server/http/web /home/appuser/app/web
COPY --from=builder /app/appbin /home/appuser/app

ENV TEMPLATES_BASEPATH=/home/appuser/app/web/templates

WORKDIR /home/appuser/app
USER nonroot

# Since running as a non-root user, port bindings < 1024 are not possible
# 8000 for HTTP; 8443 for HTTPS;
EXPOSE 8000

# this one is the probe handling server
EXPOSE 2000

CMD ["./appbin"]