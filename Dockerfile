# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build
WORKDIR /src

# Cached separately from the source copy below so `docker build` only
# re-downloads modules when go.mod/go.sum actually change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/masterwill-api ./cmd/api

# Migrations are embedded into the binary (see internal/db/migrate.go), so
# the runtime image needs nothing but the binary itself — distroless static
# has no shell or package manager, which is exactly the attack surface a
# single-binary Go service doesn't need.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/masterwill-api ./masterwill-api

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/masterwill-api"]
