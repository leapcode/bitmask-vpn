import QtQuick 2.9
import QtQuick.Controls 2.2

Page {
    title: qsTr("Help")

    Column {
        anchors.centerIn: parent
        spacing: 10

        Button {
            anchors.horizontalCenter: parent.horizontalCenter
            text: qsTr("Donate")
            onClicked: stackView.push("Donate.qml")
        }
        Button {
            anchors.horizontalCenter: parent.horizontalCenter
            text: qsTr("Terms of Service")
            onClicked: stackView.push("Donate.qml")
        }
        Button {
            anchors.horizontalCenter: parent.horizontalCenter
            text: qsTr("Contact Support")
            onClicked: stackView.push("Donate.qml")
        }
        Button {
            anchors.horizontalCenter: parent.horizontalCenter
            text: qsTr("Report bug")
            onClicked: stackView.push("Donate.qml")
        }
    }
}
