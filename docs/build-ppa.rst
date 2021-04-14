ppa howto
=========

* Add changes to changelog (bump native dot-version, change release)
* Upload changes file

.. code:: bash

  debuild -i -S
  dput --force ppa:kalikaneko/ppa ../riseup-vpn_0.21.2.2_source.changes

