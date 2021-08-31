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
        height: 300
        radius: 10
        color: Theme.bgColor
        antialiasing: true
        anchors {
            fill: parent
            margins: 20
            bottomMargin: 30
        }
        border {
            color: Theme.accentOff
            width: 2
        }
    }

    ToolButton {
        id: settingsButton
        objectName: "settingsButton"
        font.pixelSize: Qt.application.font.pixelSize * 1.2
        opacity: 1

        anchors {
            top: parent.top
            left: parent.left
            topMargin: Theme.windowMargin + 10
            leftMargin: Theme.windowMargin + 10
        }

        onClicked: {
            if (stackView.depth > 1) {
                stackView.pop()
            } else {
                settingsDrawer.open()
            }
        }

        Icon {
            id: settingsImage
            width: 16
            height: 16
            anchors.centerIn: settingsButton
            source: stackView.depth
                    > 1 ? "../resources/arrow-left.svg" : "../resources/gear-fill.svg"
        }
    }

    Rectangle {
        id: statusLabelWrapper
        height: 45
        anchors {
            top: statusBoxBackground.top
            topMargin: 40
            horizontalCenter: parent.horizontalCenter
        }
        BoldLabel {
            id: connectionState
            anchors.top: parent.top
            anchors.horizontalCenter: parent.horizontalCenter
            horizontalAlignment: Text.AlignHCenter
            text: ""
        }
    }

    Column {
        id: col
        width: parent.width * 0.8
        anchors.horizontalCenter: parent.horizontalCenter

        VerticalSpacer {
            id: spacerPreImg
            visible: true
            height: 150
        }

        Image {
            id: connectionImage
            height: 160
            source: "../resources/spy.gif"
            anchors.horizontalCenter: parent.horizontalCenter
            fillMode: Image.PreserveAspectFit
        }

        VerticalSpacer {
            id: spacerPostImg
            visible: true
            height: 30
        }

        MaterialButton {
            id: toggleVPN
            spacing: 8

            anchors.horizontalCenter: parent.horizontalCenter
            Layout.alignment: Qt.AlignBottom

            font {
                capitalization: Font.Capitalize
                family: lightFont.name
                bold: false
            }

            onClicked: {
                if (vpn.state === "on") {
                    backend.switchOff()
                } else if (vpn.state === "off") {
                    backend.switchOn()
                } else {
                    console.debug("unknown state")
                }
            }
        }
    }
}
