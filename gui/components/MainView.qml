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
                        donateDialog.open()
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
                        stackView.push("About.qml")
                        settingsDrawer.close()
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
        id: donateDialog 
        width: 350
        title: qsTr("Please donate!")
        standardButtons: Dialog.Ok

        Text {
            id: donateText
            width: 300
            wrapMode: Text.Wrap
            horizontalAlignment: Text.AlignHCenter
            anchors {
                topMargin: 20
                bottomMargin: 40
                horizontalCenter: parent.horizontalCenter
            }
            font.pixelSize: 12
            text: qsTr("This service is paid for entirely by donations from users like you. The cost of running the VPN is approximately 5 USD per person every month, but every little bit counts.")
        }

        Label {
            id: donateURL
            anchors {
                top: donateText.bottom
                topMargin: 20
                horizontalCenter: parent.horizontalCenter
            }
            font.pixelSize: 14
            text: getLink(ctx.donateURL)
            onLinkActivated: Qt.openUrlExternally(ctx.donateURL)
        }


        Image {
            height: 50
            source: "../resources/donate.svg"
            fillMode: Image.PreserveAspectFit
            anchors {
                topMargin: 20
                top: donateURL.bottom
                horizontalCenter: parent.horizontalCenter
            }
        }

        onAccepted: Qt.openUrlExternally(ctx.donateURL)
    }

    function getLink(url) {
        return "<a href='#'>" + url + "</a>"
    }
}
