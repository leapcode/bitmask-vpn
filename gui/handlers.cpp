#include <QTimer>
#include <QDebug>
#include <QDesktopServices>
#include <QUrl>

#include "handlers.h"
#include "lib/libgoshim.h"

GoString toGoStr(QString s)
{
    const char *c = s.toUtf8().constData();
    return (GoString){c, (long int)strlen(c)};
}


Backend::Backend(QObject *parent) : QObject(parent)
{
}

QString Backend::getVersion()
{
    return QString(GetBitmaskVersion());
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

void Backend::donateSeen()
{
    DonateSeen();
}

void Backend::useLocation(QString label)
{
    UseLocation(toGoStr(label));
}

void Backend::useAutomaticGateway()
{
    UseAutomaticGateway();
}

void Backend::setTransport(QString transport)
{
    SetTransport(toGoStr(transport));
}

void Backend::setUDP(bool udp)
{
    SetUDP(udp);
}

void Backend::setSnowflake(bool snowflake)
{
    SetSnowflake(snowflake);
}

QString Backend::getTransport()
{
    return QString(GetTransport());
}

void Backend::login(QString username, QString password)
{
    Login(toGoStr(username), toGoStr(password));
}

void Backend::resetError(QString errlabel)
{
    ResetError(toGoStr(errlabel));
}

void Backend::resetNotification(QString label)
{
    ResetNotification(toGoStr(label));
}

void Backend::quit()
{
    Quit();
    emit this->quitDone();
}

