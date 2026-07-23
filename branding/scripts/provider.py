import datetime
import os


def getDefaultProviders(config):
    # returns a list of providers
    provider = os.environ.get('PROVIDERS') or os.environ.get('PROVIDER')
    if provider:
        print('[+] Got providers {} from environment'.format(provider))
    else:
        print('[+] Using default providers from config file')
        provider = config['default']['provider']
    return [p for p in provider.split(',') if p]


def getDefaultProvider(config):
    # returns the default provider
    env = os.environ.get('DEFAULT_PROVIDER')
    if env is not None:
        return env.strip()
    if config.has_option('default', 'default'):
        return config['default'].get('default', '').strip()
    return getDefaultProviders(config)[0]


def getBrandingProvider(config):
    # returns the provider used for branding (binary name, app name, assets).
    # This is the canonical provider identifier (the INI section header),
    # not the display name.
    env = os.environ.get('PROVIDER')
    if env:
        return env.strip()
    return config['default']['provider'].strip()


def getProviderData(provider, config):
    print("[+] Configured provider:", provider)
    try:
        c = config[provider]
    except Exception:
        raise ValueError('Cannot find provider')

    d = dict()
    # 'provider' is the canonical identifier (the INI section header, e.g.
    # 'coopvpn'). 'name' is the display name shown in the UI (e.g. 'CoopVPN').
    d['provider'] = provider
    keys = ('name', 'applicationName', 'binaryName', 'auth', 'authEmptyPass',
            'providerURL', 'tosURL', 'helpURL',
            'askForDonations', 'donateURL', 'apiURL',
            'apiVersion', 'geolocationAPI', 'caCertString',
            'STUNServers', 'countryCodeLookupURL')
    boolValues = ['askForDonations', 'authEmptyPass']
    intValues = ['apiVersion', ]
    listValues = ['STUNServers']

    for value in keys:
        if value not in c:
            continue
        d[value] = c.get(value)
        if value in boolValues:
            d[value] = bool(d[value])
        elif value in intValues:
            d[value] = int(d[value])
        elif value in listValues:
            if d[value].strip() == "":
                d[value] = []
            else:
                d[value] = d[value].split(",")
                # remove spaces
                d[value] = [x.strip() for x in d[value]]

    d['timeStamp'] = '{:%Y-%m-%d %H:%M:%S}'.format(
        datetime.datetime.now())

    return d
