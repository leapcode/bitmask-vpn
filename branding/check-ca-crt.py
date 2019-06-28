#!/usr/bin/env python3
import re
import sys
import urllib.request

SCRIPT_NAME = 'check-ca-crt.py'

USAGE = '''Check that the stored provider CA matches the one announced online.
Usage: {name} <provider> <uri>

Example: {name} riseup black.riseup.net'''.format(name=SCRIPT_NAME)


def getLocalCert(provider):
    sanitized = re.sub(r'[^\w\s-]', '', provider).strip().lower()
    with open('config/{provider}-ca.crt'.format(provider=sanitized)) as crt:
        return crt.read().strip()


def getRemoteCert(uri):
    fp = urllib.request.urlopen('https://' + uri + '/ca.crt')
    remote_cert = fp.read().decode('utf-8').strip()
    fp.close()
    return remote_cert


if __name__ == '__main__':

    if len(sys.argv) != 3:
        print('[!] Not enough arguments')
        print(USAGE)
        sys.exit(1)

    provider = sys.argv[1]
    uri = sys.argv[2]

    local = getLocalCert(provider)
    remote = getRemoteCert(uri)

    try:
        assert local == remote
    except AssertionError:
        print('[!] ERROR: remote and local CA certs do not match')
        sys.exit(1)
    else:
        print('OK')
