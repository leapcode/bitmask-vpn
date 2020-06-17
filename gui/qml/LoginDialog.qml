import QtQuick 2.0
import QtQuick.Dialogs 1.2
import QtQuick.Controls 1.4

Dialog {
    standardButtons: StandardButton.Ok
    title: "Login"
    Column {
        anchors.fill: parent
        Text {
            text: "Log in with your library credentials"
        }
        TextField {
            id: username
            placeholderText: "patron id"
        }
        TextField {
            id: password
            placeholderText: "password"
            echoMode: TextInput.PasswordEchoOnEdit
        }
    }

    visible: false
    //visible: ctx.showLogin == true
    //onAccepted: backend.login(username.text, password.text)
    onRejected: backend.quit()  // TODO: it doesn't close
}
