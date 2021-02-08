import QtQuick 2.0
import QtQuick.Dialogs 1.2

MessageDialog {
    title: qsTr("Initialization Error")
    modality: Qt.NonModal
    text: ""
    onAccepted: retryOrQuit()
    onRejected: retryOrQuit()

    Component.onCompleted: {
        buttons: MessageDialog.Ok
    }

    function retryOrQuit() {
        if (ctx.loginDialog == 'true') {
            login.visible = true
        } else {
            backend.quit()
        }
    }
}
