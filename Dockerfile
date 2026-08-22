FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/harborflow ./cmd/server

FROM debian:bookworm-slim
RUN useradd --create-home --uid 10001 harborflow
WORKDIR /app
COPY --from=build /out/harborflow /app/harborflow
COPY migrations /app/migrations
RUN mkdir -p /var/lib/harborflow && chown -R harborflow:harborflow /app /var/lib/harborflow
USER harborflow
ENV PORT=8080 DATABASE_PATH=/var/lib/harborflow/harborflow.db
EXPOSE 8080
ENTRYPOINT ["/app/harborflow"]
