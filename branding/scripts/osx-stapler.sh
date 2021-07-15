#!/bin/bash
# Notes to script notarization steps.
# To be called from the root folder.
# Taken from https://oozou.com/blog/scripting-notarization-for-macos-app-distribution-38

# TODO: put pass in keychain
# --password "@keychain:notarization-password"

USER=info@leap.se

requestInfo=$(xcrun altool --notarize-app \
	-t osx -f build/installer/${APPNAME}-installer-${VERSION}.zip \
	--primary-bundle-id="se.leap.bitmask.${TARGET}" \
	-u ${USER} \
	-p ${OSXAPPPASS})

uuid=$(python3 branding/scripts/osx-staple-uuid.py $requestInfo)

current_status="in progress"

while [[ "$currentStatus" == "in progress" ]]; do

sleep 15

statusResponse=$(xcrun altool --notarization-info "$uuid" \
    --username ${USER} \
    --password ${OSXAPPPASS})
current_status=$(python3 branding/scripts/osx-staple-status.py $statusResponse)
done


if [[ "$current_status" == "success" ]]; then
  # staple notarization here
  xcrun stapler staple build/installer/${APPNAME}-installer-${VERSION}.app
  create-dmg deploy/${APPNAME}-${VERSION}.dmg build/installer/${APPNAME}-installer-${VERSION}.app
else
  echo "Error! The status was $current_status. There were errors. Please check the LogFileURL for error descriptions"
  exit 1
fi
