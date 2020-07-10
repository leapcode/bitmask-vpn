import QtQuick 2.0
import QtQuick.Dialogs 1.2
import QtQuick.Controls 1.4

Dialog {
    standardButtons: StandardButton.Ok
    title: qsTr("Login Success")
    text: qsTr("You are now logged in, connecting now")

    visible: ctxSystray.loginConfirmationDialog == true
}
