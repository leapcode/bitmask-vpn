#ifndef HANDLERS_H
#define HANDLERS_H

#include <QDebug>
#include <QObject>
#include "qjsonmodel.h"
#include "lib/libgoshim.h"
#include <QJSValue>

#if defined(_WIN32) || defined(WIN32) || defined(__CYGWIN) || defined(__MINGW32__)
#define OS_WIN
#endif

GoString toGoStr(QString s);

class QJsonWatch : public QObject {

    Q_OBJECT

public:

signals:

    void jsonChanged(QString json);

};

class Backend : public QObject {

    Q_OBJECT

public:
    explicit Backend(QObject *parent = 0);

signals:
    void quitDone();
    void localeChanged(QString locale);

public slots:
    QString getVersion();
    void switchOn();
    void switchOff();
    void donateAccepted();
    void donateSeen();
    void useLocation(QString username);
    void useAutomaticGateway();
    void setTransport(QString transport);
    void setUDP(bool udp);
    void setSnowflake(bool snowflake);
    QString getTransport();
    void login(QString username, QString password);
    void resetError(QString errlabel);
    void resetNotification(QString label);
    void quit();
    void setLocale(QString locale);
    void switchProvider(QString provider, const QJSValue &callback);
    void switchProvider(QString provider);
};

#endif  // HANDLERS_H
