osx build
=============

Cheat-sheet
------------------

tl;dr:

.. code:: bash

  export RELEASE=yes
  export OSXAPPPASS=my-apple-app-pass
  make clean && make vendor && make build
  make installer
  make sign_installer
  make notarize_installer
  make notarize_staple
  make create_dmg

Sign the release
-------------------

in recent osx releases, it's not ok to just sign the insallers anymore. you
have to sign and then notarize with their service. here are some notes that use
ad-hoc targets in the main makefile, but we should keep an eye on any future
integration of this process in the more or less official Qt tools (QTIFW).

First, we build the regular installer (use RELEASE=yes to do a codesign step
with macqtdeploy, note that this increases build time considerably):

.. code:: bash

  make build
  RELEASE=yes make installer
  make sign_installer

Now we export the app-specific password and we proceed to notarization. If you
don't know what is this pass, you can create one in your Apple developer
account. Contact their friendly support for more info, but don't expect they
understand you do not really own any Apple Hardware. Sense of humor is not
universal.

Security -> App-specific passwords -> Generate
If you need to revoke these tokens, click on 'view history'.

https://appleid.apple.com/account/manage

According to https://developer.apple.com/documentation/xcode/notarizing_macos_software_before_distribution/customizing_the_notarization_workflow:

To avoid including your password as cleartext in a script, you can provide a
reference to a keychain item, as shown in the previous example. This assumes
the keychain holds a keychain item named AC_PASSWORD with an account value
matching the username AC_USERNAME.

.. code:: bash

  export OSXAPPPASS=my-apple-app-pass
  make notarize_installer

Between the output of the last command, you will get a Request UUID. You should pass that request uid in the appropriate 
environment variable to check the status of the notarization process. Obviously, since the recent changes in Apple policies,
you need to be in posession of a valid membership

.. code:: bash

  altool[5281:91963] No errors uploading 'build/installer/RiseupVPN-installer-0.20.4-175-gee4eb90.zip'.
  RequestUUID = fe9a4324-bdcb-4c52-b857-f089dc904695
  
  OSXMORDORUID=fe9a4324-bdcb-4c52-b857-f089dc904695 make notarize_check
  xcrun altool --notarization-info fe9a4324-bdcb-4c52-b857-f089dc904695 -u "info@leap.se" -p my-apple-app-pass
  2020-12-11 22:21:59.940 altool[5787:96428] No errors getting notarization info.
  
     RequestUUID: fe9a4324-bdcb-4c52-b857-f089dc904695
            Date: 2020-12-11 21:13:10 +0000
          Status: success
      LogFileURL: https://osxapps-ssl.itunes.apple.com/itunes-assets/Enigma114/v4/0f/c9/1e/0fc91e64-2c9f-74e5-3cf6-96b8f3bf7170/developer_log.json?accessKey=1607916119_6680812212684569509_nLlPw6tYxTSiWZfFTb0atP9zZ3CEGDfW0btWV51xhjWHiCFqBt%2BneXd5Vp40eQCSx8e1W5PYCIe2db7JGbhoTeJsYxl7UmYssRvYpTxYJl8z90uwB9jkbS1fsd7niaAn%2BQs7xHdv%2BB9jaKQI8LJ%2BwYY8RPq1QaeCJxBIdeG44DY%3D
     Status Code: 0
  Status Message: Package Approved

If everything is ok, now you can finish the process, stapling the notarization info and creating the dmg.

.. code:: bash

  make notarize_staple
  make create_dmg

If everything went well, you should have a .dmg for your release under the `deploy` folder.

.. code:: bash

  created: /Users/admin/leap/bitmask-vpn/deploy/RiseupVPN-0.20.4-175-gee4eb90.dmg
