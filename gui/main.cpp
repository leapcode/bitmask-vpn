#include <csignal>
#include <unistd.h>
#include <QtGui/qfontdatabase.h>
#include <QApplication>
#include <QTimer>
#include <QTranslator>
#include <QCommandLineParser>
#include <QQuickWindow>
#include <QQuickStyle>
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

QString getAppName(QJsonValue info, QString provider) {
    for (auto p: info.toArray()) {
        QJsonObject item = p.toObject();
        if (item["name"].toString().toLower() == provider.toLower()) {
            return item["applicationName"].toString();
        }
    }
    return "BitmaskVPN";
}

int main(int argc, char **argv) {
    signal(SIGINT, signalHandler);

    Backend backend;

    QApplication::setAttribute(Qt::AA_EnableHighDpiScaling);
    QApplication::setApplicationVersion(backend.getVersion());
    QApplication app(argc, argv);
    app.setQuitOnLastWindowClosed(false);
    app.setAttribute(Qt::AA_UseHighDpiPixmaps);

    /* load providers json */
    QFile providerJson (":/providers.json");
    providerJson.open(QIODevice::ReadOnly | QIODevice::Text);
    QJsonModel *providers = new QJsonModel;
    QByteArray providerJsonBytes = providerJson.readAll();
    providers->loadJson(providerJsonBytes);
    QJsonValue defaultProvider = providers->json().object().value("default");
    QJsonValue providersInfo = providers->json().object().value("providers");
    QString appName = getAppName(providersInfo, defaultProvider.toString());

    QApplication::setApplicationName(appName);

    QCommandLineParser parser;
    parser.setApplicationDescription(
        appName +
        QApplication::translate(
            "main", ": a fast and secure VPN. Powered by Bitmask."));
    parser.addHelpOption();
    parser.addVersionOption();
    parser.addOptions({
        {
            {"n", "no-systray"},
            QApplication::translate("main",
                                    "Do not show the systray icon (useful "
                                    "together with Gnome Shell "
                                    "extension, or to control VPN by other means)."),
        },
        {
            {"w", "web-api"},
            QApplication::translate(
                "main",
                "Enable Web API."),
        },
        {
            {"i", "install-helpers"},
            QApplication::translate(
                "main",
                "Install helpers (Linux only, requires sudo)."),
        },
        {
            {"o", "obfs4"},
            QApplication::translate(
                "main",
                "Use obfs4 to obfuscate the traffic, if available in the provider."),
        },
        {
            {"a", "disable-autostart"},
            QApplication::translate(
                "main",
                "Disable autostart for the next run."),
        },
    });
    QCommandLineOption webPortOption("web-port", QApplication::translate("main", "Web API port (default: 8080)"), "port", "8080");
    parser.addOption(webPortOption);
    QCommandLineOption startVPNOption("start-vpn", QApplication::translate("main", "Start the VPN, either 'on' or 'off'."), "status", "");
    parser.addOption(startVPNOption);
    parser.process(app);

    bool hideSystray    = parser.isSet("no-systray");
    bool availableSystray = true;
    bool installHelpers = parser.isSet("install-helpers");
    bool webAPI         = parser.isSet("web-api");
    QString webPort     = parser.value("web-port");
    bool version        = parser.isSet("version");
    bool obfs4          = parser.isSet("obfs4");
    bool disAutostart   = parser.isSet("disable-autostart");
    QString startVPN    = parser.value("start-vpn");

    if (version) {
        qDebug() << backend.getVersion();
        exit(0);
    }

    if (startVPN != "" && startVPN != "on" && startVPN != "off") {
        qDebug() << "Error: --start-vpn must be either 'on' or 'off'";
        exit(0);
    }

    if (hideSystray)
        qDebug() << "Not showing systray icon because --no-systray option is set.";

    if (installHelpers) {
        qDebug() << "Will try to install helpers with sudo";
        InstallHelpers();
        exit(0);
    }

#ifdef Q_OS_UNIX
    if (getuid() == 0) {
        qDebug() << "Please don't run as root. Aborting.";
        exit(0);
    }
#endif

    if (!QSystemTrayIcon::isSystemTrayAvailable()) {
        qDebug() << "No systray icon available.";
        availableSystray = false;
    }

    /* set window icon */
    app.setWindowIcon(QIcon(":/vendor/icon.svg"));

    /* load translations */
    QTranslator translator;
    translator.load(QLocale(), QLatin1String("main"), QLatin1String("_"), QLatin1String(":/i18n"));
    app.installTranslator(&translator);


    QQmlApplicationEngine engine;
    QQmlContext *ctx = engine.rootContext();

    QJsonModel *model = new QJsonModel;

    // FIXME use qgetenv
    QString desktop = QString::fromStdString(getEnv("XDG_CURRENT_DESKTOP"));
    QString debug = QString::fromStdString(getEnv("DEBUG"));

    /* the backend handler has slots for calling back to Go when triggered by
       signals in Qml. */
    ctx->setContextProperty("backend", &backend);

    /* set the json model, load providers.json */
    ctx->setContextProperty("jsonModel", model);
    ctx->setContextProperty("providers", providers);
    ctx->setContextProperty("desktop", desktop);

    /* set some useful flags */
    ctx->setContextProperty("systrayVisible", !hideSystray);
    ctx->setContextProperty("systrayAvailable", availableSystray);
    ctx->setContextProperty("qmlDebug", debug == "1");

    //XXX we're doing configuration via config file, but this is a mechanism
    //to change to Dark Theme if desktop has it.
    //qputenv("QT_QUICK_CONTROLS_MATERIAL_VARIANT", "Dense");
    //QQuickStyle::setStyle("Material");
    engine.load(QUrl(QStringLiteral("qrc:/main.qml")));

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

    /* we send json as bytes because it breaks as a simple string */
    QString QProvidersJSON(providers->json().toJson(QJsonDocument::Compact));

    /* let the Go side initialize its internal state */
    InitializeBitmaskContext(
            toGoStr(defaultProvider.toString()),
            (char*)providerJsonBytes.data(), providerJsonBytes.length(),
            obfs4, disAutostart, toGoStr(startVPN));

    /* if requested, enable web api for controlling the VPN */
    if (webAPI)
        EnableWebAPI(toGoStr(webPort));

    if (engine.rootObjects().isEmpty())
        return -1;

    /* kick off your shoes, put your feet up */
    return app.exec();
}
