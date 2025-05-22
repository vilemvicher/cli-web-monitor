# CLI Web Monitor

`cli-web-monitor` is a simple command-line application written in Go that monitors the availability and responsiveness of one or more web URLs.  

It periodically sends HTTP GET requests to each specified URL, collects performance statistics, and displays them in a real-time table in the terminal.

## Features

- Accepts one or more URLs via command-line arguments.
- Sends HTTP requests every 5 seconds.
- Each request has a 10-second timeout.
- Requests to each URL are sequential; requests to different URLs are concurrent.
- Outputs a real-time updated table showing:
    - URL
    - Response time (min / avg / max)
    - Response size (min / avg / max)
    - Success ratio (`OK/Total`, where 2xx and 3xx are considered successful)
- Clean shutdown via `CTRL+C`:
    - Waits for in-flight requests to finish.
    - Displays final stats.
- REST API for requests for specific url:
  - `/stats/{website}`
  - query for pagination - `/stats/{website}?page=2`

## Installation

- Requires Go SDK 1.24.0 installed with OS configuration in path
- Clone:
```bash
git clone https://github.com/vilemvicher/cli-web-monitor.git
```

## How to run and use

- App accepts 1-n of URLs to monitor as arguments.  
- From project root you can run for example:
```bash
go run .  https://www.google.com https://www.catchhotels.com https://seznam.cz 
```

## Testing
```bash
go test -v ./monitor
```
```bash
go test ./monitor -race
```

## Runtime analysis
```bash
go tool pprof main cpu.pprof
```
```bash
go tool trace trace.out
```

