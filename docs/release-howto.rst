how to make a release
=====================
1. Tag the release
2. Build the latest builder image:

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

This is a bit complicated, since it is a two-stage build. It will need you have
cloned the secrets folder containing the windows authenticode. You also have to
have wine (32 bits) installed in your host machine.

```
make package_win_in_docker
```

5. Build the OSX package: 

(TBD)

6. Build the debian package:

(TBD)
