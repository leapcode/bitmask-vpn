import QtQuick 2.9
import QtQuick.Controls 2.2

import "../themes/themes.js" as Theme

Page {
    title: qsTr("Select Location")

    ListView {
        id: gwList
        focus: true
        currentIndex: -1
        anchors.fill: parent
        spacing: 1

        delegate: ItemDelegate {
            id: loc
            Rectangle {
                width: parent.width
                height: 1
                color: Theme.borderColor
            }
            width: parent.width
            text: model.text
            highlighted: ListView.isCurrentItem
            icon.color: "transparent"
            icon.source: model.icon
            onClicked: {
                model.triggered()
                stackView.pop()
            }
            MouseArea {
                property var onMouseAreaClicked: function () {
                    parent.clicked()
                }
                id: mouseArea
                anchors.fill: loc
                cursorShape: Qt.PointingHandCursor
                onReleased: {
                    onMouseAreaClicked()
                }
            }
        }

        model: ListModel {
            ListElement {
                text: qsTr("Paris")
                triggered: function () {}
                icon: "../resources/reception-4.svg"
            }
            ListElement {
                text: qsTr("Montreal")
                triggered: function () {}
                icon: "../resources/reception-4.svg"
            }
            ListElement {
                text: qsTr("Seattle")
                triggered: function () {}
                icon: "../resources/reception-2.svg"
            }
        }
    }
}
