import QtQuick 2.0
import QtQuick.Controls 2.4
import QtQuick.Dialogs 1.2
import QtQuick.Controls.Material 2.1

ToolBar {
    visible: stackView.depth > 1
    Material.foreground: Material.Black
    Material.background: "#ffffff"
    Material.elevation: 0

    contentHeight: settingsButton.implicitHeight

    ToolButton {
        id: settingsButton
        anchors.left: parent.left
        font.pixelSize: Qt.application.font.pixelSize * 1.6
        icon.source: "../resources/arrow-left.svg"
        onClicked: {
            if (stackView.depth > 1) {
                stackView.pop()
            } else {
                settingsDrawer.open()
            }
        }
    }

    Label {
        text: stackView.currentItem.title
        anchors.centerIn: parent
    }
}
