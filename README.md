# Bible CLI

A simple Go CLI tool for fetching Bible verses from the ESV API.

## Setup

1. Get a free API key from [ESV API](https://api.esv.org/)
2. Set your API key as an environment variable:
   ```bash
   export ESV_TOKEN='your_api_key_here'
   ```

## Build

```bash
go build -o bible-cli
```

## Usage

Get a random verse:
```bash
./bible-cli
```

Get a specific verse:
```bash
./bible-cli John 3:16
./bible-cli Psalm 23:1-6
```

## Install (Optional)

```bash
go install
# Now you can use it from anywhere:
bible-cli Romans 8:28
```