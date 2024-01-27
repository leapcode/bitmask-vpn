import QtQuick 2.15
import QtQuick.Controls 2.2
import QtQuick.Layouts 1.14
import Qt5Compat.GraphicalEffects

import "../themes/themes.js" as Theme


/* TODO
 [ ] corner case: manual override, not full list yet
     [x] persist bridges
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
    property var switchingLocationLabel: qsTr("Switching gateway…")
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
            Rectangle {
                id: recommendedHeader
                height: 20
                Label {
                    id: recommendedLabel
                    //: Location Selection: label for radio button that selects automatically
                    text: qsTr("Recommended")
                    font.weight: Font.Bold
                    font.bold: true
                }
                Image {
                    id: lightning 
                    smooth: true
                    width: 16
                    source: "../resources/lightning.svg"
                    fillMode: Image.PreserveAspectFit
                    verticalAlignment: Image.AlignVCenter
                    anchors {
                        left: recommendedLabel.right
                        top: parent.top
                        leftMargin: 5
                        topMargin: 2
                        //verticalCenterOffset: 3
                    }
                }
                MultiEffect {
                    anchors.fill: lightning
                    source: lightning
                    colorizationColor: "black"
                    colorization: 1.0
                    antialiasing: true
                }
            }
            WrappedRadioButton {
                id: autoRadioButton
                text: getAutoLabel()
                ButtonGroup.group: locsel
                checked: false
                anchors {
                    top: recommendedHeader.bottom
                    leftMargin: -5
                }
                HoverHandler {
                    cursorShape: Qt.PointingHandCursor
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
                        font.pixelSize: Theme.fontSize - 3
                        anchors {
                            topMargin: 5
                            top: manualLabel.bottom
                        }
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
                                    HoverHandler {
                                        cursorShape: Qt.PointingHandCursor
                                    }
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
                when: ctx != undefined && ctx.status == "on"
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
        /* There's been some discussion about whether to include this.
         An argument is that it is not 100% sure that we're going to connect
         to this "recommended" gateway. However, it's fair to tell the user what's likely
         to be the recomended location, to make a better choice. ALso, we can
         implement a warning if finally connecting to a different location.
         That said, all is made worse by the fact that menshen will not return
         the "right" location if we're connecting  from the vpn, a proxy etc... For that we need to modify menshen to accept a location parameter.
         Disabling the hint for now, but some agreement needs to be done on android + desktop about this behavior.
        if (ctx && ctx.locations && ctx.bestLocation) {
            let best = ctx.locationLabels[ctx.bestLocation]
            let label = best[0] + ", " + best[1]
            l += " (" + label + ")"
        }
        */
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
        return h + 50
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

    function getLocationColor() {
        if (ctx && ctx.status == "on") {
            return "black"
        } else {
            // TODO darker gray
            return "gray"
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
