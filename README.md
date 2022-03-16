# ClamAV API

A stateless Go server utilizing ClamAV to scan files sent by HTTP request. The server responds with the scan results. It can for example be used as microservice to detect malicious files uploaded by users. The Docker container is based on the [official ClamAV Docker image](https://github.com/Cisco-Talos/clamav).

## Local Setup

### Prerequisites

To build the Docker image locally you need the following:

- Docker
- Go

### Build locally

Clone the respository with:

```
git clone https://github.com/jz222/clamav-api.git
```

Build the Docker image by running the build script with:

```
bash build.sh
```

Run the Docker container with:

```
docker run -it -p 8080:8080 clamav-api
```

## Docker Image

If you don't want to build the image locally, you can pull it directly from Docker Hub by running:

```
docker run -it -p 8080:8080 jz222/clamav-api:2.0.1
```

## Updates

The Docker image should be build regularly to update ClamAV.

## API

The server exposes the POST endpoint `/scan`. It accepts multipart form data and expects the field `file` with any file with a size of maximum 40mb.

```
curl --form file='@example.pdf' http://localhost:8080/scan
```

**Example Response: No Malicious File Detected**

```json
{
   "success": true,
   "data": {
      "isMalicious": false,
      "result": "/root/uploads/121736192-Untitled.txt: OK\n\n----------- SCAN SUMMARY -----------\nKnown viruses: 8608180\nEngine version: 0.104.2\nScanned directories: 0\nScanned files: 1\nInfected files: 0\nData scanned: 0.27 MB\nData read: 0.13 MB (ratio 2.12:1)\nTime: 23.505 sec (0 m 23 s)\nStart Date: 2022:03:16 22:20:51\nEnd Date:   2022:03:16 22:21:14\n"
   }
}
```

**Example Response: Malicious File Detected**

```json
{
   "success": true,
   "data": {
      "isMalicious": true,
      "detectedVirus": "Win.Test.EICAR_HDB-1",
      "result": "/root/uploads/786884069-eicar_com.zip: Win.Test.EICAR_HDB-1 FOUND\n\n----------- SCAN SUMMARY -----------\nKnown viruses: 8608180\nEngine version: 0.104.2\nScanned directories: 0\nScanned files: 1\nInfected files: 1\nData scanned: 0.00 MB\nData read: 0.00 MB (ratio 0.00:1)\nTime: 22.922 sec (0 m 22 s)\nStart Date: 2022:03:16 22:16:06\nEnd Date:   2022:03:16 22:16:29\n"
   }
}
```

## Environment Variables

| Name | Description                                              | Default |
|------|----------------------------------------------------------|---------|
| PORT | Determines the port on which the server should listen on | 8080    |