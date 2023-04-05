#!/bin/bash -e

# Install node modules
SRC_DIR="cmd/fits-api/assets/dependencies"
npm install --prefix $SRC_DIR

# Copy required files to assets/assets/dependencies folder
OUT_DIR="cmd/fits-api/assets/assets/dependencies"
mkdir -p $OUT_DIR

MODULE="@fortawesome/fontawesome-free"
mkdir -p "$OUT_DIR/$MODULE/css"
mkdir -p "$OUT_DIR/$MODULE/webfonts"
cp $SRC_DIR/node_modules/$MODULE/css/all.min.css "$OUT_DIR/$MODULE/css/all.min.css"
cp $SRC_DIR/node_modules/$MODULE/webfonts/fa-brands-400.woff2 "$OUT_DIR/$MODULE/webfonts/fa-brands-400.woff2"
cp $SRC_DIR/node_modules/$MODULE/webfonts/fa-brands-400.ttf "$OUT_DIR/$MODULE/webfonts/fa-brands-400.ttf"
cp $SRC_DIR/node_modules/$MODULE/webfonts/fa-regular-400.woff2 $OUT_DIR/$MODULE/webfonts/fa-regular-400.woff2
cp $SRC_DIR/node_modules/$MODULE/webfonts/fa-regular-400.ttf $OUT_DIR/$MODULE/webfonts/fa-regular-400.ttf
cp $SRC_DIR/node_modules/$MODULE/webfonts/fa-solid-900.woff2 "$OUT_DIR/$MODULE/webfonts/fa-solid-900.woff2"
cp $SRC_DIR/node_modules/$MODULE/webfonts/fa-solid-900.ttf "$OUT_DIR/$MODULE/webfonts/fa-solid-900.ttf"
cp $SRC_DIR/node_modules/$MODULE/webfonts/fa-v4compatibility.woff2 "$OUT_DIR/$MODULE/webfonts/fa-v4compatibility.woff2"
cp $SRC_DIR/node_modules/$MODULE/webfonts/fa-v4compatibility.ttf "$OUT_DIR/$MODULE/webfonts/fa-v4compatibility.ttf"

MODULE="bootstrap"
mkdir -p "$OUT_DIR/$MODULE"
cp $SRC_DIR/node_modules/$MODULE/dist/js/bootstrap.bundle.min.js "$OUT_DIR/$MODULE/bootstrap.bundle.min.js"
cp $SRC_DIR/node_modules/$MODULE/dist/js/bootstrap.bundle.min.js.map "$OUT_DIR/$MODULE/bootstrap.bundle.min.js.map"

MODULE="leaflet"
mkdir -p "$OUT_DIR/$MODULE"
cp $SRC_DIR/node_modules/$MODULE/dist/leaflet.js "$OUT_DIR/$MODULE/leaflet.js"
cp $SRC_DIR/node_modules/$MODULE/dist/leaflet.css "$OUT_DIR/$MODULE/leaflet.css"

MODULE="jquery"
mkdir -p "$OUT_DIR/$MODULE"
cp $SRC_DIR/node_modules/$MODULE/dist/jquery.min.js "$OUT_DIR/$MODULE/jquery.min.js"

MODULE="dygraphs"
mkdir -p "$OUT_DIR/$MODULE"
cp $SRC_DIR/node_modules/$MODULE/dygraph-combined.js "$OUT_DIR/$MODULE/dygraph-combined.js"
cp $SRC_DIR/node_modules/$MODULE/dygraph-combined.js.map "$OUT_DIR/$MODULE/dygraph-combined.js.map"

