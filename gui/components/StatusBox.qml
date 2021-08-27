import QtQuick 2.12
import QtQuick.Controls 2.12
import QtGraphicalEffects 1.14
import QtQuick.Layouts 1.14

import QtQuick.Templates 2.12 as T
import QtQuick.Controls.impl 2.12
import QtQuick.Controls.Material 2.12
import QtQuick.Controls.Material.impl 2.12
import "../themes/themes.js" as Theme

Item {
    id: statusbox
    anchors.fill: parent

    VPNState {
        id: vpn
    }

    Rectangle {
        id: statusBoxBackground
        anchors.fill: parent
        anchors.margins: 20
        anchors.bottomMargin: 30
        height: 300
        radius: 10
        color: Theme.bgColor
        border.color: Theme.accentOff
        border.width: 2
        antialiasing: true
    }

    ToolButton {
        id: settingsButton
        objectName: "settingsButton"
        opacity: 1

        font.pixelSize: Qt.application.font.pixelSize * 1.6
        anchors.top: parent.top
        anchors.left: parent.left
        anchors.topMargin: Theme.windowMargin + 10
        anchors.leftMargin: Theme.windowMargin + 10

        onClicked: {
            if (stackView.depth > 1) {
                stackView.pop()
            } else {
                settingsDrawer.open()
            }
        }

        Icon {
            id: settingsImage
            width: 24
            height: 24
            // TODO move arrow left to toolbar top
            source: stackView.depth
                    > 1 ? "../resources/arrow-left.svg" : "../resources/gear-fill.svg"
            anchors.centerIn: settingsButton
        }
    }

    Column {
        id: col
        anchors.centerIn: parent
        anchors.topMargin: 24
        width: parent.width * 0.8

        BoldLabel {
            id: connectionState
            text: ""
            anchors.horizontalCenter: parent.horizontalCenter
            horizontalAlignment: Text.AlignHCenter
        }

        VerticalSpacer {
            id: spacerPreImg
            visible: false
            height: 40
        }

        Image {
            id: connectionImage
            height: 200
            source: "../resources/spy.gif"
            fillMode: Image.PreserveAspectFit
        }

        VerticalSpacer {
            id: spacerPostImg
            visible: false
            height: 35
        }

        MaterialButton {
            id: toggleVPN
            anchors.horizontalCenter: parent.horizontalCenter
            Layout.alignment: Qt.AlignBottom
            font.capitalization: Font.Capitalize
            spacing: 8

            onClicked: {
                if (vpn.state === "on") {
                    console.debug("should turn off")
                    backend.switchOff()
                } else if (vpn.state === "off") {
                    console.debug("should turn on")
                    backend.switchOn()
                } else {
                    console.debug("unknown state")
                }
            }


            /*
             XXX this hijacks click events, so better no pointing for now.
            MouseArea {
                anchors.fill: toggleVPN
                hoverEnabled: true
                cursorShape: !hoverEnabled ? Qt.ForbiddenCursor : Qt.PointingHandCursor
            }
            */
        }
    }
}
