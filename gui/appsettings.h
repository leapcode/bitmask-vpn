#ifndef APPSETTINGS_H
#define APPSETTINGS_H

#include <QApplication>
#include <QSettings>

class appSettings : public QSettings {
  Q_OBJECT

public:
  explicit appSettings(QObject *parent = 0)
      : QSettings(QSettings::UserScope,
                  QApplication::instance()->organizationName(),
                  QApplication::instance()->applicationName(), parent) {}
  Q_INVOKABLE inline void setValue(const QString &key, const QVariant &value) {
    QSettings::setValue(key, value);
  }
  Q_INVOKABLE inline QVariant
  value(const QString &key, const QVariant &defaultValue = QVariant()) const {
    return QSettings::value(key, defaultValue);
  }
};
Q_DECLARE_METATYPE(appSettings *)

#endif
