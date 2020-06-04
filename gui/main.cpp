#include <QGuiApplication>
#include <QQmlApplicationEngine>
#include <QQuickWindow>
#include <QTimer>
#include <QtQml>
#include <string>

#include "handlers.h"
#include "qjsonmodel.h"
#include "lib/libgoshim.h"

/* Hi! I'm Troy McClure and I'll be your guide today. You probably remember me
   from blockbusters like "here be dragons" and "darling, I wrote a little
   contraption". */

/* Our glorious global object state. In here we store a serialized snapshot of
   the context from the application "backend", living in the linked Go-land
   lib. */

static char *json;

/* We are interested in observing changes to this global json variable.
   The jsonWatchdog bridges the gap from pure c callbacks to the rest of the c++
   logic. QJsonWatch comes from QObject so it can emit signals. */

QJsonWatch *qw;

struct jsonWatchdog {
    jsonWatchdog() { qw = new QJsonWatch; }
    void changed() { emit qw->jsonChanged(QString(json)); }
};

/* we need C wrappers around every C++ object, so that we can invoke their methods
   from the function pointers passed as callbacks to CGO. */
extern "C" {
static void *newWatchdog(void) { return (void *)(new jsonWatchdog); }
static void jsonChanged(void *ptr) {
    if (ptr != NULL) {
        jsonWatchdog *klsPtr = static_cast<jsonWatchdog *>(ptr);
        klsPtr->changed();
    }
}
}

void *wd = newWatchdog();

/* onStatusChanged is the C function that we register as a callback with CGO,
   to be called from the Go side. It pulls a string serialization of the
   context object, than we then pass along to Qt objects and to Qml. */
void onStatusChanged() {
    char *ctx = RefreshContext();
    json = ctx;
    /* the method wrapped emits a qt signal */
    jsonChanged(wd);
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

    QGuiApplication::setAttribute(Qt::AA_EnableHighDpiScaling);
    QGuiApplication app(argc, argv);
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
    QObject::connect(qw, &QJsonWatch::jsonChanged, [ctx, model](QString js) {
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
