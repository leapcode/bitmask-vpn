ppa howto
=========

* Add changes to changelog (bump native dot-version, change release)
* Upload changes file

.. code:: bash

  debuild -i -S
  dput --force ppa:kalikaneko/ppa ../riseup-vpn_0.21.2.2_source.changes

Using kali's ppa
----------------

.. code:: bash

  sudo gpg --homedir=/tmp --no-default-keyring --keyring /usr/share/keyrings/kali-ppa-archive-keyring.gpg --keyserver keyserver.ubuntu.com --recv-keys 0xbe23fb4a0e9db36ecb9ab8be23638bf72c593bc1
  sudo add-apt-repository ppa:kalikaneko/ppa
  sudo apt update
  sudo apt install riseup-vpn

