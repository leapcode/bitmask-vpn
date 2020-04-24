import datetime
import os


def getDefaultProvider(config):
    provider = os.environ.get('PROVIDER')
    if provider:
        print('[+] Got provider {} from environment'.format(provider))
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
            'askForDonations', 'donateURL', 'apiURL',
            'geolocationAPI', 'caCertString')

    for value in keys:
        d[value] = c.get(value)

    d['timeStamp'] = '{:%Y-%m-%d %H:%M:%S}'.format(
        datetime.datetime.now())

    return d

