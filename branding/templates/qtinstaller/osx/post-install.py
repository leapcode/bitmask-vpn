#!/usr/bin/env python

import os
import shutil
import sys
import subprocess

HELPER = "bitmask-helper"
HELPER_PLIST = "/Library/LaunchDaemons/se.leap.bitmask-helper.plist"
_dir = os.path.dirname(os.path.realpath(__file__))

def main():
    log = open(os.path.join(_dir, 'post-install.log'), 'w')
    log.write('Checking for admin privileges...\n')

    _id = os.getuid()
    if _id != 0:
      err  = "error: need to run as root. UID: %s\n" % str(_id)
      logErr(log, err)
    
    # failure: sys.exit(1)
    
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


def logErr(log, msg):
    log.write(msg)
    sys.exit(1)

def isHelperRunning():
    ps = _getProcessList()
    return HELPER in ps 

def unloadHelper():
    out = subprocess.call(["launchctl", "unload", HELPER_PLIST])
    out2 = subprocess.call(["pkill", "-9", "bitmask-helper"])  # just in case
    return out == 0

def fixHelperOwner(log):
    path = os.path.join(_dir, HELPER)
    try:
        os.chown(path, 0, 0)
    except OSError as exc:
        log.write(str(exc))
        return False
    return True

def copyLaunchDaemon():
    plist = "se.leap.bitmask-helper.plist"
    path = os.path.join(_dir, plist)
    dest = os.path.join('/Library/LaunchDaemons', plist)
    _p = _dir.replace("/", "\/")
    subprocess.call(["sed", "-i.back", "s/PATH/%s/" % _p, path])
    shutil.copy(path, dest)

def launchHelper():
    out = subprocess.call(["launchctl", "load", "/Library/LaunchDaemons/se.leap.bitmask-helper.plist"])
    return out == 0

def grantPermissionsOnLogFolder():
    helperDir = os.path.join(_dir, 'helper')
    try:
        os.makedirs(helperDir)
    except Exception:
        pass
    os.chown(helperDir, 0, 0)

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
