Howto i18n
----------

The translations are done in transifex. To help us contribute your translations there and/or review the existing
ones:
https://www.transifex.com/otf/bitmask/bitmask-desktop/

When a string has being modified you need to regenerate the locales:
```
  make generate_locales
```


To fetch the translations from transifex you need to use the Transifex cli:
https://developers.transifex.com/docs/cli and an api (API\_TOKEN is the transifex API
token)
```
  API_TOKEN='xxxxxxxxxxx' tx pull
```

If you want to add a new language you can:
```
  API_TOKEN='xxxxxxxxxxx' tx pull -a
```

Sometimes language codes are not what you expect. This applies for missing languages as
well. When you check in transifex, you can also see what is used there, for example fa_IR
or es_AR, es or es_CU. When you want to use some language in general instead of some
regional version you can use the mapping in the .tx/config. Examples: fa_IR maps to fa. 

For this project we expect files to be like main_es_AR.ts or main_pl.ts See 
https://doc.qt.io/QtForMCUs-2.5/qtul-cmake-getting-started.html

Testing the translations
------------------------

Pass the language env vars:

LANG=es_ES LANGUAGE=es_ES make run
