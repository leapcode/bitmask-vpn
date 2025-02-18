pragma ComponentBehavior: Bound

import QtQuick
import QtQuick.Controls
import QtQuick.Controls.Material

Item {
    id: providerSelection

    Rectangle {
        id: pageHeader
        color: "transparent"
        anchors.top: parent.top
        height: welcome.height + selectProviderLabel.height + providerTrustMsg.height + 5
        width: parent.width
        anchors.horizontalCenter: parent.horizontalCenter

        Label {
            id: welcome
            text: qsTr("Welcome!")
            font.bold: true
            font.pixelSize: 14
            anchors.top: parent.top
            leftPadding: 20
            rightPadding: 20
        }

        Label {
            id: selectProviderLabel
            text: qsTr("Select Your Provider")
            font.bold: true
            font.pixelSize: 14
            anchors.top: welcome.bottom
            leftPadding: 20
            rightPadding: 20
        }

        Label {
            id: providerTrustMsg
            text: qsTr("When using a VPN you are transferring your trust from your Internet Service Provider to your VPN provider. Bitmask only connects to providers with a clear history of privacy protection and advocacy")
            font.pixelSize: 12
            wrapMode: Text.WordWrap
            anchors.top: selectProviderLabel.bottom
            anchors.topMargin: 10
            width: parent.width
            leftPadding: 20
            rightPadding: 20
        }
    }

    Rectangle {
        id: selectProvider
        color: "white"
        radius: 8
        Material.elevation: 20
        height: {
            return 90;
        }
        anchors.top: pageHeader.bottom
        anchors.topMargin: 10
        anchors.horizontalCenter: parent.horizontalCenter
        width: parent.width - 14

        ButtonGroup {
            id: providerSel
        }

        ScrollView {
            id: frame
            clip: true
            ScrollBar.horizontal.policy: ScrollBar.AlwaysOff
            leftPadding: 15
            height: 90
            width: parent.width

            ListView {
                id: providersView
                model: root.providersModel
                delegate: MaterialRadioButton {
                    required property string providerName

                    text: providerName
                    ButtonGroup.group: providerSel
                    checked: providerName === root.ctx.provider
                    HoverHandler {
                        cursorShape: Qt.PointingHandCursor
                    }
                    onClicked: function () {
                        if (providerName === "Add new provider") {
                            addNewProviderInputBox.visible = true;
                            addProviderViaInviteCodeBox.visible = false;
                            providerSetupPage.providerName = "";
                        } else if (providerName == "Enter invite Code") {
                            addProviderViaInviteCodeBox.visible = true;
                            addNewProviderInputBox.visible = false;
                            providerSetupPage.providerName = "";
                        } else {
                            addNewProviderInputBox.visible = false;
                            addProviderViaInviteCodeBox.visible = false;

                            console.log("Provider name: ", providerName);
                            providerSetupPage.providerName = providerName;
                        }
                    }
                }
            }
        }
    }

    Rectangle {
        id: addNewProviderInputBox
        anchors.top: selectProvider.bottom
        anchors.topMargin: 15
        height: trustedProviderMsg.implicitHeight + providerSyntaxCheck.implicitHeight + 55
        width: parent.width - 14
        anchors.horizontalCenter: parent.horizontalCenter
        color: "ghostwhite"
        visible: false

        Label {
            id: trustedProviderMsg
            text: qsTr("Bitmask connects to trusted providers that are not publicly listed. Enter your provider's url below.")
            font.pixelSize: 12
            wrapMode: Text.WordWrap
            anchors.top: parent.top
            width: parent.width
            topPadding: 10
            leftPadding: 10
            rightPadding: 10
        }

        TextField {
            id: addProviderInput
            placeholderText: qsTr("Enter the provider's URL here:")

            font.pixelSize: 12
            wrapMode: Text.WordWrap
            height: 35
            width: parent.width - 14
            anchors.top: trustedProviderMsg.bottom
            anchors.left: parent.left
            anchors.topMargin: 10
            anchors.leftMargin: 10
            anchors.rightMargin: 10
            onTextChanged: function () {
                var input = text;
                var pattern = /^(http:\/\/www\.|https:\/\/www\.|http:\/\/|https:\/\/)?[a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?$/;
                console.log("User input:", input);
                if (pattern.test(input)) {
                    providerSetupPage.providerName = input;
                    providerSyntaxCheckResult.text = "Good";
                    providerSyntaxCheckResult.color = "green";
                } else {
                    providerSyntaxCheckResult.text = "Bad";
                    providerSyntaxCheckResult.color = "red";
                    providerSetupPage.providerName = "";
                }
            }
        }

        Label {
            id: providerSyntaxCheck
            text: qsTr("Syntax check:")
            font.pixelSize: 12
            wrapMode: Text.WordWrap
            anchors.bottom: parent.bottom
            anchors.left: parent.left
            bottomPadding: 10
            leftPadding: 10
            rightPadding: 10
        }

        Label {
            id: providerSyntaxCheckResult
            font.pixelSize: 12
            wrapMode: Text.WordWrap
            anchors.bottom: parent.bottom
            anchors.left: providerSyntaxCheck.right
            bottomPadding: 10
            text: ""
        }
    }

    Rectangle {
        id: addProviderViaInviteCodeBox
        anchors.top: selectProvider.bottom
        anchors.topMargin: 15
        height: trustedProviderMsg.implicitHeight + providerSyntaxCheck.implicitHeight + 55
        width: parent.width - 14
        anchors.horizontalCenter: parent.horizontalCenter
        visible: false
        color: "ghostwhite"

        Label {
            id: trustedInviteCodeMsg
            text: qsTr("Bitmask allows you to connect to providers using a private Invite Code.")
            font.pixelSize: 12
            wrapMode: Text.WordWrap
            anchors.top: parent.top
            width: parent.width
            topPadding: 10
            leftPadding: 10
            rightPadding: 10
        }

        TextField {
            id: addProviderViaInviteCodeInput
            placeholderText: qsTr("Enter your trusted Invite Code here:")

            font.pixelSize: 12
            wrapMode: Text.WordWrap
            height: 35
            width: parent.width - 14
            anchors.top: trustedInviteCodeMsg.bottom
            anchors.left: parent.left
            anchors.topMargin: 10
            anchors.leftMargin: 10
            anchors.rightMargin: 10
            onTextChanged: function () {
                var input = text;
                var pattern = /^obfsvpnintro:\/\//;
                console.log("User input:", input);
                if (pattern.test(input)) {
                    providerSetupPage.providerName = input;
                    inviteCodeSyntaxCheckResult.text = "Good";
                    inviteCodeSyntaxCheckResult.color = "green";
                } else {
                    inviteCodeSyntaxCheckResult.text = "Bad";
                    inviteCodeSyntaxCheckResult.color = "red";
                    providerSetupPage.providerName = "";
                }
            }
        }

        Label {
            id: inviteCodeSyntaxCheck
            text: qsTr("Syntax check:")
            font.pixelSize: 12
            wrapMode: Text.WordWrap
            anchors.bottom: parent.bottom
            anchors.left: parent.left
            bottomPadding: 10
            leftPadding: 10
            rightPadding: 10
        }

        Label {
            id: inviteCodeSyntaxCheckResult
            font.pixelSize: 12
            wrapMode: Text.WordWrap
            anchors.bottom: parent.bottom
            anchors.left: inviteCodeSyntaxCheck.right
            bottomPadding: 10
            text: ""
        }
    }
}
