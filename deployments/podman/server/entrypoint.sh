#!/bin/sh
set -e

echo "Running installation step..."
/usr/local/bin/gophkeeper -d -install

echo "Starting main server..."
exec /usr/local/bin/gophkeeper -d