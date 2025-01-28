import QtQuick
import QtQuick.Controls
import QtQuick.Effects

Page {
    StatusBox {
        id: statusbox
    }

    function setStatusBoxStateStarting() {
        statusbox.setStatusStarting()
    }
}
