#!/bin/sh
# sass sass/main.scss:static/css/main.css
./tailwindcss -i ./static/css/input.css -o ./static/css/output.css --minify
templ generate