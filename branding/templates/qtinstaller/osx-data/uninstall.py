#!/usr/bin/env python

# Uninstall script for BitmaskVPN.

import os
import shutil
import sys
import subprocess

HELPER = "bitmask-helper"
HELPER_PLIST = "/Library/LaunchDaemons/se.leap.bitmask-helper.plist"

_dir = os.path.dirname(os.path.realpath(__file__))

def main(stage="uninstall"):
    logfile = "bitmask-{stage}.log".format(stage=stage)
    log = open(os.path.join('/tmp', logfile), 'w')
    log.write('Checking for admin privileges...\n')

    _id = os.getuid()
    log.write("UID: %s\n" % str(_id))
    if int(_id) != 0:
      err  = "error: need to run as root. UID: %s\n" % str(_id)
      logErr(log, err)
    
    # failure: sys.exit(1)

    log.write('Checking if helper is running\n')
    
    if isHelperRunning():
        log.write("Trying to stop bitmask-helper...\n")
	# if this fail, we can check if the HELPER_PLIST is there
        ok = unloadHelper()
        log.write("success: %s \n" % str(ok))

    log.write("Removing LaunchDaemon\n")
    out = removeLaunchDaemon()
    log.write("result: %s \n" % str(out))
    
    # all done
    log.write(stage + ' script: done\n')
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

def removeLaunchDaemon():
    return subprocess.call(["rm", "-f", HELPER_PLIST])

def _getProcessList():
    _out = []
    output = subprocess.Popen(["ps", "-ceA"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    stdout, stderr = output.communicate()
    for line  in stdout.split('\n'):
        cmd = line.split(' ')[-1]
        _out.append(cmd.strip())
    return _out

if __name__ == "__main__":
    stage="uninstall"
    if len(sys.argv) > 2 and sys.argv[2] == "pre":
        stage="pre-install"
    main(stage)
