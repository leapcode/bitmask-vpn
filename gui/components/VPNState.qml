import QtQuick 2.0
import QtQuick.Controls 2.12

import "../themes/themes.js" as Theme

StateGroup {
    property var initializing: "initializing"
    property var off: "off"
    property var on: "on"
    property var starting: "starting"
    property var stopping: "stopping"
    property var failed: "failed"

    state: ctx ? ctx.status : vpnStates.off

    states: [
        State {
            name: initializing
        },
        State {
            name: off
            PropertyChanges {
                target: connectionState
                text: qsTr("Connection\nUnsecured")
            }
            PropertyChanges {
                target: statusBoxBackground
                border.color: Theme.accentOff
            }
            PropertyChanges {
                target: connectionImage
                source: "../resources/spy.gif"
            }
            PropertyChanges {
                target: toggleVPN
                text: qsTr("Turn on")
            }
            PropertyChanges {
                target: systray
                icon.source: icons["off"]
            }
            PropertyChanges {
                target: systray.statusItem
                text: toHuman("off")
            }
            StateChangeScript {
                script: {}
            }
        },
        State {
            name: on
            PropertyChanges {
                target: connectionState
                text: qsTr("Connection\nSecured")
            }
            PropertyChanges {
                target: statusBoxBackground
                border.color: Theme.accentOn
            }
            PropertyChanges {
                target: connectionImage
                source: "../resources/riseup-icon.svg"
            }
            PropertyChanges {
                target: toggleVPN
                text: qsTr("Turn off")
            }
            PropertyChanges {
                target: systray
                tooltip: toHuman("on")
                icon.source: icons["on"]
            }
            PropertyChanges {
                target: systray.statusItem
                text: toHuman("on")
            }
            StateChangeScript {
                script: {}
            }
        },
        State {
            name: starting
            PropertyChanges {
                target: connectionState
                text: qsTr("Connecting")
            }
            PropertyChanges {
                target: statusBoxBackground
                border.color: Theme.accentConnecting
            }
            PropertyChanges {
                target: connectionImage
                source: "../resources/birds.svg"
                anchors.horizontalCenter: parent.horizontalCenter
            }
            PropertyChanges {
                target: toggleVPN
                text: qsTr("Cancel")
            }
            PropertyChanges {
                target: systray
                tooltip: toHuman("connecting")
                icon.source: icons["wait"]
            }
            PropertyChanges {
                target: systray.statusItem
                text: toHuman("connecting")
            }
            StateChangeScript {
                script: {}
            }
        },
        State {
            name: stopping
            PropertyChanges {
                target: connectionState
                text: "Switching\nOff"
            }
            PropertyChanges {
                target: statusBoxBackground
                border.color: Theme.accentConnecting
            }
            PropertyChanges {
                // ?? is this image correct?
                target: connectionImage
                source: "../resources/birds.svg"
                anchors.horizontalCenter: parent.horizontalCenter
            }
            PropertyChanges {
                target: systray
                tooltip: toHuman("stopping")
                icon.source: icons["wait"]
            }
            PropertyChanges {
                target: systray.statusItem
                text: toHuman("stopping")
            }
        },
        State {
            name: failed
        }
    ]
    transitions: [
        Transition {
            to: on
            ColorAnimation {
                target: statusBoxBackground
                duration: 500
            }
        },
        Transition {
            to: off
            ColorAnimation {
                target: statusBoxBackground
                duration: 500
            }
        },
        Transition {
            to: starting
            ColorAnimation {
                target: statusBoxBackground
                duration: 500
            }
        },
        Transition {
            to: stopping
            ColorAnimation {
                target: statusBoxBackground
                duration: 500
            }
        }
    ]
    function toHuman(st) {
        switch (st) {
        case "off":
            //: %1 -> application name
            return ctx ? qsTr("%1 off").arg(ctx.appName) : qsTr("off")
        case "on":
            //: %1 -> application name
            return qsTr("%1 on").arg(ctx.appName)
        case "connecting":
            //: %1 -> application name
            return qsTr("Connecting to %1").arg(ctx.appName)
        case "stopping":
            //: %1 -> application name
            return qsTr("Stopping %1").arg(ctx.appName)
        case "failed":
            //: %1 -> application name
            return qsTr("%1 blocking internet").arg(
                        ctx.appName) // TODO failed is not handled yet
        }
    }
}
