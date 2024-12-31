FROM golang:alpine
LABEL authors="vorticist"
WORKDIR /app
COPY . .
RUN go build -o the-account .
CMD ["/app/the-account"]

FROM golang:alpine
LABEL authors="vorticist"

# Set the working directory
WORKDIR /app

# Copy the source code
COPY . .

# Generate a time-based version code (letters only) and set it at build time
RUN VERSION_CODE=$(date +%s | sha256sum | tr -dc 'a-zA-Z' | head -c 8) && \
    echo "Version Code: $VERSION_CODE" && \
    go build -o the-account -ldflags="-X 'vortex.studio/account/internal/handlers.VersionCode=$VERSION_CODE'"

# Run the built application
CMD ["/app/the-account"]
