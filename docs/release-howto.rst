Release procedure
=====================
1. Bump the static release file in `pkg/version`. After a release, this should read something like `0.21.2+git`. This file is used to generate version strings from tarballs.

2. Tag the release
3. Build the latest builder image:

```
make builder_image
```

3. Build the snap package:

With everything ready on the docker image, this one should be built "in a snap"
(badum-tsss).

```
make package_snap_in_docker
```

4. Build the windows installer:

(TBD)

5. Build the OSX package: 

(TBD)

6. Build the debian package:

(TBD)

7. Upload builds, renew the *-latest* symlinks and their `lastver` files (important!)
