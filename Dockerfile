# Step 1: Build Stage
# Use the official Golang image to create a build artifact.
# This image is based on Debian and includes the Go toolchain.
FROM golang:1.19 as builder

# Set the working directory inside the container.
WORKDIR /app

# Copy go.mod and go.sum to download dependencies.
# Copying just these files first allows Docker to cache the downloaded dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code.
COPY . .

# Build the Go app as a static binary.
# -o specifies the output binary name.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Step 2: Runtime Stage
# Use a smaller, "distroless" base image for the runtime.
FROM gcr.io/distroless/base-debian10

# Copy the pre-built binary from the previous stage.
COPY --from=builder /app/main /

# Run the binary.
CMD ["/main"]
