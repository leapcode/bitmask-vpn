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
            'geolocationAPI', 'caCertString')
    boolValues = ['askForDonations', 'authEmptyPass']

    for value in keys:
        d[value] = c.get(value)
        if value in boolValues:
            d[value] = bool(d[value])

    d['timeStamp'] = '{:%Y-%m-%d %H:%M:%S}'.format(
        datetime.datetime.now())

    return d
