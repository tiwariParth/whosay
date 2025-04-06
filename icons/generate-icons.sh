#!/bin/bash

# Check if Inkscape is installed (used for SVG to PNG conversion)
if ! command -v convert &> /dev/null; then
    echo "ImageMagick not found. Installing..."
    sudo apt update
    sudo apt install -y imagemagick
fi

# Create PNGs in different sizes from the SVG logo
echo "Generating icon files..."

# Generate the different sizes needed for Flatpak
convert -background none -resize 64x64 whosay-logo.svg whosay-64.png
convert -background none -resize 128x128 whosay-logo.svg whosay-128.png
convert -background none -resize 256x256 whosay-logo.svg whosay-256.png
convert -background none -resize 512x512 whosay-logo.svg whosay-512.png

echo "Done! Generated icon files:"
ls -l whosay-*.png

echo "These icons are ready to use with your Flatpak package."
