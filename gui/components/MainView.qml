import QtQuick
import QtQuick.Controls
import QtQuick.Dialogs
import QtQuick.Controls.Material
import QtQuick.Layouts
import "../themes/themes.js" as Theme

Page {
    id: mainView
    StackView {
        id: stackView
        anchors.fill: parent
        initialItem: Home {
            id: home
        }
    }

    Drawer {
        id: settingsDrawer
        width: parent.width * 0.65
        height: parent.height
        background: Rectangle {
            Rectangle {
                x: parent.width
                width: 1
                height: parent.height
                color: "papayawhip"
                radius: 0
            }
        }

        ListView {
            focus: true
            currentIndex: -1
            anchors.fill: parent

            model: navItems
            delegate: ItemDelegate {
                width: parent.width
                text: model.modelData.text
                highlighted: ListView.isCurrentItem
                icon.color: "transparent"
                icon.source: model.modelData.icon
                onClicked: {
                    settingsDrawer.close();
                    model.modelData.triggered();
                }
            }
        }
    }

    property bool showSwitchProvider: (ctx && ctx.providers && ctx.providers.length > 1)
                                     || (ctx && ctx.appName === "Bitmask")

    property var navItems: {
        var items = [
            { text: qsTr("Preferences"), icon: "../resources/tools.svg",
              triggered: function () { stackView.push("Preferences.qml"); } }
        ];
        if (isDonationService) {
            items.push({ text: qsTr("Donate"), icon: "../resources/donate.svg",
              triggered: function () { Qt.openUrlExternally(ctx.donateURL); } });
        }
        items.push(
            { text: qsTr("Help"), icon: "../resources/help.svg",
              triggered: function () { stackView.push("Help.qml"); } },
            { text: qsTr("About"), icon: "../resources/about.svg",
              triggered: function () { stackView.push("About.qml"); } },
            { text: qsTr("Quit"), icon: "../resources/quit.svg",
              triggered: function () { Qt.callLater(backend.quit); } }
        );
        if (showSwitchProvider) {
            items.push({ text: qsTr("Switch Provider"), icon: "../resources/switch_provider.svg",
              triggered: function () {
                  stackView.push("SwitchProvider.qml",
                      { "isSwitchProvider": true, "parentStackView": stackView });
              } });
        }
        return items;
    }

    header: Header {
        id: header
        visible: {
            if (stackView.currentItem && stackView.currentItem.isSwitchProvider) {
                return false;
            }
            return stackView.depth > 1 || forceInitialVisibility;
        }
    }

    Keys.onPressed: {
        // shortcuts for avid users :)
        // bug: doesnt work until the stack is pushed once
        if (event.key == Qt.Key_G && stackView.depth == 1) {
            console.debug("Open Locations");
            stackView.push("Locations.qml");
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
        onAccepted: Qt.openUrlExternally(ctx.donateURL)
    }

    function getLink(ctx) {
        if (!ctx) {
            return "";
        }
        let url = ctx.donateURL;
        return "<style>a:link {color:'" + Theme.blue + "'; }</style><a href='#'>" + url + "</a>";
    }

    function loadMainView() {
        stackView.pop();
    }

    function setStatusStarting() {
        home.setStatusBoxStateStarting();
    }

    Component.onCompleted: {
        root.openDonateDialog.connect(donateDialog.open);
    }
}
