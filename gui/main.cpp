#include <csignal>
#include <QApplication>
#include <QTimer>
#include <QTranslator>
#include <QCommandLineParser>
#include <QQuickWindow>
#include <QSystemTrayIcon>
#include <QtQml>

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

void signalHandler(int) {
    Quit();
    exit(0);
}

int main(int argc, char **argv) {
    signal(SIGINT, signalHandler);
    bool debugQml = getEnv("DEBUG_QML_DATA") == "yes";

    Backend backend;

    QApplication::setAttribute(Qt::AA_EnableHighDpiScaling);
    QApplication::setApplicationName(backend.getAppName());
    QApplication::setApplicationVersion(backend.getVersion());
    QApplication app(argc, argv);


    QCommandLineParser parser;
    parser.setApplicationDescription(backend.getAppName() + ": a fast and secure VPN. Powered by Bitmask.");
    parser.addHelpOption();
    parser.addVersionOption();
    parser.process(app);

    const QStringList args = parser.positionalArguments();

    if (args.at(0) == "install-helpers") {
        qDebug() << "Will try to install helpers with sudo";
        InstallHelpers();
        exit(0);
    }

    if (!QSystemTrayIcon::isSystemTrayAvailable()) {
        qDebug() << "No systray icon available. Things might not work for now, sorry...";
    }

    QTranslator translator;
    translator.load(QLocale(), QLatin1String("main"), QLatin1String("_"), QLatin1String(":/i18n"));
    app.installTranslator(&translator);
    
    app.setQuitOnLastWindowClosed(false);
    QQmlApplicationEngine engine;
    QQmlContext *ctx = engine.rootContext();

    QJsonModel *model = new QJsonModel;
    std::string json = R"({"appName": "unknown", "provider": "unknown"})";
    model->loadJson(QByteArray::fromStdString(json));

    /* the backend handler has slots for calling back to Go when triggered by
       signals in Qml. */
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
