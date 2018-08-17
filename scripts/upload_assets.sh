#!/bin/bash

CONFIG=$@

for line in $CONFIG; do
  eval "$line"
done

owner="snowdrop"
repo="k8s-supervisor"

AUTH="Authorization: token $github_api_token"

GH_API="https://api.github.com"
GH_REPO="$GH_API/repos/$owner/$repo"

WGET_ARGS="--content-disposition --auth-no-challenge --no-cookie"
CURL_ARGS="-LJO#"

TAG="v0.3.0"
GH_TAGS="$GH_REPO/releases/tags/$TAG"

BIN_DIR="./dist/bin/"
RELEASE_DIR="./dist/release"
APP="sb"

JSON='{"tag_name": "'"$TAG"'","target_commitish": "master","name": "'"$TAG"'","body": "'"$TAG"'-release","draft": false,"prerelease": false}'

#echo "Create Release"
#curl -H "$AUTH" \
#     -H "Content-Type: application/json" \
#     -d "$JSON" \
#     $GH_REPO/releases

# Read asset tags.
# response=$(curl -sH "$AUTH" $GH_TAGS)
#
# # Get ID of the asset based on given filename.
# eval $(echo "$response" | grep -m 1 "id.:" | grep -w id | tr : = | tr -cd '[[:alnum:]]=')
# [ "$id" ] || { echo "Error: Failed to get release id for tag: $tag"; echo "$response" | awk 'length($0)<100' >&2; exit 1; }
# echo "ID : $id"
#
# # upload assets
# for arch in `ls -1 $BIN_DIR/`;do
#     suffix=""
#     if [[ $arch == windows-* ]]; then
#         suffix=".exe"
#     fi
#     target_file=$RELEASE_DIR/$APP-$arch$suffix.tar.gz
#     content_type=$(file -b --mime-type $target_file)
#
#     GH_ASSET="https://uploads.github.com/repos/$owner/$repo/releases/$id/assets?name=$(basename $target_file)"
#
#     curl -i -H "Authorization: token $github_api_token" \
#             -H "Content-Type: application/gzip" \
#             -d @$target_file $GH_ASSET
# done