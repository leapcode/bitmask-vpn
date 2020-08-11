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
    app.setQuitOnLastWindowClosed(false);

    QCommandLineParser parser;
    parser.setApplicationDescription(
        backend.getAppName() +
        QApplication::translate(
            "main", ": a fast and secure VPN. Powered by Bitmask."));
    parser.addHelpOption();
    parser.addVersionOption();
    parser.addOptions({
        {
            {"n", "no-systray"},
            QApplication::translate("main",
                                    "Do not show the systray icon (useful "
                                    "together with gnome shell "
                                    "extension, or to control VPN by other means)."),
        },
        {
            {"w", "web-api"},
            QApplication::translate(
                "main",
                "Enable web api."),
        },
        {
            {"i", "install-helpers"},
            QApplication::translate(
                "main",
                "Install helpers (linux only, requires sudo)."),
        },
    });
    QCommandLineOption webPortOption("web-port", QApplication::translate("main", "Web api port (default: 8080)"), "port", "8080");
    parser.addOption(webPortOption);
    parser.process(app);

    bool hideSystray    = parser.isSet("no-systray");
    bool installHelpers = parser.isSet("install-helpers");
    bool webAPI         = parser.isSet("web-api");
    QString webPort     = parser.value("web-port");

    if (hideSystray) {
        qDebug() << "Not showing systray icon because --no-systray option is set.";
    }

    if (installHelpers) {
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
    
    QQmlApplicationEngine engine;
    QQmlContext *ctx = engine.rootContext();

    QJsonModel *model = new QJsonModel;

    /* the backend handler has slots for calling back to Go when triggered by
       signals in Qml. */
    ctx->setContextProperty("backend", &backend);

    /* we pass the json model and set some useful flags */
    ctx->setContextProperty("jsonModel", model);
    ctx->setContextProperty("debugQml", debugQml);
    ctx->setContextProperty("systrayVisible", !hideSystray);

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

    /* if requested, enable web api for controlling the VPN */
    if (webAPI) {
        char* wp = webPort.toLocal8Bit().data();
        GoString p = {wp, (long int)strlen(wp)};
        EnableWebAPI(p);
    };

    /* kick off your shoes, put your feet up */
    return app.exec();
}
