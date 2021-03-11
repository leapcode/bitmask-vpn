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
            // FIXME - we probably want to distinguish 
            // fatal from recoverable errors. For the time being
            // we can avoid quitting so that people can try reconnects if it's
            // a network problem, it's confusing to quit the app.
            // backend.quit()
        }
    }
}
