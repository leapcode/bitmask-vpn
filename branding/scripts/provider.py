import datetime
import os


def getDefaultProviders(config):
    # returns a list of providers
    provider = os.environ.get('PROVIDER')
    if provider:
        print('[+] Got provider {} from environment'.format(provider))
    else:
        print('[+] Using default provider from config file')
        provider = config['default']['provider']
    return provider.split(',')


def getProviderData(provider, config):
    print("[+] Configured provider:", provider)
    try:
        c = config[provider]
    except Exception:
        raise ValueError('Cannot find provider')

    d = dict()
    keys = ('name', 'applicationName', 'binaryName', 'auth', 'authEmptyPass',
            'providerURL', 'tosURL', 'helpURL',
            'askForDonations', 'donateURL', 'apiURL',
            'apiVersion', 'geolocationAPI', 'caCertString')
    boolValues = ['askForDonations', 'authEmptyPass']
    intValues = ['apiVersion', ]

    for value in keys:
        if value not in c:
            continue
        d[value] = c.get(value)
        if value in boolValues:
            d[value] = bool(d[value])
        elif value in intValues:
            d[value] = int(d[value])

    d['timeStamp'] = '{:%Y-%m-%d %H:%M:%S}'.format(
        datetime.datetime.now())

    return d
