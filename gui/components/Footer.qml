import QtQuick 2.0
import QtQuick.Controls 2.4
import QtQuick.Controls.Material 2.1
import QtQuick.Layouts 1.14
import QtGraphicalEffects 1.0

import "../themes/themes.js" as Theme

ToolBar {

    Material.background: Theme.bgColor
    Material.foreground: "black"
    Material.elevation: 0
    visible: stackView.depth > 1 && ctx !== undefined ? false : true

    Item {

        id: footerRow
        width: root.width

        ToolButton {
            id: gwButton
            visible: hasMultipleGateways()

            anchors {
                verticalCenter: parent.verticalCenter
                leftMargin: 10
                // TODO discuss where this should be aligned
                //leftMargin: 22
                left: parent.left
                verticalCenterOffset: 5
            }
            /*
            background.implicitHeight: 32
            background.implicitWidth: 32
            */
            icon {
                width: 20
                height: 20
                source: stackView.depth > 1 ? "" : "../resources/globe.svg"
            }
            onClicked: stackView.push("Locations.qml")
        }

        Label {
            id: locationLabel
            anchors {
                left: gwButton.right
                leftMargin: -7
                verticalCenter: parent.verticalCenter
                verticalCenterOffset: 7
            }
            text: locationStr()
            color: getLocationColor()
        }

        Item {
            Layout.fillWidth: true
            height: gwButton.implicitHeight
        }

        Image {
            id: bridge
            smooth: true
            visible: isBridgeSelected()
            width: 40
            source: "../resources/bridge.svg"
            fillMode: Image.PreserveAspectFit
            anchors {
                verticalCenter: parent.verticalCenter
                verticalCenterOffset: 5
                right: gwQuality.left
                rightMargin: 10
            }
        }

        Image {
            id: gwQuality
            height: 24
            width: 24
            source: "../resources/reception-0.svg"
            anchors.right: parent.right
            anchors.rightMargin: 20
            anchors.verticalCenter: parent.verticalCenter
            anchors.verticalCenterOffset: 0
            // TODO refactor with SignalIcon
            ColorOverlay{
                anchors.fill: gwQuality
                source: gwQuality
                color: getSignalColor()
                antialiasing: true
            }
        }
    }

    function getSignalColor() {
        if (ctx && ctx.status == "on") {
            return "green"
        } else {
            return "black"
        }
    }

    StateGroup {
        state: ctx ? ctx.status : "off"
        states: [
            State {
                name: "on"
                PropertyChanges {
                    target: gwQuality
                    source: "../resources/reception-4.svg"
                }
            },
            State {
                name: "off"
                PropertyChanges {
                    target: gwQuality
                    source: "../resources/reception-0.svg"
                }
            }
        ]
    }

    function locationStr() {
        if (ctx && ctx.status == "on") {
            if (ctx.currentLocation && ctx.currentCountry) {
                let s = ctx.currentLocation + ", " + ctx.currentCountry
                if (root.selectedGateway == "auto") {
                    s = "🗲 " + s
                }
                return s
            }
        }
        if (root.selectedGateway == "auto") {
            if (ctx && ctx.locations && ctx.bestLocation) {
                return "🗲 " + getCanonicalLocation(ctx.bestLocation)
            } else {
                return qsTr("Recommended")
            }
        }
        if (ctx && ctx.locations && ctx.locationLabels) {
            return getCanonicalLocation(root.selectedGateway)
        }
    }

    // returns the composite of Location, CC
    function getCanonicalLocation(label) {
        try {
            let loc = ctx.locationLabels[label]
            return loc[0] + ", " + loc[1]
        } catch(e) {
            return "unknown"
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

    function hasMultipleGateways() {
        let provider = getSelectedProvider(providers)
        if (provider == "riseup") {
            return true
        } else {
            if (!ctx) {
                return false
            }
            return ctx.locations.length > 0
        }
    }

    function getSelectedProvider(providers) {
        let obj = JSON.parse(providers.getJson())
        return obj['default']
    }

    function isBridgeSelected() {
        if (ctx && ctx.transport == "obfs4") {
            return true
        } else {
            return false
        }
    }
}
