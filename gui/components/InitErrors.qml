import QtQuick 2.9
import QtQuick.Controls 2.2
import Qt5Compat.GraphicalEffects

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
        }
    ]
}
