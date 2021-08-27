import QtQuick 2.9
import QtQuick.Controls 2.4

Rectangle {

    anchors.fill: parent
    anchors.topMargin: 40

    property var backgroundSrc
    property var backgroundVisible

    Image {
        source: parent.backgroundSrc
        visible: parent.backgroundVisible
        fillMode: Image.PreserveAspectCrop
        anchors.fill: parent
        opacity: 0.8
    }

    Component.onCompleted: {
        /* default for riseup, needs customizing */
        backgroundSrc = "qrc:/assets/img/bird.jpg"
    }
}
