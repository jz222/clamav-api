env CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o build/clamav-api .
docker build -t clamav-api .