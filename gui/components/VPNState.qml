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

    property bool startingUI: false

    state: ctx ? ctx.status : off

    states: [
        State {
            name: initializing
        },
        State {
            when: ctx != undefined && ctx.status == "off" && startingUI == true
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
                source: customTheme.iconConnecting
                anchors.horizontalCenter: parent.horizontalCenter
            }
            PropertyChanges {
                target: toggleVPN
                enabled: false
                // XXX this is a fake cancel, won't do anything at this point. We need
                // to queue this action for when the openvpn process becomes available.
                text: ("Cancel")
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
        },
        State {
            name: "off"
            PropertyChanges {
                target: connectionState
                text: qsTr("Unsecured\nConnection")
            }
            PropertyChanges {
                target: statusBoxBackground
                border.color: Theme.accentOff
            }
            PropertyChanges {
                target: connectionImage
                source: customTheme.iconOff
            }
            PropertyChanges {
                target: toggleVPN
                enabled: true
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
                script: {
                    console.debug("status off")
                }
            }
        },
        State {
            name: on
            PropertyChanges {
                target: connectionState
                text: qsTr("Secured\nConnection")
            }
            PropertyChanges {
                target: statusBoxBackground
                border.color: Theme.accentOn
            }
            PropertyChanges {
                target: connectionImage
                source: customTheme.iconOn
            }
            PropertyChanges {
                target: toggleVPN
                enabled: true
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
                script: {
                    vpn.startingUI = false
                }
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
                source: customTheme.iconConnecting
                anchors.horizontalCenter: parent.horizontalCenter
            }
            PropertyChanges {
                target: toggleVPN
                enabled: true
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
                script: {
                    vpn.startingUI = false
                }
            }
        },
        State {
            name: stopping
            /*
            this transition is bad. let's just remove the status
            switch...
            PropertyChanges {
                target: connectionState
                text: "Switching\nOff"
            }
            PropertyChanges {
                target: connectionImage
                source: "../resources/ravens.svg"
                anchors.horizontalCenter: parent.horizontalCenter
            }
            */
            PropertyChanges {
                target: statusBoxBackground
                border.color: Theme.accentConnecting
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
