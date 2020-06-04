#include <QTimer>
#include <QDebug>

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

void Backend::quit()
{
    Quit();
    emit this->quitDone();
}
