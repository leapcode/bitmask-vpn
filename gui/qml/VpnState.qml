import QtQuick 2.0
import QtQuick.Controls 1.4

StateGroup {

    state: ctx ? ctx.status : ""

    states: [
        State {
            name: "initializing"
        },
        State {
            name: "off"
            PropertyChanges {
                target: systray
                tooltip: toHuman("off")
                icon.source: icons["off"]
            }
            PropertyChanges {
                target: vpntoggle
                checked: false
            }
            PropertyChanges {
                target: statusItem
                text: toHuman("off")
            }
            PropertyChanges {
                target: autoSelectionItem
		text: qsTr("Best")
            }
            PropertyChanges {
                target: mainStatus
                text: toHuman("off")
            }
            PropertyChanges {
                target: mainCurrentGateway
                text: ""
            }
        },
        State {
            name: "on"
            PropertyChanges {
                target: systray
                tooltip: toHuman("on")
                icon.source: icons["on"]
            }
            PropertyChanges {
                target: vpntoggle
                checked: true
            }
            PropertyChanges {
                target: statusItem
                text: toHuman("on")
            }
            PropertyChanges {
                target: autoSelectionItem
		text: {
			if (autoSelectionButton.checked) {
				//: %1 -> location to which the client is connected to
				qsTr("Best (%1)").arg(locationStr())
			} else {
				qsTr("Best")
			}
		}
            }
            PropertyChanges {
                target: mainStatus
                text: toHuman("on")
            }
            PropertyChanges {
                target: mainCurrentGateway
		//: %1 -> location to which the client is connected to
                text: qsTr("Connected to %1").arg(locationStr())
            }
        },
        State {
            name: "starting"
            PropertyChanges {
                target: systray
                tooltip: toHuman("connecting")
                icon.source: icons["wait"]
            }
            PropertyChanges {
                target: statusItem
                text: toHuman("connecting")
            }
            PropertyChanges {
                target: autoSelectionItem
		text: {
			if (autoSelectionButton.checked) {
				//: %1 -> location to which the client is connected to
				qsTr("Best (%1)").arg(locationStr())
			} else {
				qsTr("Best")
			}
		}
            }
            PropertyChanges {
                target: mainStatus
                text: qsTr("Connecting...")
            }
            PropertyChanges {
                target: mainCurrentGateway
                text: ""
            }
        },
        State {
            name: "stopping"
            PropertyChanges {
                target: systray
                tooltip: toHuman("stopping")
                icon.source: icons["wait"]
            }
            PropertyChanges {
                target: statusItem
                text: toHuman("stopping")
            }
            PropertyChanges {
                target: autoSelectionItem
		text: qsTr("Best")
            }
            PropertyChanges {
                target: mainStatus
                text: toHuman("stopping")
            }
            PropertyChanges {
                target: mainCurrentGateway
                text: ""
            }
        },
        State {
            name: "failed"
            PropertyChanges {
                target: systray
                tooltip: toHuman("failed")
                icon.source: icons["wait"]
            }
            PropertyChanges {
                target: statusItem
                text: toHuman("failed")
            }
            PropertyChanges {
                target: autoSelectionItem
		text: qsTr("Best")
            }
            PropertyChanges {
                target: mainStatus
                text: toHuman("failed")
            }
            PropertyChanges {
                target: mainCurrentGateway
                text: ""
            }
        }
    ]
}
