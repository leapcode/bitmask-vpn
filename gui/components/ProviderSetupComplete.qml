pragma ComponentBehavior: Bound

import QtQuick
import QtQuick.Controls
import QtQuick.Controls.Material
import "../themes/themes.js" as Theme

Item {
    id: providerSetupComplete

    Rectangle {
        id: pageHeader
        color: "transparent"
        anchors.top: parent.top
        height: needCircumventionLabel.height + needCircumventionMsg.height + 5
        width: parent.width

        Label {
            id: needCircumventionLabel
            text: qsTr("You're all set!")
            font.bold: true
            font.pixelSize: 14
            anchors.top: parent.top
            wrapMode: Text.WordWrap
            leftPadding: 20
            rightPadding: 20
            width: parent.width
            clip: false
        }

        Label {
            id: needCircumventionMsg
            text: qsTr("Click the button below to connect")
            font.pixelSize: 12
            wrapMode: Text.WordWrap
            anchors.top: needCircumventionLabel.bottom
            anchors.topMargin: 10
            width: parent.width
            leftPadding: 20
            rightPadding: 20
            clip: false
        }

        RoundButton {
            id: toggleVPN
            width: 60
            height: 60
            anchors.horizontalCenter: parent.horizontalCenter
            anchors.top: needCircumventionMsg.bottom
            anchors.topMargin: 180
            display: AbstractButton.IconOnly
            Accessible.name: qsTr("Turn on")
            Accessible.role: Accessible.Button
            HoverHandler {
                cursorShape: Qt.PointingHandCursor
            }
            contentItem: Image {
                anchors.fill: parent
                source: Theme.buttonDisconnected
                mipmap: true
            }

            onClicked: {
                mainView.loadMainView();
                mainView.setStatusStarting();
                appsettings.setValue("provider", root.ctx.provider);
                backend.switchOn();
            }
        }
    }
}
