# Tor-Powered Web Content Scraper

This application is a privacy-oriented web scraping tool developed in **Go**. It connects to target websites through the Tor network **(SOCKS5 proxy)** to fetch and archive page content locally while maintaining anonymity.


## Prerequisites
To run this application, you must have a Tor client (Tor Browser or the standalone Tor Service) active on your system.

- Default Proxy Address: 127.0.0.1:9150

## Usage

### Define Targets

Create a file named **targets.yaml** in the root directory. Add the URLs you wish to scrape, one per line:


```yaml
  http://example.com
  http://check.torproject.org
  http://v27qk46u7...onion
```

### Run the Application

Install the necessary dependencies and start the program:

```bash
go mod tidy
go run main.go
```

##  Output Structure
The application stores every successful scrape within the outputs/ directory. Saved using the format outputs/YYYYMMDD_HHMMSS_domain.html. A full history of the process is kept in outputs/logs.txt.