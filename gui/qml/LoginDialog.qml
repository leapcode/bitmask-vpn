import QtQuick 2.0
import QtQuick.Dialogs 1.2
import QtQuick.Controls 1.4

Dialog {
    standardButtons: StandardButton.Ok
    title: qsTr("Login")
    Column {
        anchors.fill: parent
        Text {
            text: qsTr("Log in with your library credentials")
        }
        TextField {
            id: username
            placeholderText: qsTr("patron id")
        }
        TextField {
            id: password
            placeholderText: qsTr("password")
            echoMode: TextInput.PasswordEchoOnEdit
        }
    }

    visible: false
    onAccepted: backend.login(username.text, password.text)
    onRejected: backend.quit()
}
