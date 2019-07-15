#!/usr/bin/env python3

# Generates a simplified file with variables that
# can be imported from the main vendorized Makefile.

import os
import sys

import configparser

from provider import getDefaultProvider
from provider import getProviderData


VERSION = os.environ.get('VERSION', 'unknown')

TEMPLATE = """
# Variables for the build of {applicationName}.
# Generated automatically. Do not edit.
APPNAME := {applicationName}
BINNAME := {binaryName}
VERSION := {version}
"""


def writeOutput(data, outfile):

    configString = TEMPLATE.format(
        binaryName=data['binaryName'],
        applicationName=data['applicationName'],
        version=data['version'],
    )

    with open(outfile, 'w') as outf:
        outf.write(configString)


if __name__ == "__main__":
    env_provider_conf = os.environ.get('PROVIDER_CONFIG')
    if env_provider_conf:
        if os.path.isfile(env_provider_conf):
            print("[+] Overriding provider config per "
                  "PROVIDER_CONFIG variable")
            configfile = env_provider_conf

    config = configparser.ConfigParser()
    config.read(configfile)
    provider = getDefaultProvider(config)
    data = getProviderData(provider, config)

    if len(sys.argv) != 2:
        print('Usage: generate-vendor-make.py <output_file>')
        sys.exit(1)

    outputf = sys.argv[1]
    data['version'] = VERSION

    writeOutput(data, outputf)
