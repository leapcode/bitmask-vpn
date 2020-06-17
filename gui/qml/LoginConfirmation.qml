import QtQuick 2.0
import QtQuick.Dialogs 1.2
import QtQuick.Controls 1.4

Dialog {
    standardButtons: StandardButton.Ok
    title: "Login Success"
    text: "You are now logged in, connecting now"

    visible: ctxSystray.loginConfirmationDialog == true
}
