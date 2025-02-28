#!/bin/zsh

padding=50 && convert -background none -density 300 lulm.svg -resize $((512-2*$padding))x$((512-2*$padding)) -gravity center -extent 512x512 macos/icon.png
