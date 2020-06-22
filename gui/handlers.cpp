#include <QTimer>
#include <QDebug>
#include <QDesktopServices>
#include <QUrl>

#include "handlers.h"
#include "lib/libgoshim.h"

Backend::Backend(QObject *parent) : QObject(parent)
{
}

void Backend::switchOn()
{
    SwitchOn();
}

void Backend::switchOff()
{
    SwitchOff();
}

void Backend::unblock()
{
    Unblock();
}

void Backend::donateAccepted()
{
    DonateAccepted();
}

void Backend::quit()
{
    Quit();
    emit this->quitDone();
}

