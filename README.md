# Gocoder

Gocoder is a fast, scalable geocoding service written in Go, capable of converting place names to coordinates (forward geocoding) and coordinates to place names (reverse geocoding). It processes data from OpenStreetMap, Who's On First administrative boundaries, and Wikimedia importance scores, creating a highly optimized geographic database.

## Architecture

### Key Data Structures

Gocoder utilizes several custom made data structures to ensure efficient geocoding:

* **Trie:** Efficient prefix-based matching for exact place names.
* **Fuzzy Index:** Approximate matching using n-gram indexing and Levenshtein distance.
* **KD-Tree:** Spatial indexing for reverse geocoding.
* **Administrative Boundaries:** R-tree indexing for quick administrative lookups.

### Supported Languages

* Supports all languages available in OpenStreetMap datasets.
* Automatic fallback when language-specific data is unavailable.

## Installation

### Prerequisites

* Go 1.23.3 or newer
* Docker (optional)
* osmium-tool
* Sufficient storage (varies by geographic coverage)

### Building from Source

```bash
git clone https://github.com/hstin-de/gocoder
cd gocoder
go mod download
CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-w -s' -o gocoder .
```

### Docker Setup

Docker images are published on [GitHub Container Registry (GHCR)](https://ghcr.io/hstin-de/gocoder).

```bash
docker pull ghcr.io/hstin-de/gocoder:latest
```

## Usage

### Step 1: Generate Database

#### Download Required Data

Download the required data manually:

* **OpenStreetMap Data** (example Germany): [Geofabrik](https://download.geofabrik.de/europe/germany-latest.osm.pbf)
* **Who's On First Database**: [geocode.earth](https://data.geocode.earth/wof/dist/sqlite/whosonfirst-data-admin-latest.db.bz2) (unzip first)
* **Wikimedia Importance Scores**: [Nominatim](https://nominatim.org/data/wikimedia-importance.csv.gz)

or use the provided download script to automatically fetch and extract all required datasets:

```bash
# Make script executable and run
chmod +x download-data.sh
./download-data.sh
```

#### Generate Database

Run database generation using Docker:

```bash
docker run --rm \
  -v $(pwd)/data:/data \
  -v $(pwd)/database:/app/database \
  -e PLANET=/data/germany-latest.osm.pbf \
  -e WHOS_ON_FIRST=/data/whosonfirst-data-admin-latest.db \
  -e WIKIMEDIA_IMPORTANCE=/data/wikimedia-importance.csv.gz \
  -e OUTPUT=/app/database/germany.gpkg \
  ghcr.io/hstin-de/gocoder:latest ./gocoder generate
```

For configuration parameters, refer to the [Configuration](CONFIGURATION.md) file.

Generation includes:

1. Administrative boundary extraction.
2. Who's On First area processing.
3. OpenStreetMap node enrichment.
4. Index building (Trie, KD-Tree).
5. Binary database serialization.

### Step 3: Start Server

**Docker Compose:**

```yaml
version: '3.8'

services:
  gocoder:
    image: ghcr.io/hstin-de/gocoder:latest
    ports:
      - "3000:3000"
    volumes:
      - ./database:/app/database
    environment:
      - DATABASE=/app/database/germany.gpkg
      - ENABLE_FORWARD=true
      - ENABLE_REVERSE=true
    command: ["./gocoder", "server"]
    restart: unless-stopped
```

**Docker Run:**

```bash
docker run -d \
  -p 3000:3000 \
  -v $(pwd)/database:/app/database \
  -e DATABASE=/app/database/germany.gpkg \
  -e ENABLE_FORWARD=true \
  -e ENABLE_REVERSE=true \
  ghcr.io/hstin-de/gocoder:latest
```

## API Reference

### Forward Geocoding

* **Endpoint**: `GET /`

**Parameters:**

* `q`: Search query (required).
* `max`: Max results (default: 10).
* `complete`: Return all results (default: false).
* `cache`: Enable caching (default: true).
* `lang`: Language preference.

**Example:**

```bash
curl "http://localhost:3000/?q=Berlin&max=5&lang=en"
```

### Reverse Geocoding

* **Endpoint**: `GET /reverse`

**Parameters:**

* `lat`: Latitude (required).
* `lng`: Longitude (required).
* `lang`: Language preference.

**Example:**

```bash
curl "http://localhost:3000/reverse?lat=52.517&lng=13.389&lang=en"
```

### Node Lookup

* **Endpoint**: `GET /node/:id`

**Example:**

```bash
curl "http://localhost:3000/node/240109189?lang=en"
```

## Performance

* Worldwide search: Typically <10ms, cached ~1ms.
* Exact Match: Trie lookup, O(m).
* Fuzzy Search: O(n*k) complexity.
* Reverse Geocoding: KD-tree, average O(log n).
* RAM usage: Forward and reverse ~2.5GB; Reverse only ~280MB.

## Licensing

This service processes data subject to licenses from OpenStreetMap, Who's On First, and Wikimedia. Ensure compliance with their respective licenses.

gocoder is released under the [MIT License](LICENSE).