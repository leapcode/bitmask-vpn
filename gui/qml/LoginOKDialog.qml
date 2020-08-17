import QtQuick 2.0
import QtQuick.Dialogs 1.2
import QtQuick.Controls 1.4

Dialog {
    standardButtons: StandardButton.Ok
    title: qsTr("Login Successful")
    Column {
        anchors.fill: parent
        Text {
            text: qsTr("Login successful. You can now start the VPN.")
        }
    }

    // TODO implement cleanNotifications on backend
    function _loginOk() {
        loginDone = true;
    }

    visible: false
    onAccepted: _loginOk()
    onRejected: _loginOk()
}
