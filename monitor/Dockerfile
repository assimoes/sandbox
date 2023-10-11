# Start from the official Go image as a builder
FROM golang:1.18 AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files to the workspace
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code from the current directory to the workspace
COPY monitor/ monitor/
COPY shared/ shared/
COPY logger/ logger/

# Build the Go app as a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o monitor/monitor ./monitor

# Use the scratch image for the final stage
FROM scratch

# Copy the statically linked binary from the builder stage
COPY --from=builder /app/monitor/monitor /monitor

# Set the binary as the entrypoint
ENTRYPOINT ["/monitor"]
