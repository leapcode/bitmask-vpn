How to create VMs for building and testing
============================================================

For Debian and Ubuntu, we want to support the two latest LTS (long term support) releases. For each release, we need to build packages for each distro.

Release overview

- https://www.debian.org/releases/
- https://www.releases.ubuntu.com/



Download and setup VMs
-------------------------

To get VMs, you can use:

- quickemu https://github.com/quickemu-project/quickemu
- create Virtualbox VMs by hand
- vagrant


.. code:: bash

  mkdir -p ~/leap/vms & cd ~/leap/vms
  quickget xubuntu 24.04
  quickget xubuntu 22.04
  quickget debian 12.5.0 xfce
  quickget debian 11.9.0 xfce
  
  # start vm and install OS (with --display spice you have a shared clipboard)
  quickemu --vm xubuntu-24.04.conf --display spice


Install tools & dependencies
---------------------------------

.. code:: bash
  
   # install base
  sudo apt-get update
  sudo apt-get dist-upgrade
  sudo apt-get install -y firefox featherpad tmux vim git make fd-find ripgrep magic-wormhole
  
  # install make deps (check branding/templates/debian/control-template)
  sudo apt install golang make pkg-config g++ git libqt6svg6-dev qt6-tools-dev qt6-tools-dev-tools qt6-base-dev libqt6qml6 qt6-declarative-dev dh-golang libgl-dev  qt6-5compat-dev qt6-declarative-dev-tools qt6-l10n-tools
  
  # install deps (check branding/templates/debian/control-template)
  sudo apt install libqt6core6 libqt6gui6 libqt6qml6 libqt6widgets6 libstdc++6 libqt6svg6 qml6-module-qtquick qml6-module-qtquick-controls qml6-module-qtquick-dialogs qml6-module-qtquick-layouts qml6-module-qtqml-workerscript qml6-module-qtquick-templates qml6-module-qt-labs-settings qml6-module-qtquick-window qml6-module-qt-labs-platform qml6-module-qtcore qml6-module-qt5compat-graphicaleffects openvpn policykit-1-gnome
  
  sudo ln -s $(qmake6 -query "QT_INSTALL_BINS")/lrelease /usr/local/bin/lrelease


If go < 1.20 (Debian 12)
---------------------------------

The go package of Debian 12 is too old (< 1.20). Please install the `golang-go` package of `bookworm-backports`. 

- https://backports.debian.org/Instructions/
- https://packages.debian.org/bookworm-backports/golang/golang


Build desktop client
---------------------------------

You can override the version with env VERSION= (required for all targets)

.. code:: bash
  
  git clone https://0xacab.org/leap/bitmask-vpn.git
  cd bitmask-vpn
  sudo make depends
  PROVIDER=bitmask make vendor
  QMAKE=qmake6 make build

  # install helper on Linux (only for manual testing, gets installed by the pckage)
  build/qt/release/bitmask-vpn --install-helpers


Build deb package
---------------------------------

.. code:: bash
  
  # create debian package (you can also set the version with VERSION=)
  make package_deb
  sudo dpkg -i  deploy/bitmask-vpn_0.24.5-66-gd52c528_amd64.deb
