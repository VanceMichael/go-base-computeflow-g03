FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/computeflow ./cmd/server

FROM debian:bookworm-slim
RUN useradd --create-home --uid 10001 computeflow
WORKDIR /app
COPY --from=build /out/computeflow /app/computeflow
COPY migrations /app/migrations
RUN mkdir -p /var/lib/computeflow && chown -R computeflow:computeflow /app /var/lib/computeflow
USER computeflow
ENV PORT=8080 DATABASE_PATH=/var/lib/computeflow/computeflow.db
EXPOSE 8080
ENTRYPOINT ["/app/computeflow"]
