#!/bin/bash
set -e

cd "$(dirname "$0")"

APPS=("system" "play" "upload")

for app in "${APPS[@]}"; do
  echo "编译 $app ..."
  cd "app/$app"
  go build -o "./$app" .
  cd ../..
  echo "$app 编译完成 -> app/$app/$app"
done

echo "全部编译完成"
