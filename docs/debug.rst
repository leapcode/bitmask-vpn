Troubleshooting
===============

This document contains some useful debug information.

OSX
---
If you're having troubles with old versions of RiseupVPN that did not have an
uninstaller, and the new installer is not cleanly replacing the previous
install, you might need to manually clean things up. You will need root access to
stop the privileged helper.

First, see if the helper is running:

.. code:: bash

  pgrep bitmask-helper

To stop it:

.. code:: bash

  sudo launchctl unload /Library/LaunchDaemons/se.leap.bitmask-helper.plist

To start it:

.. code:: bash

  sudo launchctl load /Library/LaunchDaemons/se.leap.bitmask-helper.plist
  sudo launchctl start /Library/LaunchDaemons/se.leap.bitmask-helper.plist

Check that it's running:

.. code:: bash

  pgrep bitmask-helper

Manually check that the web api is running, and that it reports a version that matches what you currently have installed:

.. code:: bash

  curl http://localhost:7171/version

Also, you can check that the path near the end of the file /Library/LaunchDaemons/se.leap.bitmask-helper.plist
matches the current path where you installed RiseupVPN.app.

Cleaning up
~~~~~~~~~~~
If you have things messed up and you want to completely delete the bitmask-helper:

.. code:: bash

  sudo launchctl unload /Library/LaunchDaemons/se.leap.bitmask-helper.plist
  sudo rm -rf /Library/LaunchDaemons/se.leap.bitmask-helper.plist

Make sure that "pgrep bitmask-helper" does not return any pid.

Now you can move /Applications/RiseupVPN.app to the Trash, and launch a
recent installer to get a clean install.

Windows
-------
In Windows you can use PowerShell to see if there's an old service Running (it
can be from RiseupVPN, CalyxVPN, LibraryVPN etc...).

.. code:: powershell

  PS C:\Users\admin> Get-Service bitmask-helper-v2

You can also stop it (needs admin)

.. code:: powershell

  PS C:\Users\admin> Stop-Service bitmask-helper-v2

