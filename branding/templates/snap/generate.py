#!/usr/bin/env python3
"""
generate.py

Generate a snap package for a given provider.
"""

import json
import os
from string import Template


TEMPLATE = 'snapcraft-template.yaml'
SNAPCRAFT = 'snapcraft.yaml'


here = os.path.split(os.path.realpath(__file__))[0]
data = json.load(open(os.path.join(here, 'data.json')))


template = Template(open(TEMPLATE).read())

with open(SNAPCRAFT, 'w') as output:
    output.write(template.safe_substitute(data))

print("[+] Snapcraft spec written to {path}".format(
    path=os.path.abspath(SNAPCRAFT)))
