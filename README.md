# Concurrent Web Content Aggregator

A high-performance web content aggregator built in Go that fetches, processes, and normalizes data from multiple sources concurrently.

## Features

- **Concurrent Processing**: Efficiently fetch content from multiple sources using Go's goroutines and channels
- **Smart Rate Limiting**: Respect each domain's robots.txt and implement per-domain rate limiting
- **Configurable Behavior**: Easily adjust concurrency limits, timeout settings, and retry policies
- **Multiple Output Formats**: Export aggregated data as JSON, CSV, or view in a web interface
- **Resilient Error Handling**: Graceful recovery from network issues and transient failures
- **Content Filtering**: Define specific data to extract from each source
- **Data Normalization**: Transform varied content structures into a standardized format

## Requirements

- Go 1.18 or higher
- Docker (optional, for containerized deployment)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/CyberwizD/Concurrent-Web-Content-Aggregator.git
cd Concurrent-Web-Content-Aggregator

# Install dependencies
go mod download

# Build the binary
go build -o aggregator ./cmd/aggregator
```

### Using Docker

```bash
# Build the Docker image
docker build -t Concurrent-Web-Content-Aggregator .

# Run the container
docker run -v $(pwd)/configs:/app/configs -v $(pwd)/output:/app/output git remote add origin Concurrent-Web-Content-Aggregator
```

## Configuration

The application is configured using YAML files in the `configs` directory:

### config.yaml

```yaml
app:
  concurrency:
    max_fetchers: 10
    max_parsers: 5
  timeouts:
    request: 10s
    total: 5m
  retry:
    max_attempts: 3
    backoff_initial: 1s
    backoff_factor: 2.0
  output:
    format: json
    destination: file
    file_path: "./output/results.json"
```

### sources.yaml

```yaml
sources:
  - name: example_news
    url: "https://example.com/news"
    rate_limit:
      requests_per_minute: 30
    parser: html
    selectors:
      title: ".article-title"
      content: ".article-body"
      date: ".publish-date"
  
  - name: another_source
    url: "https://anothersource.com/api/articles"
    rate_limit:
      requests_per_minute: 60
    parser: json
    mappings:
      title: "headline"
      content: "body"
      date: "publishedAt"
```

## Usage

### Command Line Interface

```bash
# Run with default configuration
./aggregator

# Specify custom config files
./aggregator --config ./my-configs/config.yaml --sources ./my-configs/sources.yaml

# Limit to specific sources
./aggregator --sources-filter "example_news,another_source"

# Set output format
./aggregator --output json

# Enable debug logging
./aggregator --log-level debug
```

### API Service

Start the API server:

```bash
./aggregator --serve --port 8080
```

Then access the API:

```bash
# Fetch aggregated content
curl http://localhost:8080/api/content

# Start a new aggregation job
curl -X POST http://localhost:8080/api/aggregate
```

### Web Interface

Start the web server:

```bash
./aggregator --web --port 8080
```

Then access the web interface at `http://localhost:8080`

## Project Structure

```
content-aggregator/
├── cmd/                      # Application entry points
│   └── aggregator/           # Main command
├── configs/                  # Configuration files
├── internal/                 # Internal packages
│   ├── fetcher/              # HTTP request handling
│   ├── parser/               # Content parsing
│   ├── model/                # Data models
│   ├── coordinator/          # Concurrency management
│   ├── normalizer/           # Data normalization
│   └── aggregator/           # Main orchestration logic
├── pkg/                      # Reusable packages
│   ├── config/               # Configuration loading
│   └── util/                 # Utilities
├── web/                      # Web interface
└── output/                   # Default output directory
```

## How It Works

1. The **Coordinator** reads the configuration and initializes the system
2. It creates worker pools for fetching and parsing
3. The **Fetcher** pool makes HTTP requests with respect to rate limits and robots.txt
4. The **Parser** pool extracts structured data from responses
5. The **Normalizer** transforms data into a standard format
6. The **Aggregator** combines data from all sources
7. The **Output Handler** delivers results to the configured destination

## Extending

### Adding a New Parser

1. Create a new file in `internal/parser/` implementing the Parser interface
2. Register your parser in `internal/parser/parser.go`
3. Add the parser type to your source configuration

### Adding a New Output Format

1. Create a new implementation in `internal/aggregator/output.go`
2. Register your output format in the configuration system

## Benchmark Results

| Configuration | Sources | Total Time | Memory Usage |
|---------------|---------|------------|--------------|
| Default       | 10      | 2.5s       | 45MB         |
| High Concurrency | 10   | 1.2s       | 120MB        |
| Many Sources  | 100     | 12s        | 150MB        |

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [goquery](https://github.com/PuerkitoBio/goquery) for HTML parsing
- [robotstxt](https://github.com/temoto/robotstxt) for robots.txt parsing
- [go-rate](https://github.com/beefsack/go-rate) for rate limiting
