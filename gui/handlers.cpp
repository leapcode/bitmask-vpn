#include <QTimer>
#include <QDebug>
#include <QDesktopServices>
#include <QUrl>

#include "handlers.h"
#include "lib/libgoshim.h"

Backend::Backend(QObject *parent) : QObject(parent)
{
}

QString Backend::getAppName()
{
    return QString(GetAppName());
}

QString Backend::getVersion()
{
    return QString(GetVersion());
}

void Backend::switchOn()
{
    SwitchOn();
}

void Backend::switchOff()
{
    SwitchOff();
}

void Backend::donateAccepted()
{
    DonateAccepted();
}

void Backend::login(QString username, QString password)
{
    // TODO: there has to be a cleaner way to do the conversion
    char * u = new char [username.length()+1];
    char * p = new char [password.length()+1];
    strcpy(u, username.toStdString().c_str());
    strcpy(p, password.toStdString().c_str());
    Login(u, p);
    delete [] u;
    delete [] p;
}

void Backend::quit()
{
    Quit();
    emit this->quitDone();
}

