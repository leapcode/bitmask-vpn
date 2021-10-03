import QtQuick 2.9
import QtQuick.Controls 2.2
import QtQuick.Layouts 1.14
import QtGraphicalEffects 1.0

import "../themes/themes.js" as Theme


/* TODO
 [ ] corner case: manual override, not full list yet
     [ ] persist bridges
     [ ] persist manual selection
     [ ] display the location we know
 [ ] corner case: user selects bridges with manual selection
     (I think the backend should discard any manual selection when selecting bridges...
      unless the current selection provides the bridge, in which case we can maintain it)
 */
ThemedPage {

    id: locationPage
    title: qsTr("Select Location")

    // TODO add ScrollIndicator
    // https://doc.qt.io/qt-5.12//qml-qtquick-controls2-scrollindicator.html

    //: this is in the radio button for the auto selection
    property var autoSelectionLabel: qsTr("Automatically use best connection")
    //: Location Selection: label for radio buttons that selects manually
    property var manualSelectionLabel: qsTr("Manually select")
    //: A little display to signal that the clicked gateway is being switched to
    property var switchingLocationLabel: qsTr("Switching gateways...")
    //: Subtitle to explain that only bridge locations are shown in the selector
    property var onlyBridgesWarning: qsTr("Only locations with bridges")

    property bool switching: false

    ButtonGroup {
        id: locsel
    }

    Rectangle {
        id: autoBox
        width: root.width * 0.90
        height: 90
        radius: 10
        color: "white"

        anchors {
            horizontalCenter: parent.horizontalCenter
            top: parent.top
            margins: 10
        }

        Rectangle {
            anchors {
                fill: parent
                margins: 10
            }
            Label {
                id: recommendedLabel
                //: Location Selection: label for radio button that selects automatically
                text: qsTr("Recommended")
                font.bold: true
            }
            WrappedRadioButton {
                id: autoRadioButton
                text: getAutoLabel()
                ButtonGroup.group: locsel
                checked: false
                anchors {
                    top: recommendedLabel.bottom
                    leftMargin: -5
                }
                onClicked: {
                    root.selectedGateway = "auto"
                    console.debug("Selected gateway: auto")
                    backend.useAutomaticGateway()
                }
            }
        }
    }

    Rectangle {
        id: manualBox
        visible: root.locationsModel.length > 0
        width: root.width * 0.90
        radius: 10
        color: Theme.fgColor
        height: getManualBoxHeight()

        anchors {
            horizontalCenter: parent.horizontalCenter
            top: autoBox.bottom
            margins: 10
        }

        ScrollView {
            id: frame
            clip: true
            anchors.fill: parent
            ScrollBar.vertical.policy: ScrollBar.AlwaysOff

            Flickable {
                id: flickable
                contentHeight: getManualBoxHeight()
                width: parent.width

                ScrollIndicator.vertical: ScrollIndicator {
                    size: 5
                    contentItem: Rectangle {
                        implicitWidth: 5
                        implicitHeight: 100
                        color: "grey"
                    }
                }

                Rectangle {
                    anchors {
                        fill: parent
                        margins: 10
                    }
                    Label {
                        id: manualLabel
                        text: manualSelectionLabel
                        font.bold: true
                    }
                    Label {
                        id: bridgeWarning
                        text: onlyBridgesWarning
                        color: "gray"
                        visible: isBridgeSelected()
                        wrapMode: Text.Wrap
                        anchors {
                            topMargin: 5
                            top: manualLabel.bottom
                        }
                        font.pixelSize: Theme.fontSize - 3
                    }

                    ColumnLayout {
                        id: gatewayListColumn
                        width: parent.width
                        spacing: 1
                        anchors {
                            topMargin: 10
                            top: getManualAnchor()
                        }

                        Repeater {
                            id: gwManualSelectorList
                            width: parent.width
                            model: root.locationsModel

                            RowLayout {
                                width: parent.width
                                WrappedRadioButton {
                                    text: getLocationLabel(modelData)
                                    location: modelData
                                    ButtonGroup.group: locsel
                                    checked: false
                                    enabled: locationPage.switching ? false : true
                                    onClicked: {
                                        if (ctx.status == "on") {
                                            locationPage.switching = true
                                        }
                                        root.selectedGateway = location
                                        backend.useLocation(location)
                                    }
                                }
                                Item {
                                    Layout.fillWidth: true
                                }
                                Image {
                                    height: 30
                                    width: 30
                                    smooth: true
                                    visible: isBridgeSelected()
                                    fillMode: Image.PreserveAspectFit
                                    source: "../resources/bridge.svg"
                                    Layout.alignment: Qt.AlignRight
                                    Layout.rightMargin: 10
                                }
                                SignalIcon {
                                    quality: getSignalFor(modelData)
                                    Layout.alignment: Qt.AlignRight
                                    Layout.rightMargin: 20
                                }
                            }
                        }
                    }
                }
            } //flickable
        } // scrollview
    } // manualbox

    StateGroup {
        states: [
            State {
                when: locationPage.switching && ctx.status != "on"
                PropertyChanges {
                    target: manualLabel
                    text: switchingLocationLabel
                }
            },
            State {
                when: ctx && ctx.status == "on"
                PropertyChanges {
                    target: manualLabel
                    text: manualSelectionLabel
                }
                StateChangeScript {
                    script: {
                        locationPage.switching = false
                    }
                }
            }
        ]
    }

    function getAutoLabel() {
        let l = autoSelectionLabel
        if (ctx && ctx.locations && ctx.bestLocation) {
            let best = ctx.locationLabels[ctx.bestLocation]
            let label = best[0] + ", " + best[1]
            l += " (" + label + ")"
        }
        return l
    }

    function getLocationLabel(location) {
        if (!ctx) {
            return ""
        }
        let l = ctx.locationLabels[location]
        return l[0] + ", " + l[1]
    }

    function getManualBoxHeight() {
        let h = Math.min(
            root.locationsModel.length * 35,
            root.appHeight - autoBox.height - 100
        )
        if (bridgeWarning.visible) {
            h += bridgeWarning.height
        }
        return h + 30
    }

    function getSignalFor(location) {
        // this is an ad-hoc solution for the no-menshen, riseup case.
        // when menshen is deployed we'll want to tweak the values for each bucket.
        let load = ctx.locations[location]
        switch (true) {
        case (load > 0.5):
            return "good"
        case (load > 0.25):
            return "medium"
        default:
            return "low"
        }
    }

    function isBridgeSelected() {
        if (ctx && ctx.transport == "obfs4") {
            return true
        } else {
            return false
        }
    }

    function getManualAnchor() {
        if (isBridgeSelected()) {
            return bridgeWarning.bottom
        } else {
            return manualLabel.bottom
        }
    }

    Component.onCompleted: {
        if (root.selectedGateway == "auto") {
            autoRadioButton.checked = true
        } else {
            let match = false
            for (var i = 1; i < locsel.buttons.length; i++) {
                let b = locsel.buttons[i]
                if (b.location == root.selectedGateway) {
                    match = true
                    b.checked = true
                }
            }
        }
    }
}
