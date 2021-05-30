import QtQuick 2.9
import QtQuick.Controls 2.4

Rectangle {

    anchors.fill: parent;
    anchors.topMargin: 40;

    Image {
        source: "qrc:/assets/img/bird.jpg";
        fillMode: Image.PreserveAspectCrop;
        anchors.fill: parent; 
        opacity: 0.8;
    }
}
