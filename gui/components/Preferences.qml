import QtQuick 2.9
import QtQuick.Controls 2.2
import QtQuick.Controls.Material 2.1

import "../themes/themes.js" as Theme

ThemedPage {
    title: qsTr("Preferences")

    Column {
        id: prefCol
        // FIXME checkboxes in Material style force lineHeights too big.
        // need to override the style
        // See: https://bugreports.qt.io/browse/QTBUG-95385
        topPadding: root.width * 0.05
        leftPadding: root.width * 0.1
        rightPadding: root.width * 0.15

        Rectangle {
            id: turnOffWarning
            visible: false
            height: 40
            width: 300
            color: Theme.bgColor

            anchors.horizontalCenter: parent.horizontalCenter

            Label {
                color: "red"
                text: qsTr("Turn off the VPN to make changes")
                width: prefCol.width
            }
        }


        Label {
            text: qsTr("Anti-censorship")
            font.bold: true
        }

        CheckBox {
            id: useBridgesCheckBox
            checked: false
            text: qsTr("Use obfs4 bridges")
            onClicked: {
                // TODO there's a corner case that needs to be dealt with in the backend,
                // if an user has a manual location selected and switches to bridges:
                // we need to fallback to "auto" selection if such location does not 
                // offer bridges
                useBridges(checked)
            }
        }

        CheckBox {
            id: useSnowflake
            text: qsTr("Use Snowflake (experimental)")
            enabled: false
            checked: false
        }

        Label {
            text: qsTr("Transport")
            font.bold: true
        }

        CheckBox {
            id: useUDP
            text: qsTr("UDP")
            enabled: false
            checked: false
            onClicked: {
                doUseUDP(checked)
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

    Component.onCompleted: {
        if (ctx && ctx.transport == "obfs4") {
            useBridgesCheckBox.checked = true
        }
    }
}
