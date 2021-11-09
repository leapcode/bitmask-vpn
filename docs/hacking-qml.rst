QML best practices
==================
* https://github.com/Furkanzmc/QML-Coding-Guide/blob/master/README.md
* lint your qml files::

  make qmllint

Debugging
---------
In windows you need to add some flags to obtain QML debug:

  QT_FORCE_STDERR_LOGGING=1 QT_LOGGING_DEBUG=1 ./riseup-vpn.exe
