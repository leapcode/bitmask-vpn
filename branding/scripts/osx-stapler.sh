#!/bin/bash
# Notes to script notarization steps.
# Taken from https://oozou.com/blog/scripting-notarization-for-macos-app-distribution-38

# TODO: put pass in keychain

# 1. create dmb
hdiutil create -format UDZO -srcfolder yourFolder YourApp.dmg

# 2. send notarization request
requestInfo=$(xcrun altool --notarize-app \
   --file "YourApp.dmg" \
   --username "yourDeveloperAccountEmail@email.com" \
   --password "@keychain:notarization-password" \
   --asc-provider "yourAppleTeamID" \
   --primary-bundle-id "com.your.app.bundle.id")

current_status = "in progress"

while [[ "$currentStatus" == "in progress" ]]; do

sleep 15

statusResponse=$(xcrun altool --notarization-info "$uuid" \
    --username "yourDeveloperAccountEmail@email.com" \
    --password "@keychain:notarization-password")

# TODO change to python ---- ruby script ------------------------------------
# the response is a multiline string, with the status being on its own line
# using the format "Status: <status here>"
# Split each line into its own object in an array
response_objects = ARGV[0].split("\n")

# get line that contains the "Status:" text
status_line = response_objects.select { |data| data.include?('Status:') }[0]

# get text describing the status (should be either "in progress" or "success")
current_status = "#{status_line.split('Status: ').last}"

# respond with value
puts current_status
# -- end ruby script --------------------------------------------------------

current_status=$(ruby status.rb "$statusResponse")
done

if [[ "$current_status" == "success" ]]; then
  # staple notarization here
  xcrun stapler staple "YourApp.dmg"
else
  echo "Error! The status was $current_status. There were errors. Please check the LogFileURL for error descriptions"
  exit 1
fi
