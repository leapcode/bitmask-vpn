import QtQuick
import QtQuick.Controls
import QtQuick.Effects
import QtQuick.Layouts
import QtQuick.Templates as T
import QtQuick.Controls.impl
import QtQuick.Controls.Material
import QtQuick.Controls.Material.impl
import "../themes/themes.js" as Theme

Item {
    id: statusbox
    anchors.fill: parent

    Rectangle {
        id: statusBoxBackground
        anchors.fill: parent
        Image {
            id: backgroundImage
            anchors.fill: parent
            source: customTheme.bgDisconnected
        }
    }

    VPNState {
        id: vpn
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
            settingsDrawer.open();
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
            FadeBehavior on text {
            }
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
            width: parent.width
            speed: 0.8
            source: customTheme.iconOff
            anchors.horizontalCenter: parent.horizontalCenter
            fillMode: Image.PreserveAspectFit
            OpacityAnimator on opacity {
                id: fadeIn
                from: 0.5
                to: 1
                duration: 1000
            }
            onStatusChanged: {
                playing = (status == AnimatedImage.Ready);
                fadeIn.start();
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
                    backend.switchOff();
                } else if (vpn.state === "off") {
                    vpn.startingUI = true;
                    backend.switchOn();
                } else {
                    console.debug("unknown state");
                }
            }
        }
    }

    Footer {
        id: footer
        anchors {
            bottom: parent.bottom
            bottomMargin: 10
            left: parent.left
            leftMargin: 9
            right: parent.right
            rightMargin: 8
        }
    }

    function isSnowflakeOn() {
        return ctx != undefined && ctx.snowflake == "true" && ctx.snowflakeProgress != "100";
    }
}
