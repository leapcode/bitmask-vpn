#!/usr/bin/python

# Generate bundles for brandable Bitmask Lite.

# (c) LEAP Encryption Access Project
# (c) Kali Kaneko 2018-2019

import json
import os
import os.path
import shutil
import stat

from string import Template

here = os.path.split(os.path.abspath(__file__))[0]

ENTRYPOINT = 'bitmask-vpn'
HELPER = 'bitmask-helper'
OPENVPN = 'openvpn-osx'
TEMPLATE_INFO = 'template-info.plist'
TEMPLATE_HELPER = 'template-helper.plist'

TEMPLATE_PREINSTALL = 'template-preinstall'
TEMPLATE_POSTINSTALL = 'template-postinstall'


data = json.load(open(os.path.join(here, 'data.json')))
APPNAME = data.get('applicationName')
VERSION = data.get('version', 'unknown')

APP_PATH = os.path.abspath(here + '/../dist/' + APPNAME + ".app")
STAGING = os.path.abspath(here + '/../staging/')
ASSETS = os.path.abspath(here + '/../assets/')
ICON = os.path.join(ASSETS, 'icon.icns')
SCRIPTS = os.path.join(os.path.abspath(here), 'scripts')
INFO_PLIST = APP_PATH + '/Contents/Info.plist'
HELPER_PLIST = os.path.join(SCRIPTS, 'se.leap.bitmask-helper.plist')
PREINSTALL = os.path.join(SCRIPTS, 'preinstall')
POSTINSTALL = os.path.join(SCRIPTS, 'postinstall')
RULEFILE = os.path.join(here, 'bitmask.pf.conf')
VPN_UP = os.path.join(here, 'client.up.sh')
VPN_DOWN = os.path.join(here, 'client.down.sh')

try:
    os.makedirs(APP_PATH + "/Contents/MacOS")
except Exception:
    pass
try:
    os.makedirs(APP_PATH + "/Contents/Resources")
except Exception:
    pass
try:
    os.makedirs(APP_PATH + "/Contents/helper")
except Exception:
    pass


data['entrypoint'] = ENTRYPOINT
data['info_string'] = APPNAME + " " + VERSION
data['bundle_identifier'] = 'se.leap.' + data['applicationNameLower']
data['bundle_name'] = APPNAME

# utils


def copy_payload(filename, destfile=None):
    if destfile is None:
        destfile = APP_PATH + "/Contents/MacOS/" + filename
    else:
        destfile = APP_PATH + destfile
    shutil.copyfile(STAGING + '/' + filename, destfile)
    cmode = os.stat(destfile).st_mode
    os.chmod(destfile, cmode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH)


def generate_from_template(template, dest, data):
    print("[+] File written from template to", dest)
    template = Template(open(template).read())
    with open(dest, 'w') as output:
        output.write(template.substitute(data))


# 1. Generation of the Bundle Info.plist
# --------------------------------------

generate_from_template(TEMPLATE_INFO, INFO_PLIST, data)


# 2. Generate PkgInfo
# -------------------------------------------

with open(APP_PATH + "/Contents/PkgInfo", "w") as f:
    # is this enough? See what PyInstaller does.
    f.write("APPL????")


# 3. Copy the binary payloads
# --------------------------------------------

copy_payload(ENTRYPOINT)
copy_payload(HELPER)
copy_payload(OPENVPN, destfile='/Contents/Resources/openvpn.leap')

# 4. Copy the app icon from the assets folder
# -----------------------------------------------

shutil.copyfile(ICON, APP_PATH + '/Contents/Resources/app.icns')


# 5. Generate the scripts for the installer
# -----------------------------------------------

# Watch out that, for now, all the brandings are sharing the same helper name.
# This is intentional: I prefer not to have too many root helpers laying around
# until we consolidate a way of uninstalling and/or updating them.
# This also means that only one of the derivatives will work at a given time
# (ie, uninstall bitmask legacy to use riseupvpn).
# If this bothers you, and it should, let's work on improving uninstall and
# updates.

generate_from_template(TEMPLATE_HELPER, HELPER_PLIST, data)
generate_from_template(TEMPLATE_PREINSTALL, PREINSTALL, data)
generate_from_template(TEMPLATE_POSTINSTALL, POSTINSTALL, data)

# 6. Copy helper pf rule file
# ------------------------------------------------

shutil.copy(RULEFILE, APP_PATH + '/Contents/helper/')

# 7. Copy openvpn up/down scripts
# ------------------------------------------------

shutil.copy(VPN_UP,   APP_PATH + '/Contents/helper/')
shutil.copy(VPN_DOWN, APP_PATH + '/Contents/helper/')


# 8. Generate uninstall script
# -----------------------------------------------
# TODO copy the uninstaller script from bitmask-dev
# TODO substitute vars
# this is a bit weak procedure for now.
# To begin with, this assumes everything is hardcoded into
# /Applications/APPNAME.app
# We could consider moving the helpers into /usr/local/sbin,
# so that the plist files always reference there.


# We're all set!
# -----------------------------------------------
print("[+] Output written to build/{provider}/dist/{appname}.app".format(
    provider=data['name'].lower(),
    appname=APPNAME))
