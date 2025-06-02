#!/bin/bash
# download-data.sh - Download required datasets for Gocoder

mkdir -p data
cd data

# Download the global data
wget https://data.geocode.earth/wof/dist/sqlite/whosonfirst-data-admin-latest.db.bz2
wget https://nominatim.org/data/wikimedia-importance.csv.gz

# Extract the Who's On First database
bunzip2 whosonfirst-data-admin-latest.db.bz2

# Download the osm file (Germany)
wget https://download.geofabrik.de/europe/germany-latest.osm.pbf