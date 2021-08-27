import QtQuick 2.9
import QtQuick.Controls 2.2
import QtGraphicalEffects 1.0

Page {
    id: splash
    property int timeoutInterval: 1600

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

        Spinner {}
    }

    Timer {
        id: splashTimer
    }

    function delay(delayTime, cb) {
        splashTimer.interval = delayTime
        splashTimer.repeat = false
        splashTimer.triggered.connect(cb)
        splashTimer.start()
    }

    function loadMainViewWhenReady() {
        console.debug("ready?")
        if (ctx && ctx.isReady) {
            console.debug("ready?", ctx.isReady)
            // FIXME check errors == None
            loader.source = "MainView.qml"
        } else {
            delay(100, loadMainViewWhenReady)
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

    Component.onCompleted: {

    }
}
