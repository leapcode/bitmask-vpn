import QtQuick 2.9
import QtQuick.Controls 2.2

Page {
    title: qsTr("Preferences")

    Column {
        spacing: 2
        topPadding: root.width * 0.2
        leftPadding: root.width * 0.15
        rightPadding: root.width * 0.15

        Label {
            text: qsTr("Anti-censorship")
            font.bold: true
        }

        CheckBox {
            checked: false
            text: qsTr("Use Bridges")
        }
    }
}
