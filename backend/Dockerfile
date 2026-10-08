FROM golang:1.25-alpine AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /api ./cmd/api

FROM alpine:3
RUN adduser -D app
COPY --from=build /api /usr/local/bin/api
USER app
CMD ["api"]
