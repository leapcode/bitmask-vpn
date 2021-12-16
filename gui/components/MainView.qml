import QtQuick 2.0
import QtQuick.Controls 2.4
import QtQuick.Dialogs 1.2
import QtQuick.Controls.Material 2.1
import QtQuick.Layouts 1.14

import "../themes/themes.js" as Theme

Page {
    StackView {
        id: stackView
        anchors.fill: parent
        initialItem: Home {}
    }

    NavigationDrawer {
        id: settingsDrawer
        Rectangle {
            anchors.fill: parent
            color: "white"
            ListView {
                focus: true
                currentIndex: -1
                anchors.fill: parent

                model: navModel
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
                        settingsDrawer.toggle()
                        model.triggered()
                    }
                }
            }
        }
     }

     ListModel {
        id: navModel
        ListElement {
            text: qsTr("Preferences")
            icon: "../resources/tools.svg"
            triggered: function() {
                stackView.push("Preferences.qml")
            }
        }
        ListElement {
            text: qsTr("Donate")
            icon: "../resources/donate.svg"
            triggered: function() {
                Qt.openUrlExternally(ctx.donateURL)
            }
        }
        ListElement {
            text: qsTr("Help")
            icon: "../resources/help.svg"
            triggered: function() {
                stackView.push("Help.qml")
            }
        } // -> can link to another dialog with report bug / support / contribute / FAQ
        ListElement {
            text: qsTr("About")
            icon: "../resources/about.svg"
            triggered: function() {
                stackView.push("About.qml")
            }
        }
        ListElement {
            text: qsTr("Quit")
            icon: "../resources/quit.svg"
            triggered: function() {
                if (ctx.status == "on") {
                    backend.switchOff()
                }
                Qt.callLater(backend.quit)
            }
        }
    } // end listmodel

    header: Header {
        id: header
    }
    footer: Footer {
        id: footer
    }

    Keys.onPressed: {
        // shortcuts for avid users :)
        // bug: doesnt work until the stack is pushed once
        if (event.key == Qt.Key_G && stackView.depth == 1) {
            console.debug("Open Locations")
            stackView.push("Locations.qml")
        }
    }

    Dialog {
        id: donateDialog 
        width: root.appWidth
        title: qsTr("Please donate!")
        standardButtons: Dialog.Yes | Dialog.No

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
            text: qsTr("This service is paid for entirely by donations from users like you. The cost of running the VPN is approximately 5 USD per person every month, but every little bit counts. Do you want to donate now?")
        }

        Label {
            id: donateURL
            anchors {
                top: donateText.bottom
                topMargin: 20
                horizontalCenter: parent.horizontalCenter
            }
            font.pixelSize: 14
            textFormat: Text.RichText
            text: getLink(ctx)
            onLinkActivated: Qt.openUrlExternally(ctx.donateURL)
        }


        Image {
            height: 40
            source: "../resources/donate.svg"
            fillMode: Image.PreserveAspectFit
            anchors {
                topMargin: 20
                top: donateURL.bottom
                bottomMargin: 50
                horizontalCenter: parent.horizontalCenter
            }
        }
        onYes: Qt.openUrlExternally(ctx.donateURL)
    }

    function getLink(ctx) {
	if (!ctx) {
		return ""
	}
	let url = ctx.donateURL
        return "<style>a:link {color:'" + Theme.blue + "'; }</style><a href='#'>" + url + "</a>"
    }

    Component.onCompleted: {
        root.openDonateDialog.connect(donateDialog.open)
    }
}
