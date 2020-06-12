#include <QApplication>
#include <QSystemTrayIcon>
#include <QTimer>
#include <QtQml>
#include <QQmlApplicationEngine>
#include <QQuickWindow>
#include <string>

#include "handlers.h"
#include "qjsonmodel.h"
#include "lib/libgoshim.h"

/* Hi! I'm Troy McClure and I'll be your guide today. You probably remember me
   from blockbusters like "here be dragons" and "darling, I wrote a little
   contraption". */

QJsonWatch *qw = new QJsonWatch;

/* onStatusChanged is the C function that we register as a callback with CGO.
   It pulls a string serialization of the context object, than we then pass
   along to Qml via signals. */

void onStatusChanged() {
    char *ctx = RefreshContext();
    emit qw->jsonChanged(QString(ctx));
    free(ctx);
}

std::string getEnv(std::string const& key)
{
    char const* val = getenv(key.c_str());
    return val == NULL ? std::string() : std::string(val);
}

int main(int argc, char **argv) {

    bool debugQml = getEnv("DEBUG_QML_DATA") == "yes";

    if (argc > 1 && strcmp(argv[1], "install-helpers") == 0) {
        qDebug() << "Will try to install helpers with sudo";
        InstallHelpers();
        exit(0);
    }

    QApplication::setAttribute(Qt::AA_EnableHighDpiScaling);
    QApplication app(argc, argv);

    if (!QSystemTrayIcon::isSystemTrayAvailable()) {
        qDebug() << "No systray icon available. Things might not work for now, sorry...";
    }
    
    app.setQuitOnLastWindowClosed(false);
    QQmlApplicationEngine engine;
    QQmlContext *ctx = engine.rootContext();

    QJsonModel *model = new QJsonModel;
    std::string json = R"({"appName": "unknown", "provider": "unknown"})";
    model->loadJson(QByteArray::fromStdString(json));

    /* the backend handler has slots for calling back to Go when triggered by
       signals in Qml. */
    Backend backend;
    ctx->setContextProperty("backend", &backend);

    /* set the json model, load the qml */
    ctx->setContextProperty("jsonModel", model);
    ctx->setContextProperty("debugQml", debugQml);

    engine.load(QUrl(QStringLiteral("qrc:/qml/main.qml")));

    /* connect the jsonChanged signal explicitely.
        In the lambda, we reload the json in the model every time we receive an
        update from Go */
    QObject::connect(qw, &QJsonWatch::jsonChanged, [model](QString js) {
        model->loadJson(js.toUtf8());
    });

    /* connect quitDone signal, exit app */
    QObject::connect(&backend, &Backend::quitDone, []() {
            QGuiApplication::quit();
    });

    /* register statusChanged callback with CGO */
    const char *stCh = "OnStatusChanged";
    GoString statusChangedEvt = {stCh, (long int)strlen(stCh)};
    SubscribeToEvent(statusChangedEvt, (void *)onStatusChanged);

    /* let the Go side initialize its internal state */
    InitializeBitmaskContext();

    /* kick off your shoes, put your feet up */
    return app.exec();
}
