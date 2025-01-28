pragma ComponentBehavior: Bound

import QtQuick
import QtQuick.Controls
import QtQuick.Controls.Material
import QtQuick.Layouts

Item {
    id: selectCircumvention

    Rectangle {
        id: pageHeader
        color: "transparent"
        anchors.top: parent.top
        height: needCircumventionLabel.height + needCircumventionMsg.height + 5
        width: parent.width

        Label {
            id: needCircumventionLabel
            text: qsTr("Do You Require Censorship Circumvention?")
            font.bold: true
            font.pixelSize: 14
            anchors.top: parent.top
            wrapMode: Text.WordWrap
            leftPadding: 20
            rightPadding: 20
            width: parent.width
            clip: false
        }

        Label {
            id: needCircumventionMsg
            text: qsTr("If you live where the internet is censored you can use our censorship circumvention options to access all internet services. These options will slow down your connection!")
            font.pixelSize: 12
            wrapMode: Text.WordWrap
            anchors.top: needCircumventionLabel.bottom
            anchors.topMargin: 10

            width: parent.width
            leftPadding: 20
            rightPadding: 20
            clip: false
        }
    }

    Rectangle {
        id: selectCircumventionTechnique
        color: "white"
        radius: 10
        Material.elevation: 20
        height: getManualContainerHeight()
        width: parent.width - 15
        anchors.top: pageHeader.bottom
        anchors.topMargin: 20
        anchors.horizontalCenter: parent.horizontalCenter

        ListModel {
            id: circumventionOptionsModel

            ListElement {
                name: "Use standard Bitmask"
                description: ""
            }
            ListElement {
                name: "Use circumvention tech (slower)"
                description: "Bitmask will automatically try to connect you to the internet using a variety of circumvention technologies. You can fine tune this in the advanced settings."
            }
        }

        ButtonGroup {
            id: circumventionOptionSel
        }

        ScrollView {
            id: frame
            clip: true
            ScrollBar.horizontal.policy: ScrollBar.AlwaysOff
            leftPadding: 15
            width: parent.width

            ListView {
                id: circumventionOptionsView
                model: circumventionOptionsModel
                delegate: MaterialRadioButton {
                    required property string name
                    required property string description

                    text: name
                    checked: name === "Use standard Bitmask"
                    ButtonGroup.group: circumventionOptionSel
                    HoverHandler {
                        cursorShape: Qt.PointingHandCursor
                    }
                    onClicked: function () {
                        if (name === "Use standard Bitmask") {
                            circumventionDescriptionLabel.visible = false;
                            providerSetupPage.useCircumvention = false;
                        } else {
                            circumventionDescriptionLabel.visible = true;
                            circumventionDescriptionLabel.text = qsTr(description);
                            providerSetupPage.useCircumvention = true;
                        }
                        console.log("circumvention name: ", name);
                    }
                }
            }
        }

        Label {
            id: circumventionDescriptionLabel
            font.pixelSize: 12
            anchors.top: frame.bottom
            visible: false
            wrapMode: Text.WordWrap
            width: parent.width
            leftPadding: 15
            rightPadding: 15
            anchors.topMargin: 10
            color: "gray"
        }
    }

    function getManualContainerHeight() {
        let h = Math.min(70, root.appHeight - needCircumventionLabel.height - needCircumventionMsg.height - 100);
        if (circumventionDescriptionLabel.visible) {
            h += circumventionDescriptionLabel.height + 20;
        }
        return h;
    }
}
