#!/usr/bin/env python3

# Post installation script for BitmaskVPN.
# Please note that this installation will install ONE single helper with administrative privileges.
# This means that, for the time being, you can only install ONE of the BitmaskVPN derivatives at the same time.
# This might change in the future.

import glob
import os
import shutil
import sys
import subprocess
import time

HELPER = "bitmask-helper"
HELPER_PLIST = "/Library/LaunchDaemons/se.leap.bitmask-helper.plist"

_dir = os.path.dirname(os.path.realpath(__file__))
_appdir = glob.glob('{}/*VPN.app'.format(_dir))[0]

def main():
    log = open(os.path.join(_dir, 'post-install.log'), 'w')
    log.write('Checking for admin privileges...\n')

    _id = os.getuid()
    if _id != 0:
      err  = "ERROR: need to run as root. UID: %s\n" % str(_id)
      log.write(err)
      sys.exit(1)
    
    if isHelperRunning():
        log.write("Trying to stop bitmask-helper...\n")
	# if this fail, we can check if the HELPER_PLIST is there
        ok = unloadHelper()
        log.write("success: %s \n" % str(ok))

    ok = fixHelperOwner(log)
    log.write("chown helper: %s \n" % str(ok))

    log.write("Copy launch daemon...\n")
    copyLaunchDaemon()

    log.write("Trying to launch helper...\n")
    out = launchHelper()
    log.write("result: %s \n" % str(out))

    grantPermissionsOnLogFolder()
    
    # all done
    log.write('post-install script: done\n')
    sys.exit(0)

def isHelperRunning():
    ps = _getProcessList()
    return HELPER in ps 

def unloadHelper():
    out = subprocess.call(["launchctl", "unload", HELPER_PLIST])
    time.sleep(0.5)
    out2 = subprocess.call(["pkill", "-9", "bitmask-helper"])  # just in case
    time.sleep(0.5)
    return out == 0

def fixHelperOwner(log):
    path = os.path.join(_appdir, HELPER)
    try:
        os.chown(path, 0, 0)
    except OSError as exc:
        log.write(str(exc))
        return False
    return True

def copyLaunchDaemon():
    appDir = os.path.join(_dir, _appdir)
    plist = "se.leap.bitmask-helper.plist"
    plistFile = os.path.join(appDir, plist)
    escapedPath = appDir.replace("/", "\/")
    subprocess.call(["sed", "-i.back", "s/PATH/%s/g" % escapedPath, plistFile])
    shutil.copy(plistFile, HELPER_PLIST)

def launchHelper():
    out = subprocess.call(["launchctl", "load", HELPER_PLIST])
    return out == 0

def grantPermissionsOnLogFolder():
    helperDir = os.path.join(_appdir, 'helper')
    try:
        os.makedirs(helperDir)
    except Exception:
        pass
    os.chown(helperDir, 0, 0)

def _getProcessList():
    _out = []
    output = subprocess.Popen(["ps", "-ceA"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    stdout, stderr = output.communicate()
    for line in stdout.decode('utf-8').split('\n'):
        cmd = line.split(' ')[-1]
        _out.append(cmd.strip())
    return _out

if __name__ == "__main__":
    main()
