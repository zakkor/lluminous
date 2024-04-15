#!/bin/bash
npm run build && rsync -r -avz -e ssh ./dist/* root@137.184.193.108:/var/www/html/lluminous.chat