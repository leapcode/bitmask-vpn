#!/usr/bin/env python3
"""
generate.py

Generate a NSI installer for a given provider.
"""

import json
import os
from string import Template


TEMPLATE = 'template.nsi'


def get_files(which):
    files = "\n"
    if which == 'install':
        action = "File "
    elif which == 'uninstall':
        action = "Delete $INSTDIR\\"
    else:
        action = ""

    # TODO get relative path
    for item in open('payload/' + which).readlines():
        files += "  {action}{item}".format(
            action=action, item=item)
    return files


here = os.path.split(os.path.realpath(__file__))[0]
data = json.load(open(os.path.join(here, 'data.json')))
data['extra_install_files'] = get_files('install')
data['extra_uninstall_files'] = get_files('uninstall')

INSTALLER = data['applicationName'] + '-installer.nsi'


template = Template(open(TEMPLATE).read())
with open(INSTALLER, 'w') as output:
    output.write(template.safe_substitute(data))

print("[+] NSIS installer script written to {path}".format(path=INSTALLER))
