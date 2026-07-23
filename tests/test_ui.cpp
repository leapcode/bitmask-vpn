// test_ui.cpp
#include <QtQuickTest/quicktest.h>
#include <QQmlEngine>
#include <QQmlContext>

#include "../gui/qjsonmodel.h"
#include "../lib/libgoshim.h"


QString getAppName(QJsonValue info, QString provider) {
    for (auto p: info.toArray()) {
        QJsonObject item = p.toObject();
        if (item["provider"] == provider) {
            return item["applicationName"].toString();
        }
    }
    return "BitmaskVPN";
}

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

        /* load providers json */
        QFile providerJson (":/providers.json");
        providerJson.open(QIODevice::ReadOnly | QIODevice::Text);
        QJsonModel *providers = new QJsonModel;
        QByteArray providerJsonBytes = providerJson.readAll();
        providers->loadJson(providerJsonBytes);
        QJsonValue defaultProvider = providers->json().object().value("default");
        QJsonValue providersInfo = providers->json().object().value("providers");
        QString appName = getAppName(providersInfo, defaultProvider.toString());

        QByteArray defaultProviderBytes = defaultProvider.toString().toUtf8();
        InitializeTestBitmaskContext(
            GoString{defaultProviderBytes.constData(),
                     (long int)defaultProviderBytes.length()},
            (char*)providerJsonBytes.data(), providerJsonBytes.length());

        ctx->setContextProperty("jsonModel", model);
        ctx->setContextProperty("providers", providers);

        /* helper for tests */
        ctx->setContextProperty("helper", helper);
    }
};

QUICK_TEST_MAIN_WITH_SETUP(ui, Setup)

#include "test_ui.moc"
