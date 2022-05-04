FROM golang AS build

WORKDIR /app

# Install dependencies.
COPY go.* ./
RUN go mod download

# Copy source files and build application.
COPY . /app
RUN make && mv bin/* /app/bin/k3se

# Create minimal container image.
FROM gcr.io/distroless/base:nonroot AS run
COPY --from=build /app/bin/k3se /k3se
ENTRYPOINT [ "/k3se" ]
