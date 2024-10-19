# Release procedure

## Prepare source code repo for release

1. Generate the changelog and update the `CHANGELOG` file

```
$ git log --format="- %s" <last_release_tag>..HEAD
```
2. Open a Merge request with the above change
3. Create an annotated tag for the release version, the version for the app is taken from the o/p of `git desribe`

```
# tag should point to the commit that updated the CHANGELOG file
$ git tag -a 0.24.8 HEAD
```

## Build Installers for Windows and MacOS

### Steps to build the windows installer (needs Windows 10 or higher):

1. Generate the installer `.exe` file

```
$ make vendor # make sure to set the PROVIDER env variable to the correct provider
$ make build
$ make installer
```
2. Sign the installer:

```
PS> signtool sign /f .\leap.pfx /tr http://timestamp.digicert.com /td SHA256 /fd SHA256 /p <password_for_cert> <path_to_installer.exe>

```

### Steps to build the MacOS installer (needs MacOS 12 or higher):

1. Generate the installer `.app` file

```
$ make vendor # make sure to set the PROVIDER env variable to the correct provider
$ make build
$ make installer
```

2.Sign the MacOS installer:

```
$ export CODESIGN_IDENTITY=<codesign_id>
$ codesign --sign "${CODESIGN_IDENTITY}" --options runtime --timestamp --force <path_to_installer.app/Content/MacOS/installer_executable>

```
3. Create DMG to upload for Apple notarization

```
$ mkdir -p build/installer/out && cp -R build/installer/<installer.app> build/installer/out
$ cd build/installer
$ hdiutil create -volname <installer_name> -srcfolder out -ov -format UDZO <output_dmg_name.dmg>
```

4. Upload DMG for notarization

```
$ export APP_PASSWORD=<app_password>
$ xcrun notarytool submit --verbose --apple-id=<appleid> --team-id=<teamid> --password ${APP_PASSWORD} --wait --timeout 30m <path_to_dmg>

# To get logs or the notarization response for debugging
$ xcrun notarytool logs <notarization_id> --apple-id=<appleid> --team-id=<teamid> --password ${APP_PASSWORD}
```

>**IMPORTANT:** Upload builds, renew the *-latest* symlinks and their `lastver` files

>**NOTE:** Update packages for Ubuntu in the [leapcodes PPA](./build-ppa.md)

