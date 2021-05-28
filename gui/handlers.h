#ifndef HANDLERS_H
#define HANDLERS_H

#include <QDebug>
#include <QObject>
#include "qjsonmodel.h"
#include "lib/libgoshim.h"

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

public slots:
    QString getVersion();
    void switchOn();
    void switchOff();
    void donateAccepted();
    void donateSeen();
    void useLocation(QString username);
    void useAutomaticGateway();
    void useTransport(QString transport);
    QString getTransport();
    void login(QString username, QString password);
    void resetError(QString errlabel);
    void resetNotification(QString label);
    void quit();
};

#endif  // HANDLERS_H
