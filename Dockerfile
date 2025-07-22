FROM golang:1.24.5-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o .
FROM golang:1.24.5-bookworm
WORKDIR /app
RUN apt-get update && apt-get install -y tar curl && curl -fsSL https://github.com/smallstep/cli/releases/latest/download/step_linux_amd64.tar.gz | tar xz && mv step_linux_amd64/bin/step /usr/local/bin && rm -rf step_linux_amd64*
COPY --from=builder /app/auth-side-to-go .
COPY --from=builder /app/pass.txt .
CMD ["./auth-side-to-go"]
