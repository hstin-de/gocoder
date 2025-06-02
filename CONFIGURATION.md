# Configuration

Gocoder supports multiple configuration methods to customize database generation and server behavior. Configuration values are loaded in the following priority order:

1. Environment variables (highest priority)
2. JSON configuration file (`config.json`)
3. Default values (lowest priority)

## Configuration Methods

### Environment Variables

Set configuration using environment variables:

```bash
export PLANET=/path/to/germany-latest.osm.pbf
export WHOS_ON_FIRST=/path/to/whosonfirst-data-admin-latest.db
export WIKIMEDIA_IMPORTANCE=/path/to/wikimedia-importance.csv.gz
export OUTPUT=/path/to/output/germany.gpkg
export DATABASE=/path/to/database/germany.gpkg
export ENABLE_FORWARD=true
export ENABLE_REVERSE=true
export LANGUAGES=en,de,fr,es
export WIKIMEDIA_MAX_IMPORTANCE=500.0
```

### JSON Configuration File

Create a `config.json` file in your project root:

```json
{
  "languages": ["en", "de", "fr", "es", "it", "nl", "pt", "ru", "zh"],
  "wikimedia_max_importance": 500.0,
  "planet": "/data/germany-latest.osm.pbf",
  "whos_on_first": "/data/whosonfirst-data-admin-latest.db",
  "wikimedia_importance": "/data/wikimedia-importance.csv.gz",
  "output": "geocoder.gpkg",
  "database": "geocoder.gpkg",
  "enable_forward": true,
  "enable_reverse": true
}
```

### .env File Support

Gocoder automatically loads environment variables from a `.env` file if present:

```bash
# .env file
PLANET=/data/germany-latest.osm.pbf
WHOS_ON_FIRST=/data/whosonfirst-data-admin-latest.db
WIKIMEDIA_IMPORTANCE=/data/wikimedia-importance.csv.gz
OUTPUT=/app/database/germany.gpkg
DATABASE=/app/database/germany.gpkg
ENABLE_FORWARD=true
ENABLE_REVERSE=true
LANGUAGES=en,de,fr
WIKIMEDIA_MAX_IMPORTANCE=300.0
```

## Configuration Parameters

### Data Sources

#### `PLANET` / `planet`
- **Type**: String
- **Required**: Yes (for database generation)
- **Description**: Path to OpenStreetMap PBF file
- **Example**: `/data/germany-latest.osm.pbf`

#### `WHOS_ON_FIRST` / `whos_on_first`
- **Type**: String
- **Required**: Yes (for database generation)
- **Description**: Path to Who's On First SQLite database
- **Example**: `/data/whosonfirst-data-admin-latest.db`

#### `WIKIMEDIA_IMPORTANCE` / `wikimedia_importance`
- **Type**: String
- **Required**: Yes (for database generation)
- **Description**: Path to Wikimedia importance scores CSV file
- **Example**: `/data/wikimedia-importance.csv.gz`

### Output Configuration

#### `OUTPUT` / `output`
- **Type**: String
- **Default**: `geocoder.gpkg`
- **Description**: Output path for generated database file
- **Example**: `/app/database/germany.gpkg`

#### `DATABASE` / `database`
- **Type**: String
- **Default**: `geocoder.gpkg`
- **Description**: Path to database file for server runtime
- **Example**: `/app/database/germany.gpkg`

### Server Configuration

#### `ENABLE_FORWARD` / `enable_forward`
- **Type**: Boolean
- **Default**: `true`
- **Description**: Enable forward geocoding API endpoint
- **Values**: `true`, `false`

#### `ENABLE_REVERSE` / `enable_reverse`
- **Type**: Boolean
- **Default**: `true`
- **Description**: Enable reverse geocoding API endpoint
- **Values**: `true`, `false`

### Language and Data Processing

#### `LANGUAGES` / `languages`
- **Type**: String (comma-separated) / Array
- **Default**: `["en", "de", "fr", "es", "it", "nl", "pt", "ru", "zh"]`
- **Description**: Supported languages for geocoding results
- **Environment Example**: `LANGUAGES=en,de,fr,es`
- **JSON Example**: `"languages": ["en", "de", "fr"]`

#### `WIKIMEDIA_MAX_IMPORTANCE` / `wikimedia_max_importance`
- **Type**: Float
- **Default**: `500.0`
- **Description**: Scaling factor for Wikimedia importance scores (0-1) in internal ranking calculations. The original 0-1 importance scores are multiplied by this value and added to the internal ranking system.
- **Example**: `300.0` - scales importance scores up to 300 points in ranking

## Intermediate Files

The following intermediate files are automatically generated based on the output directory:

- **Bounding Boxes**: `{output_dir}/bounding_boxes.geo`
- **Countries**: `{output_dir}/countries.geojson`

These files are created during database generation and used for administrative boundary processing.

### Docker Configuration

When using Docker, mount your configuration and data directories:

```yaml
version: '3.8'
services:
  geocoder:
    image: ghcr.io/hstin-de/gocoder:latest
    ports:
      - "3000:3000"
    volumes:
      - ./data:/data
      - ./database:/app/database
      - ./config.json:/app/config.json  # Mount JSON config
    command: ["./geocoder", "server"]
```

## Storage Requirements

Database size varies significantly by geographic coverage:

- **Germany**: ~90MB
- **Europe**: ~1.2GB (estimated)
- **World**: ~2.8GB

Plan storage accordingly based on your coverage area and ensure sufficient disk space for both source data and generated databases.

## Performance Tuning

### Memory Configuration

- **Forward + Reverse**: ~2.5GB RAM
- **Reverse Only**: ~280MB RAM

Disable forward geocoding (`ENABLE_FORWARD=false`) if only reverse geocoding is needed to reduce memory usage.

### Importance Scaling

Adjust `WIKIMEDIA_MAX_IMPORTANCE` to control how much Wikimedia importance scores influence result ranking:

```bash
# Higher values give more weight to Wikipedia importance in ranking
WIKIMEDIA_MAX_IMPORTANCE=500.0  # Default - up to 500 points added to ranking
WIKIMEDIA_MAX_IMPORTANCE=300.0  # Lower influence - up to 300 points
WIKIMEDIA_MAX_IMPORTANCE=1000.0 # Higher influence - up to 1000 points
```