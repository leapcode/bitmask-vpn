Howto i18n
----------

The translations are done in transifex. To help us contribute your translations there and/or review the existing
ones:
https://www.transifex.com/otf/bitmask/bitmask-desktop/

When a string has being modified you need to regenerate the locales:
```
  make generate_locales
```


To fetch the translations from transifex (API\_TOKEN is the transifex API token):
```
  API_TOKEN='xxxxxxxxxxx' make locales
```

If you want to add a new language create an empty file `gui/i18n/main_$lang.ts` before running `make locales`.

Testing the translations
------------------------

Pass the language env vars:

LANG=es_ES LANGUAGE=es_ES make run
