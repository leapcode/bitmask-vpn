import QtQuick 2.9
import QtQuick.Controls 2.2
import QtQuick.Controls.Material 2.1

Page {
    title: qsTr("Preferences")

    Column {
        id: prefCol
        // FIXME the checkboxes seem to have a bigger lineHeight themselves, need to pack more.
        spacing: 1
        topPadding: root.width * 0.1
        leftPadding: root.width * 0.15
        rightPadding: root.width * 0.15

        Rectangle {
            id: turnOffWarning
            visible: false
            height: 40
            width: 300

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
        }

        CheckBox {
            id: useSnowflake
            checked: false
            text: qsTr("Use Snowflake (experimental)")
        }

        Label {
            text: qsTr("Transport")
            font.bold: true
        }

        CheckBox {
            id: useUDP
            checked: false
            text: qsTr("UDP")
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
                    checkable: false
                }
                PropertyChanges {
                    target: useUDP
                    checkable: false
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
                    checkable: false
                }
                PropertyChanges {
                    target: useUDP
                    checkable: false
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
                    checkable: true
                }
                PropertyChanges {
                    target: useUDP
                    checkable: true
                }
            }
        ]
    }
}
