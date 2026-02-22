#!/usr/bin/env bash

SOURCE="$1"
OUTPUT="$2"
PRIVATE="false"        
CREATOR="mktorrent script"
PIECE_LENGTH="21"       # 2^21 = 2MB pieces

if [ -z "$SOURCE" ]; then
  echo "Usage: $0 <source_file_or_dir> [output.torrent]"
  exit 1
fi

if [ ! -e "$SOURCE" ]; then
  echo "Error: Source does not exist."
  exit 1
fi

if [ -z "$OUTPUT" ]; then
  BASENAME=$(basename "$SOURCE")
  OUTPUT="${BASENAME}.torrent"
fi

TRACKERS=(
  "https://tracker.openbittorrent.com:443/announce"
  "https://tracker.leechers-paradise.org:443/announce"
  "https://tracker.internetwarriors.net:443/announce"
  "https://tracker.opentrackr.org:443/announce"
  "http://tracker.torrent.eu.org:451/announce"
)

TRACKER_ARGS=()
for t in "${TRACKERS[@]}"; do
  TRACKER_ARGS+=("-a" "$t")
done

PRIVATE_FLAG=""
if [ "$PRIVATE" = "true" ]; then
  PRIVATE_FLAG="-p"
fi

echo "Creating torrent..."
echo "Source: $SOURCE"
echo "Output: $OUTPUT"

mktorrent \
  $PRIVATE_FLAG \
  -l "$PIECE_LENGTH" \
  -n "$(basename "$SOURCE")" \
  -o "$OUTPUT" \
  -s "$CREATOR" \
  "${TRACKER_ARGS[@]}" \
  "$SOURCE"

if [ $? -eq 0 ]; then
  echo "Torrent created successfully: $OUTPUT"
else
  echo "Failed to create torrent."
  exit 1
fi
