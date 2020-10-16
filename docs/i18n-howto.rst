Howto i18n
----------

The translations are done in transifex. To help us contribute your translations there and/or review the existing
ones:
https://www.transifex.com/otf/bitmask/RiseupVPN/

When a string has being modified you need to regenerate the locales:
```
  make generate_locales
```


To fetch the translations from transifex and rebuild the catalog.go (API\_TOKEN is the transifex API token):
```
  API_TOKEN='xxxxxxxxxxx' make locales
```
There is some bug on gotext and the catalog.go generated doesn't have a package, you will need to edit
cmd/bitmask-vpn/catalog.go and to have a `package main` at the beginning of the file.

If you want to add a new language create the folder `locales/$lang` before running `make locales`.
