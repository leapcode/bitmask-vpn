import QtQuick 2.9
import QtQuick.Controls 2.4

TabBar {
    width: parent.width
    TabButton {
        text: qsTr("Status")
    }
    TabButton {
        text: qsTr("Location")
    }
    TabButton {
        text: qsTr("Bridges")
    }
}
