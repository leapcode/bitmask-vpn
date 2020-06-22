#ifndef HANDLERS_H
#define HANDLERS_H

#include <QDebug>
#include <QObject>
#include "qjsonmodel.h"

class QJsonWatch : public QObject {

    Q_OBJECT

    QJsonModel *model;

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
    void switchOn();
    void switchOff();
    void unblock();
    void donateAccepted();
    void quit();
};

#endif  // HANDLERS_H
