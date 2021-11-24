import QtQuick 2.15
import QtQuick.Controls 2.2
import QtQuick.Layouts 1.14
import QtQuick.Controls.Material 2.1

import "../themes/themes.js" as Theme

ThemedPage {
    title: qsTr("Preferences")

    Rectangle {
        anchors.horizontalCenter: parent.horizontalCenter
        width: root.appWidth * 0.80
        // FIXME - just the needed height
        height: getBoxHeight()
        radius: 10
        color: "white"

        anchors {
            fill: parent
            margins: 10
        }

        ColumnLayout {
            id: prefCol
            width: root.appWidth * 0.80

            Rectangle {
                id: turnOffWarning
                visible: false
                height: 20
                width: parent.width
                color: "white"

                Label {
                    color: "red"
                    text: qsTr("Turn off the VPN to make changes")
                    width: prefCol.width
                }
                Layout.topMargin: 10
                Layout.leftMargin: 10
                Layout.rightMargin: 10
            }

            Label {
                id: circumLabel
                text: qsTr("Censorship circumvention")
                font.bold: true
                Layout.topMargin: 10
                Layout.leftMargin: 10
                Layout.rightMargin: 10
            }

            Label {
                text: qsTr("These techniques can bypass censorship, but are slower. Please use them only if needed.")
                color: "gray"
                visible: true
                wrapMode: Text.Wrap
                font.pixelSize: Theme.fontSize - 3
                Layout.leftMargin: 10
                Layout.rightMargin: 10
                Layout.preferredWidth: 240
            }

            MaterialCheckBox {
                id: useBridgesCheckBox
                enabled: areBridgesAvailable()
                checked: false
                text: qsTr("Use obfs4 bridges")
                // TODO refactor - this sets wrapMode on checkbox
                contentItem: Label {
                    text: useBridgesCheckBox.text
                    font: useBridgesCheckBox.font
                    horizontalAlignment: Text.AlignLeft
                    verticalAlignment: Text.AlignVCenter
                    leftPadding: useBridgesCheckBox.indicator.width + useBridgesCheckBox.spacing
                    wrapMode: Label.Wrap
                }
                Layout.leftMargin: 10
                Layout.rightMargin: 10
                HoverHandler {
                    cursorShape: Qt.PointingHandCursor
                }
                onClicked: {
                    // TODO there's a corner case that needs to be dealt with in the backend,
                    // if an user has a manual location selected and switches to bridges:
                    // we need to fallback to "auto" selection if such location does not
                    // offer bridges
                    useBridges(checked)
                    useUDP.enabled = !checked
                }
            }

            MaterialCheckBox {
                id: useSnowflake
                //wrapMode: Label.Wrap
                text: qsTr("Use Snowflake (experimental)")
                enabled: false
                checked: false
                HoverHandler {
                    cursorShape: Qt.PointingHandCursor
                }
                Layout.leftMargin: 10
                Layout.rightMargin: 10
            }

            Label {
                text: qsTr("Transport")
                font.bold: true
                Layout.leftMargin: 10
                Layout.rightMargin: 10
            }

            Label {
                text: qsTr("UDP can make the VPN faster, but it might be blocked on certain networks")
                width: parent.width
                color: "gray"
                visible: true
                wrapMode: Text.Wrap
                font.pixelSize: Theme.fontSize - 3
                Layout.leftMargin: 10
                Layout.rightMargin: 10
                Layout.preferredWidth: 240
            }

            MaterialCheckBox {
                id: useUDP
                text: qsTr("UDP")
                enabled: false
                checked: false
                Layout.leftMargin: 10
                Layout.rightMargin: 10
                HoverHandler {
                    cursorShape: Qt.PointingHandCursor
                }
                onClicked: {
                    doUseUDP(checked)
                    useBridgesCheckBox.enabled = areBridgesAvailable()
                }
            }
        }
    }

    StateGroup {
        state: ctx ? ctx.status : "off"
        states: [
            State {
                name: "on"
                PropertyChanges {
                    target: turnOffWarning
                    visible: true
                }
                PropertyChanges {
                    target: useBridgesCheckBox
                    enabled: false
                }
                PropertyChanges {
                    target: useUDP
                    enabled: false
                }
            },
            State {
                name: "starting"
                PropertyChanges {
                    target: turnOffWarning
                    visible: true
                }
                PropertyChanges {
                    target: useBridgesCheckBox
                    enabled: false
                }
                PropertyChanges {
                    target: useUDP
                    enabled: false
                }
            },
            State {
                name: "off"
                PropertyChanges {
                    target: turnOffWarning
                    visible: false
                }
                PropertyChanges {
                    target: useBridgesCheckBox
                    enabled: true
                }
                PropertyChanges {
                    target: useUDP
                    enabled: true
                }
            }
        ]
    }

    function areBridgesAvailable() {
        // FIXME check if provider offers it
        let providerSupport = true
        return providerSupport && !useUDP.checked
    }

    function useBridges(value) {
        if (value == true) {
            console.debug("use obfs4")
            backend.setTransport("obfs4")
        } else {
            console.debug("use regular")
            backend.setTransport("openvpn")
        }
    }

    function doUseUDP(value) {
        if (value == true) {
            console.debug("use udp")
            backend.setUDP(true)
        } else {
            console.debug("use tcp")
            backend.setUDP(false)
        }
    }

    function getBoxHeight() {
        return prefCol.height + 15
    }

    Component.onCompleted: {
        if (ctx && ctx.transport == "obfs4") {
            useBridgesCheckBox.checked = true
        }
        if (ctx && ctx.udp == "true") {
            useUDP.checked = true
        }
        if (ctx && ctx.offersUdp == "false") {
            useUDP.enabled = false
        }
    }
}
