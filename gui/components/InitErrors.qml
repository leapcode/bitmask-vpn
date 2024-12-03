import QtQuick
import QtQuick.Controls
import QtQuick.Effects

ErrorBox {

    state: "noerror"

    states: [
        State {
            name: "noerror"
            when: root.error == ""
            PropertyChanges {
                target: splashProgress
                visible: true
            }
            PropertyChanges {
                target: splashErrorBox
                visible: false
            }
        },
        State {
            name: "nohelpers"
            when: root.error == "nohelpers"
            PropertyChanges {
                target: splashProgress
                visible: false
            }
            PropertyChanges {
                target: splashErrorBox
                errorText: qsTr("Could not find helpers. Please check your installation")
                visible: true 
            }
        },
        State {
            name: "nopolkit"
            when: root.error == "nopolkit"
            PropertyChanges {
                target: splashSpinner
                visible: false
            }
            PropertyChanges {
                target: splashErrorBox
                errorText: qsTr("Could not find polkit agent.")
                visible: true 
            }
        },
        State {
            name: "alreadyrunning"
            when: root.error == "alreadyrunning"
            PropertyChanges {
                target: splashProgress
                visible: false
            }
            PropertyChanges {
                target: splashErrorBox
                errorText: qsTr("Application is going to quit as another instance is already running. Please use the system tray icon to open it")
                visible: true 
            }
        }
    ]
}
