import QtQuick 2.0
import QtQuick.Controls 2.12

import "../themes/themes.js" as Theme

StateGroup {

    state: ctx ? ctx.status : "off"

    states: [
        State {
            name: "initializing"
        },
        State {
            name: "off"
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
                //anchors.right: parent.right
                //anchors.rightMargin: -8
                // XXX need to nulify horizontalcenter somehow,
                // it gets fixed to parent.center
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
                script: {

                }
            }
        },
        State {
            name: "on"
            PropertyChanges {
                target: connectionState
                text: qsTr("Connection\nSecure")
            }
            PropertyChanges {
                target: statusBoxBackground
                border.color: Theme.accentOn
            }
            PropertyChanges {
                target: connectionImage
                source: "../resources/riseup-icon.svg"
                // TODO need to offset the logo or increase the image
                // to fixed height
                height: 120
            }
            PropertyChanges {
                target: spacerPreImg
                visible: true
            }
            PropertyChanges {
                target: spacerPostImg
                visible: true
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
                script: {

                    // TODO check donation
                    //if (needsDonate && !shownDonate) {
                    //    donate.visible = true;
                    //    shownDonate = true;
                    //    backend.donateSeen();
                    //}
                }
            }
        },
        State {
            name: "starting"
            //when: toggleVPN.pressed == true
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
                script: {

                }
            }
        },
        State {
            name: "stopping"
            PropertyChanges {
                target: connectionState
                text: "Switching\nOff"
            }
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
            name: "failed"
            // TODO
        }
    ]


    /*
    transitions: Transition {
        from: "off"
        to: "starting"
        reversible: true

        ParallelAnimation {
            ColorAnimation { duration: 500 }
        }
    }
    */
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
