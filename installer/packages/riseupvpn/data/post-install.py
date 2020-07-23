#!/usr/bin/env python

import os
import sys
import subprocess

HELPER = "bitmask-helper"
HELPER_PLIST = "/Library/LaunchDaemons/se.leap.bitmask-helper.plist"

def main():
    _dir = os.path.dirname(os.path.realpath(__file__))
    log = open(os.path.join(_dir, 'post-install.log'), 'w')
    log.write('Checking for admin privileges...')

    _id = os.getuid()
    if _id != 0:
      err  = "error: need to run as root. UID: %s\n" % str(_id)
      logErr(log, msg)
    
    # failure: sys.exit(1)
    
    if isHelperRunning():
        log.write("Trying to stop bitmask-helper...")
	# if this fail, we can check if the HELPER_PLIST is there
        ok = unloadHelper()
        log.write("success: %s \n" % str(ok))

    ok = makeHelperExecutable()
    log.write("chmod +x helper: %s \n" % str(ok))

    # 3. cp se.leap.bitmask-helper.plist /Library/LaunchDaemons/
    copyLaunchDaemon()

    # 4. launchctl load /Library/LaunchDaemons/se.leap.bitmask-helper.plist
    launchHelper()

    # 5. chown admin:wheel /Applications/$applicationName.app/Contents/helper # is this the folder?
    grantPermissionsOnLogFolder()
    
    # all good
    log.write('post-install script: done')
    sys.exit(0)


def logErr(log, msg):
    log.write(err)
    sys.exit(1)

def isHelperRunning():
    ps = _getProcessList()
    return HELPER in ps 

def unloadHelper():
    out = subprocess.call(["launchctl", "unload", HELPER_PLIST])
    return out == 0

def makeHelperExecutable():
    out = subprocess.call(["chmod", "+x", HELPER])
    return out == 0

def copyLaunchDaemon():
    pass

def launchHelper():
    pass

def grantPermissionsOnLogFolder():
    pass

def _getProcessList():
    _out = []
    output = subprocess.Popen(["ps", "-ceA"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    stdout, stderr = output.communicate()
    for line  in stdout.split('\n'):
        cmd = line.split(' ')[-1]
        _out.append(cmd.strip())
    return _out

if __name__ == "__main__":
    main()
