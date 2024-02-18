# Web Parser

Web Parser is a Go application for counting incoming requests to a web server
within a specified time window. It provides functionality for handling HTTP
requests, tracking request counts, and persisting the counter state to a file.

## Features

- HTTP server for receiving incoming requests
- Counting requests within a specified time window
- Saving and loading counter state to/from a file
- Graceful shutdown handling
- Command-line interface for configuring server parameters

## Usage

To use Web Parser, follow these steps:

1. Clone the repository:

```bash
$ git clone https://github.com/chaitrashreegs/webparser.git
cd webparser
```

2. Build the application:

```bash
go build -o webparser cmd/main.go
```

**Note:** Makefile already has commands to build,test,run application

3. Run the application with appropriate command-line arguments:

```bash
./webparser --address <address> --port <port> --file-path <file-path> --window-size <window-size> --precision <precision>
```

Replace:
- `address` with the IP address on which the web server runs,
- `port` with the port number,
- `file-path` with the path to store the counter data file,
- `window-size`with the duration of the time window for counting requests, and
- `precision` with the precision level for calculating the window index.

4. Send requests to the web server to count them.

```bash
$curl  http://127.0.0.1:8090/counter
Total requests in the last 60 seconds: 1
```

5. Configuration Options

- `address`: IP address on which the web server runs (default: 0.0.0.0)
- `port`: Port number on which the web server runs (default: 8090)
- `file-path`: Path to store the counter data file (default: ./counter.gob)
- `window-size`: Duration of the time window for counting requests (default: 60 seconds)
- `precision`: Precision level for calculating the window index (default: 1 second)

6. TODO/Enhancements

- Dynamic buffer size allocation
- Support HTTPS server
- Add middlewares for logging and recovery
