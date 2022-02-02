import QtQuick 2.15
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
        color: customTheme.bgColor
        anchors.fill: parent
    }

    Rectangle {
        id: statusBoxBackground
        color: customTheme.fgColor
        height: 300
        radius: 10
        antialiasing: true
        anchors {
            fill: parent
            margins: 20
            bottomMargin: 30
        }
        border {
            color: Theme.accentOff
            width: 4
        }
    }

    ToolButton {
        id: settingsButton
        objectName: "settingsButton"
        font.pixelSize: Qt.application.font.pixelSize * 1.6
        opacity: 1
        anchors {
            top: parent.top
            left: parent.left
            topMargin: Theme.windowMargin + 5
            leftMargin: Theme.windowMargin + 5
        }
        HoverHandler {
            cursorShape: Qt.PointingHandCursor
        }
        onClicked: {
            settingsDrawer.toggle()
        }

        Icon {
            id: settingsImage
            width: 16
            height: 16
            anchors.centerIn: settingsButton
            source: "../resources/gear-fill.svg"
        }
    }

    Rectangle {
        id: statusLabelWrapper
        height: 45
        anchors {
            top: statusBoxBackground.top
            topMargin: 25
            horizontalCenter: parent.horizontalCenter
        }
        BoldLabel {
            id: connectionState
            anchors.top: parent.top
            anchors.horizontalCenter: parent.horizontalCenter
            horizontalAlignment: Text.AlignHCenter
            FadeBehavior on text { }
        }
        Label {
            id: snowflakeTip
            anchors.top: connectionState.bottom
            anchors.horizontalCenter: parent.horizontalCenter
            anchors.topMargin: 20
            horizontalAlignment: Text.AlignHCenter
            text: qsTr("This can take several minutes")
            font.pixelSize: Theme.fontSize * 0.8
            visible: isSnowflakeOn()
        }
        ProgressBar {
            id: snowflakeProgressBar
            anchors.top: snowflakeTip.bottom
            anchors.horizontalCenter: parent.horizontalCenter
            visible: isSnowflakeOn()
            value: 0
        }
        Label {
            id: snowflakeTag
            anchors.top: snowflakeProgressBar.bottom
            anchors.horizontalCenter: parent.horizontalCenter
            horizontalAlignment: Text.AlignHCenter
            visible: isSnowflakeOn()
        }
    }

    Column {
        id: col
        width: parent.width * 0.8
        anchors.horizontalCenter: parent.horizontalCenter

        VerticalSpacer {
            id: spacerPreImg
            visible: true
            height: 120
        }

        // TODO this can be synced with opacity serial animation, see
        // https://doc.qt.io/qt-5/qml-qtquick-animatedimage.html#example-usage
        // If you want to customize your asset, here's how:
        // convert -delay 50 -loop 0 ravens2_*.png ravens.gif

        AnimatedImage {
            id: connectionImage
            height: 160
            speed: 0.8
            source: customTheme.iconOff
            anchors.horizontalCenter: parent.horizontalCenter
            fillMode: Image.PreserveAspectFit
            OpacityAnimator on opacity{
                id: fadeIn
                from: 0.5;
                to: 1;
                duration: 1000
            }
            onStatusChanged: {
                playing = (status == AnimatedImage.Ready)
                fadeIn.start()
            }
        }

        VerticalSpacer {
            id: spacerPostImg
            visible: true
            height: 20
            Layout.alignment: Qt.AlignBottom
        }

        MaterialButton {
            id: toggleVPN
            // FIXME - this is a workaround. It will BREAK with i18n
            width: 100
            spacing: 8
            anchors.horizontalCenter: parent.horizontalCenter
            Layout.alignment: Qt.AlignBottom
            font {
                pixelSize: Theme.buttonFontSize
                capitalization: Font.Capitalize
                family: lightFont.name
                bold: false
            }
            HoverHandler {
                cursorShape: Qt.PointingHandCursor
            }

            onClicked: {
                if (vpn.state === "on" | vpn.state === "starting") {
                    backend.switchOff()
                } else if (vpn.state === "off") {
                    vpn.startingUI = true
                    backend.switchOn()
                } else {
                    console.debug("unknown state")
                }
            }
        }
    }

    function isSnowflakeOn() {
        return ctx != undefined && ctx.snowflake == "true" && ctx.snowflakeProgress != "100"
    }
}
