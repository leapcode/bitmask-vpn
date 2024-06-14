import QtQuick
import QtQuick.Controls
import QtQuick.Controls.Material
import QtQuick.Layouts
import QtQuick.Effects
import "../themes/themes.js" as Theme

ToolBar {
    Material.foreground: "black"
    Material.elevation: 10
    visible: isFooterVisible()
    background: Rectangle {
        implicitHeight: 48
        color: "transparent"

        Rectangle {
            width: parent.width
            height: 1
            anchors.bottom: parent.bottom
            color: "transparent"
        }
    }

    Rectangle {
        id: footerRow
        width: root.width - 18
        height: 48
        radius: 8
        color: "white"
        opacity: 0.9

        ToolButton {
            id: gwButton
            visible: true

            anchors {
                verticalCenter: parent.verticalCenter
                leftMargin: 10
                left: parent.left
            }
            icon {
                width: 20
                height: 20
                source: stackView.depth > 1 ? "" : "../resources/globe.svg"
            }
            HoverHandler {
                cursorShape: Qt.PointingHandCursor
            }
            onClicked: stackView.push("Locations.qml")
        }

        Image {
            id: lightning
            smooth: true
            visible: ctx != undefined & root.selectedGateway == "auto"
            width: 16
            source: "../resources/lightning.svg"
            fillMode: Image.PreserveAspectFit
            anchors {
                left: gwButton.right
                leftMargin: -10
                verticalCenter: gwButton.verticalCenter
            }
        }
        MultiEffect {
            anchors.fill: lightning
            source: lightning
            colorizationColor: getLocationColor()
            colorization: 1.0
            antialiasing: true
        }

        Label {
            id: locationLabel
            text: locationStr()
            color: getLocationColor()
            anchors {
                left: lightning.right
                verticalCenter: gwButton.verticalCenter
            }
            MouseArea {
                cursorShape: Qt.PointingHandCursor
                anchors.fill: parent
                onClicked: stackView.push("Locations.qml")
            }
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
                verticalCenterOffset: -2
                right: gwQuality.left
                rightMargin: 10
            }
        }

        // TODO refactor with SignalIcon
        // This signal image renders particularly bad at this size.
        // https://stackoverflow.com/a/23449205/1157664
        Image {
            id: gwQuality
            source: "../resources/reception-0@24.svg"
            width: 24
            sourceSize.width: 24
            smooth: false
            mipmap: true
            antialiasing: false
            anchors {
                right: parent.right
                verticalCenter: parent.verticalCenter
                verticalCenterOffset: -5
                topMargin: 5
                rightMargin: 20
            }
        }
        MultiEffect {
            anchors.fill: gwQuality
            source: gwQuality
            colorizationColor: getSignalColor()
            colorization: 1.0
            antialiasing: false
        }
    }

    function getSignalColor() {
        if (ctx && ctx.status == "on") {
            return "green";
        } else {
            return "black";
        }
    }

    StateGroup {
        state: ctx ? ctx.status : "off"
        states: [
            State {
                name: "on"
                PropertyChanges {
                    target: gwQuality
                    source: "../resources/reception-4@24.svg"
                }
            },
            State {
                name: "off"
                PropertyChanges {
                    target: gwQuality
                    source: "../resources/reception-0@24.svg"
                }
            }
        ]
    }

    function locationStr() {
        if (ctx && ctx.status == "on") {
            if (ctx.currentLocation && ctx.currentCountry) {
                let s = ctx.currentLocation + ", " + ctx.currentCountry;
                /*
                if (root.selectedGateway == "auto") {
                    s = "🗲 " + s
                }
                */
                return s;
            }
        }
        if (root.selectedGateway == "auto") {
            if (ctx && ctx.locations && ctx.bestLocation) {
                //return "🗲 " + getCanonicalLocation(ctx.bestLocation)
                return getCanonicalLocation(ctx.bestLocation);
            } else {
                return qsTr("Recommended");
            }
        }
        if (ctx && ctx.locations && ctx.locationLabels) {
            return getCanonicalLocation(root.selectedGateway);
        }
    }

    // returns the composite of Location, CC
    function getCanonicalLocation(label) {
        try {
            let loc = ctx.locationLabels[label];
            return loc[0] + ", " + loc[1];
        } catch (e) {
            return "unknown";
        }
    }

    function getLocationColor() {
        if (ctx && ctx.status == "on") {
            return "black";
        } else {
            // TODO darker gray
            return "gray";
        }
    }

    function hasMultipleGateways() {
        let provider = getSelectedProvider(providers);
        if (provider == "riseup") {
            return true;
        } else {
            if (!ctx) {
                return false;
            }
            return ctx.locations.length > 0;
        }
    }

    function getSelectedProvider(providers) {
        let obj = JSON.parse(providers.getJson());
        return obj['default'];
    }

    function isBridgeSelected() {
        if (ctx && ctx.transport == "obfs4") {
            return true;
        } else {
            return false;
        }
    }

    function isFooterVisible() {
        if (drawerOn) {
            return false;
        }
        if (stackView.depth > 1) {
            return false;
        }
        return true;
    }
}
