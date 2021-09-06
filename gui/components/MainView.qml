import QtQuick 2.0
import QtQuick.Controls 2.4
import QtQuick.Dialogs 1.2
import QtQuick.Controls.Material 2.1
import QtQuick.Layouts 1.14

Page {

    StackView {
        id: stackView
        anchors.fill: parent
        initialItem: Home {}
    }

    Drawer {
        id: settingsDrawer

        width: Math.min(root.width, root.height) / 3 * 2
        height: root.height

        ListView {
            focus: true
            currentIndex: -1
            anchors.fill: parent

            delegate: ItemDelegate {
                width: parent.width
                text: model.text
                visible: {
                    if (isDonationService) {return true}
                    return model.text != qsTr("Donate")
                }
                highlighted: ListView.isCurrentItem
                icon.color: "transparent"
                icon.source: model.icon
                onClicked: {
                    settingsDrawer.close()
                    model.triggered()
                }
            }

            model: ListModel {
                ListElement {
                    text: qsTr("Preferences")
                    icon: "../resources/tools.svg"
                    triggered: function () {
                        stackView.push("Preferences.qml")
                    }
                }
                ListElement {
                    text: qsTr("Donate")
                    icon: "../resources/donate.svg"
                    triggered: function () {
                        aboutDialog.open()
                    }
                }
                ListElement {
                    text: qsTr("Help")
                    icon: "../resources/help.svg"
                    triggered: function () {
                        stackView.push("Help.qml")
                        settingsDrawer.close()
                    }
                } // -> can link to another dialog with report bug / support / contribute / FAQ
                ListElement {
                    text: qsTr("About")
                    icon: "../resources/about.svg"
                    triggered: function () {
                        aboutDialog.open()
                    }
                }
                ListElement {
                    text: qsTr("Quit")
                    icon: "../resources/quit.svg"
                    triggered: function () {
                        Qt.callLater(backend.quit)
                    }
                }
            }

            ScrollIndicator.vertical: ScrollIndicator {}
        }
    }

    header: Header {
        id: header
    }
    footer: Footer {
        id: footer
    }

    Dialog {
        id: aboutDialog
        title: qsTr("About")
        Label {
            anchors.fill: parent
            text: qsTr("RiseupVPN\nhttps://riseupvpn.net/vpn")
            horizontalAlignment: Text.AlignHCenter
        }

        standardButtons: StandardButton.Ok
    }
}
