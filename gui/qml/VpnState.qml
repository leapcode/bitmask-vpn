import QtQuick 2.0
import QtQuick.Controls 1.4

StateGroup {

    state: ctx ? ctx.status : ""

    states: [
        State { name: "initializing" },
        State {
            name: "off"
            PropertyChanges { target: systray; tooltip: toHuman("off"); icon.source: icons["off"] }
            PropertyChanges { target: statusItem; text: toHuman("off") }
            PropertyChanges { target: mainStatus; text: toHuman("off") }
            PropertyChanges { target: mainCurrentGateway; text: "" }
            PropertyChanges { target: mainOnBtn; visible: true }
            PropertyChanges { target: mainOffBtn; visible: false }
            PropertyChanges { target: gwMarker; color: "red"}
        },
        State {
            name: "on"
            StateChangeScript {
                script: displayGatewayMarker()
            }
            PropertyChanges { target: systray; tooltip: toHuman("on"); icon.source: icons["on"] }
            PropertyChanges { target: statusItem; text: toHumanWithLocation("on") }
            PropertyChanges { target: mainStatus; text: toHuman("on") }
            PropertyChanges { target: mainCurrentGateway; text: qsTr("Connected to ") + ctx.currentGateway }
            PropertyChanges { target: mainOnBtn; visible: false }
            PropertyChanges { target: mainOffBtn; visible: true }
            PropertyChanges { target: gwMarker; color: "green"}
        },
        State {
            name: "starting"
            PropertyChanges { target: systray; tooltip: toHuman("connecting"); icon.source: icons["wait"] }
            PropertyChanges { target: statusItem; text: toHumanWithLocation("connecting") }
            PropertyChanges { target: mainStatus; text: qsTr("Connecting...") }
            PropertyChanges { target: mainCurrentGateway; text: "" }
            PropertyChanges { target: mainOnBtn; visible: false }
            PropertyChanges { target: mainOffBtn; visible: true }
            PropertyChanges { target: gwMarker; color: "orange"}
        },
        State {
            name: "stopping"
            PropertyChanges { target: systray; tooltip: toHuman("stopping"); icon.source: icons["wait"] }
            PropertyChanges { target: statusItem; text: toHuman("stopping") }
            PropertyChanges { target: mainStatus; text: toHuman("stopping") }
            PropertyChanges { target: mainCurrentGateway; text: "" }
            PropertyChanges { target: mainOnBtn; visible: true }
            PropertyChanges { target: mainOffBtn; visible: false }
            PropertyChanges { target: gwMarker; color: "orange"}
        },
        State {
            name: "failed"
            PropertyChanges { target: systray; tooltip: toHuman("failed"); icon.source: icons["wait"] }
            PropertyChanges { target: statusItem; text: toHuman("failed") }
            PropertyChanges { target: mainStatus; text: toHuman("failed") }
            PropertyChanges { target: mainCurrentGateway; text: "" }
            PropertyChanges { target: mainOnBtn; visible: true }
            PropertyChanges { target: mainOffBtn; visible: false }
            PropertyChanges { target: gwMarker; color: "red"}
        }
    ]
}
