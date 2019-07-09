#!/usr/bin/env python3

import datetime
import os
import sys

from string import Template
import configparser

OUTFILE = 'config.go'
INFILE = 'config.go.tmpl'
CONFIGFILE = 'config/vendor.conf'
SCRIPT_NAME = 'vendorize'


def getDefaultProvider(config):
    provider = os.environ.get('PROVIDER')
    if provider:
        print('[+] Got provider {} from environemnt'.format(provider))
    else:
        print('[+] Using default provider from config file')
        provider = config['default']['provider']
    return provider


def getProviderData(provider, config):
    print("[+] Configured provider:", provider)

    c = config[provider]
    d = dict()

    keys = ('name', 'applicationName', 'binaryName',
            'providerURL', 'tosURL', 'helpURL',
            'donateURL', 'apiURL', 'geolocationAPI', 'caCertString')

    for value in keys:
        d[value] = c.get(value)

    d['timeStamp'] = '{:%Y-%m-%d %H:%M:%S}'.format(
        datetime.datetime.now())

    return d


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
        print('Usage: {scriptname}.py <template> <config> <output>'.format(
            scriptname=SCRIPT_NAME))
    else:
        print(msg)
    sys.exit(1)


if __name__ == "__main__":
    infile = outfile = ""

    if len(sys.argv) > 4:
        bail()

    elif len(sys.argv) == 1:
        infile = INFILE
        outfile = OUTFILE
        configfile = CONFIGFILE
    else:
        try:
            infile = sys.argv[1]
            configfile = sys.argv[2]
            outfile = sys.argv[3]
        except IndexError:
            bail()

    env_provider_conf = os.environ.get('PROVIDER_CONFIG')
    if env_provider_conf:
        if os.path.isfile(env_provider_conf):
            print("[+] Overriding provider config per PROVIDER_CONFIG variable")
            configfile = env_provider_conf

    if not os.path.isfile(infile):
        bail('[!] Cannot find template in {path}'.format(
            path=os.path.abspath(infile)))
    elif not os.path.isfile(configfile):
        bail('[!] Cannot find config in {path}'.format(
            path=os.path.abspath(configfile)))
    else:
        print('[+] Using {path} as template'.format(
            path=os.path.abspath(infile)))
        print('[+] Using {path} as config'.format(
            path=os.path.abspath(configfile)))

    config = configparser.ConfigParser()
    config.read(configfile)

    provider = getDefaultProvider(config)
    data = getProviderData(provider, config)
    addCaData(data, configfile)
    writeOutput(data, infile, outfile)

    print('[+] Wrote configuration for {provider} to {outf}'.format(
        provider=data.get('name'),
        outf=os.path.abspath(outfile)))
