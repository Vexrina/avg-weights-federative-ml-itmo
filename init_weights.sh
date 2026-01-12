#!/usr/bin/env bash
set -e

ALIAS=myminio
BUCKET=mybucket
KEY=weights/global/latest.pt
FILE=latest.pt

if mc stat "$ALIAS/$BUCKET/$KEY" >/dev/null 2>&1; then
  echo "Object already exists, skipping"
else
  echo "Object does not exist, uploading"
  mc cp "$FILE" "$ALIAS/$BUCKET/$KEY"
fi
