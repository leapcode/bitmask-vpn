#!/usr/bin/env python3
import configparser
import json
import os
import sys


from provider import getDefaultProvider
from provider import getProviderData

OUTFILE = 'providers.json'
SCRIPT_NAME = 'gen-providers-json'


def generateProvidersJSON(configPath, outputJSONPath):
    print("output:", outputJSONPath)
    config = configparser.ConfigParser()
    config.read(configPath)

    # TODO as a first step, we just get the defaultProvider.
    # For multi-provider, just add more providers to the dict

    providers = {}
    defaultProvider = getDefaultProvider(config)
    providers['default'] = defaultProvider
    providerData = getProviderData(defaultProvider, config)
    addCaData(providerData, configPath)

    providers[defaultProvider] = providerData
    with open(outputJSONPath, 'w', encoding='utf-8') as f:
        json.dump(providers, f, ensure_ascii=False, indent=4)

def addCaData(data, configfile):
    provider = data.get('name').lower()
    folder, f = os.path.split(configfile)
    caFile = os.path.join(folder, provider + '-ca.crt')
    if not os.path.isfile(caFile):
        bail('[!] Cannot find CA file in {path}'.format(path=caFile))
    with open(caFile) as ca:
        data['caCertString'] = ca.read().strip()

def writeOutput(data, infile, outfile):

    with open(infile) as infile:
        s = Template(infile.read())

    with open(outfile, 'w') as outf:
        outf.write(s.substitute(data))

def bail(msg=None):
    if not msg:
        print("ERROR: not enough arguments!")
        print('Usage: {scriptname}.py <config> <output>'.format(
            scriptname=SCRIPT_NAME))
    else:
        print(msg)
    sys.exit(1)

if __name__ == "__main__":
    print("[+] Generating providers.json...")
    if len(sys.argv) != 3:
        bail()
    # TODO get BITMASK_BRANDING folder - get config from there, if possible.
    env_provider_conf = os.environ.get('PROVIDER_CONFIG')
    if env_provider_conf:
        if os.path.isfile(env_provider_conf):
            print("[+] Overriding provider config per "
                  "PROVIDER_CONFIG variable")
            configfile = env_provider_conf
    generateProvidersJSON(sys.argv[1], sys.argv[2])
