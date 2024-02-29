#!/usr/bin/env python3
"""
generate.py

Generate a snap package for a given provider.
"""

import json
import os
from string import Template


TEMPLATES = ('app.install', 'app.desktop', 'changelog', 'control', 'rules', 'source/include-binaries')


here = os.path.split(os.path.realpath(__file__))[0]
data = json.load(open(os.path.join(here, 'data.json')))


def write_from_template(target):
    template = Template(open(target + '-template').read())

    with open(target, 'w') as output:
        output.write(template.safe_substitute(data))


for target in TEMPLATES:
    write_from_template(target)


print("[+] Debian files written to {path}".format(
    path=os.path.abspath(here)))
