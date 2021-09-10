import QtQuick 2.9
import QtQuick.Controls 2.2
import QtGraphicalEffects 1.0

Page {
    id: splash
    property int timeoutInterval: qmlDebug ? 200 : 1600
    property alias errors: splashErrorBox

    Column {
        width: parent.width * 0.8
        anchors.horizontalCenter: parent.horizontalCenter
        anchors.topMargin: 24

        VerticalSpacer {
            visible: true
            height: root.height * 0.10
        }

        Image {
            id: connectionImage
            height: 200
            anchors.horizontalCenter: parent.horizontalCenter
            source: "../resources/icon-noshield.svg"
            fillMode: Image.PreserveAspectFit
        }

        Spinner {
            id: splashSpinner
        }

        InitErrors {
            id: splashErrorBox
        }
    }

    Timer {
        id: splashTimer
    }

    function delay(delayTime, cb) {
        splashTimer.interval = delayTime
        splashTimer.repeat = true
        splashTimer.triggered.connect(cb)
        splashTimer.start()
    }

    function loadMainViewWhenReady() {
        if (root.error != "") {
            return
        }
        if (ctx && ctx.isReady) {
            splashTimer.stop()
            loader.source = "MainView.qml"
        } else {
            if (!splashTimer.running) {
              console.debug('delay...')
              delay(500, loadMainViewWhenReady)
            }
        }
    }

    Timer {
        interval: timeoutInterval
        running: true
        repeat: false
        onTriggered: {
            loadMainViewWhenReady()
        }
    }

    Component.onCompleted: {}
}
