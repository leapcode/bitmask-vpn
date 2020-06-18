// test_ui.cpp
#include <QtQuickTest>
#include <QQmlEngine>
#include <QQmlContext>

#include "../gui/qjsonmodel.h"
#include "../lib/libgoshim.h"

class Helper :  public QObject
{
    Q_OBJECT

public:
    explicit Helper(QObject *parent = 0);

public slots:
    Q_INVOKABLE QString refreshContext();
};

Helper::Helper(QObject *parent) : QObject(parent)
{
}

Q_INVOKABLE QString Helper::refreshContext()
{
    return QString(RefreshContext());
}

class Setup : public QObject
{
    Q_OBJECT

public:
    Setup() {}

public slots:
    void qmlEngineAvailable(QQmlEngine *engine)
    {
        QQmlContext *ctx = engine->rootContext();
        QJsonModel *model = new QJsonModel;
        Helper *helper = new Helper(this);

        InitializeTestBitmaskContext();

        ctx->setContextProperty("jsonModel", model);
        ctx->setContextProperty("helper", helper);
    }
};

QUICK_TEST_MAIN_WITH_SETUP(ui, Setup)

#include "test_ui.moc"
