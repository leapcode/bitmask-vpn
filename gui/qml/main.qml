import QtQuick 2.9
import QtQuick.Dialogs 1.2 // TODO use native dialogs in labs.platform
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.4

import Qt.labs.platform 1.0

import "logic.js" as Logic

ApplicationWindow {
    id: app
    visible: false
    width: 300
    height: 600
    maximumWidth: 300
    minimumWidth: 300
    maximumHeight: 500
    minimumHeight: 300

    property var ctx
    property var loginDone
    property var allowEmptyPass
    property var needsRestart
    property var needsDonate
    property var shownDonate

    onSceneGraphError: function(error, msg) {
        console.debug("ERROR while initializing scene")
        console.debug(msg)
    }

    MainBar {
        id: bar
    }

    StackLayout {

        anchors.fill: parent
        currentIndex: bar.currentIndex

        Item {

            id: infoTab
            anchors.centerIn: parent

            BackgroundImage {
                id: background
            }

            Item {
                id: connBox
                anchors.centerIn: parent

                width: 300
                height: 300

                Rectangle {
                    anchors.fill: parent
                    color: "white"
                    opacity: 0.3
                    layer.enabled: true
                }

                Column {

                    anchors.centerIn: parent
                    spacing: 5

                    Text {
                        id: mainStatus
                        text: "off"
                        font.pixelSize: 26
                        anchors.horizontalCenter: parent.horizontalCenter
                    }

                    Text {
                        id: mainCurrentGateway
                        text: ""
                        font.pixelSize: 20
                        anchors.horizontalCenter: parent.horizontalCenter
                    }

                    VPNSwitch {
                        id: vpntoggle

                        Connections {
                            function onCheckedChanged() {
                                if (vpntoggle.checked == true
                                        && ctx.status == "off") {
                                    backend.switchOn()
                                }
                                if (vpntoggle.checked === false
                                        && ctx.status == "on") {
                                    backend.switchOff()
                                }
                            }
                        }
                    }

                    LocationText {
                        id: manualOverrideWarning
                        visible: isManualLocation()
                    }
                } // end column
            } // end inner item
        } // end outer item

        Item {

            id: gatewayTab
            anchors.centerIn: parent

            Column {

                anchors.centerIn: parent
                spacing: 10

                RadioButton {
                    id: autoSelectionButton
                    checked: !isManualLocation()
                    text: qsTr("Recommended")
                    onClicked: {
                        backend.useAutomaticGateway()
                        manualSelectionItem.checked = false
                    }
                }
                RadioButton {
                    id: manualSelectionButton
                    checked: isManualLocation()
                    text: qsTr("Manual")
                    onClicked: setGwSelection()
                }
                ComboBox {
                    id: gwSelector
                    editable: false
                    visible: manualSelectionButton.checked
                    anchors.horizontalCenter: parent.horizontalCenter

                    model: [qsTr("Recommended")]
                    onActivated: {
                        console.debug("Selected gateway:", currentText)
                        backend.useLocation(currentText.toString())
                        manualSelectionItem.checked = true
                    }

                    delegate: ItemDelegate {
                        // TODO: we could use icons
                        // https://doc.qt.io/qt-5/qml-qtquick-controls2-abstractbutton.html#icon-prop
                        background: Rectangle {
                            color: {
                                "#ffffff"
                                // FIXME locations is not defined when we launch
                                /*
                                const fullness = ctx.locations[modelData]
                                if (fullness >= 0 && fullness < 0.4) {
                                    "#83fc5a"
                                } else if (fullness >= 0.4 && fullness < 0.75) {
                                    "#fcb149"
                                } else if (fullness >= 0.75) {
                                    "#fc5a5d"
                                } else {
                                    "#ffffff"
                                }
                                */
                            }
                        }
                        contentItem: Text {
                            text: modelData
                            font: gwSelector.font
                            color: "#000000"
                        }
                    }
                }
            } // end column
        } // end item

        BridgesItem {
            id: bridgesTab
        }

    } // end stacklayout

    Connections {
        target: jsonModel
        function onDataChanged() {
            ctx = JSON.parse(jsonModel.getJson())
            // TODO pass QML_DEBUG variable to be hyper-verbose
            //console.debug(jsonModel.getJson())

            gwSelector.model = Object.keys(ctx.locations)

            if (ctx.donateDialog == 'true') {
                Logic.setNeedsDonate(true);
            }
            if (ctx.loginDialog == 'true') {
                console.debug(jsonModel.getJson())
                console.debug("DEBUG: should display login")
                login.visible = true
            }
            if (ctx.loginOk == 'true') {
                loginOk.visible = true
            }
            if (ctx.errors) {
                login.visible = false
                if (ctx.errors == "nohelpers") {
                    showInitFailure(
                                qsTr("Could not find helpers. Please check your installation"))
                } else if (ctx.errors == "nopolkit") {
                    showInitFailure(qsTr("Could not find polkit agent."))
                } else {
                    showInitFailure()
                }
            }
            if (ctx.donateURL) {
                donateItem.visible = true
            }
        }
    }

    function showInitFailure(msg) {
        console.debug("ERRORS:", ctx.errors)
        if (msg == undefined) {
            if (ctx.errors == 'bad_auth_502'
                    || ctx.errors == 'bad_auth_timeout') {
                msg = qsTr("Oops! The authentication service seems down. Please try again later")
                initFailure.title = qsTr("Service Error")
            } else if (ctx.errors == 'bad_auth') {
                if (allowEmptyPass) {
                    // For now, this is a libraryVPN, so we can be explicit about what credentials are here.
                    // Another option to consider is to customize the error strings while vendoring.
                    //: Incorrect library card number
                    msg = qsTr("Please check your Patron ID")
                } else {
                    msg = qsTr("Could not log in with those credentials, please retry")
                }
                initFailure.title = qsTr("Login Error")
            } else {
                //: %1 -> application name
                //: %2 -> error string
                msg = qsTr("Got an error starting %1: %2").arg(ctx.appName).arg(
                            ctx.errors)
            }
        }
        initFailure.text = msg
        initFailure.visible = true
    }

    function isManualLocation() {
        if (!ctx) {
            return false
        }
        return ctx.manualLocation == "true"
    }

    function setGwSelection() {

        if (!isManualLocation()) {
            manualSelectionItem.checked = false
            bar.currentIndex = 1
            app.visible = true
            app.show()
            app.raise()
            return
        }

        // last used manual selection
        const location = ctx.currentLocation.toLowerCase()
        const idx = gwSelector.model.indexOf(location)
        gwSelector.currentIndex = idx
        backend.useLocation(location)
    }

    Component.onCompleted: {
        Logic.debugInit()
        loginDone = false
        allowEmptyPass = Logic.shouldAllowEmptyPass(providers)
        needsRestart = false;
        shownDonate = false;

        /* this is a temporary workaround until general GUI revamp for 0.21.8 */
        let provider = Logic.getSelectedProvider(providers);
        if (provider == "calyx") {
            background.color = "#8EA844";
            background.backgroundVisible = false;
            gwSelector.visible = false;
            manualSelectionButton.visible = false;
        }

        if (!systrayAvailable) {
          app.visible = true
          app.raise()
        }
    }


    property var icons: {
        "off": "qrc:/assets/icon/png/white/vpn_off.png",
        "on": "qrc:/assets/icon/png/white/vpn_on.png",
        "wait": "qrc:/assets/icon/png/white/vpn_wait_0.png",
        "blocked": "qrc:/assets/icon/png/white/vpn_blocked.png"
    }

    VpnState {
        id: vpn
    }

    SystemTrayIcon {

        id: systray
        visible: systrayVisible

        /* the systray menu cannot be buried in a child qml file because
         * otherwise the ids are not available
         * from other components
         */
        menu: Menu {

            id: systrayMenu

            MenuItem {
                id: statusItem
                text: qsTr("Checking status…")
                enabled: false
            }

            MenuSeparator {}

            MenuItem {
                id: autoSelectionItem
                text: qsTr("Recommended")
                checkable: true
                checked: !isManualLocation()
                onTriggered: {
                    backend.useAutomaticGateway()
                    manualSelectionItem.checked = false
                }
            }

            /* a minimal segfault for submenu */
            // Menu {}
            MenuItem {
                id: manualSelectionItem
                text: {
                    if (isManualLocation()) {
                        locationStr()
                    } else {
                        qsTr("Pick location…")
                    }
                }
                checkable: true
                checked: isManualLocation()
                onTriggered: setGwSelection()
            }

            MenuSeparator {}

            MenuItem {
                text: {
                    if (vpn.state == "failed")
                        qsTr("Reconnect")
                    else
                        qsTr("Turn on")
                }
                onTriggered: {
                    backend.switchOn()
                }
                visible: ctx ? (ctx.status == "off"
                                || ctx.status == "failed") : false
            }

            MenuItem {
                text: {
                    if (ctx && ctx.status == "starting")
                        qsTr("Cancel")
                    else
                        qsTr("Turn off")
                }
                onTriggered: {
                    backend.switchOff()
                }
                visible: ctx ? (ctx.status == "on" || ctx.status == "starting"
                                || ctx.status == "failed") : false
            }

            MenuSeparator {}

            MenuItem {
                text: qsTr("About…")
                onTriggered: {
                    about.visible = true
                    requestActivate()
                }
            }

            MenuItem {
                id: donateItem
                text: qsTr("Donate…")
                visible: ctx ? ctx.donateURL : false
                onTriggered: {
                    donate.visible = true
                }
            }

            MenuSeparator {}

            MenuItem {
                text: qsTr("Help…")

                onTriggered: {
                    console.debug(Qt.resolvedUrl(ctx.helpURL))
                    Qt.openUrlExternally(Qt.resolvedUrl(ctx.helpURL))
                }
            }

            MenuItem {
                text: qsTr("Report a bug…")

                onTriggered: {
                    Qt.openUrlExternally(
                                Qt.resolvedUrl(
                                    "https://0xacab.org/leap/bitmask-vpn/issues"))
                }
            }

            MenuSeparator {}

            MenuItem {
                text: qsTr("Quit")
                onTriggered: backend.quit()
            }
        }

        Component.onCompleted: {
            systray.icon.source = icons["off"]
            tooltip = qsTr("Checking status…")
            console.debug("systray init completed")
            if (systrayVisible) {
                console.log("show systray")
                if (Qt.platform.os === "windows") {
                    let appname = ctx ? ctx.appName : "VPN"
                    Logic.showNotification(
                        ctx,
                        appname
                        + " is up and running. Please use system tray icon to control it.")
                }
            }
        }
    }

    DonateDialog {
        id: donate
        visible: false
    }

    AboutDialog {
        id: about
        visible: false
    }

    LoginDialog {
        id: login
        visible: false
    }

    LoginOKDialog {
        id: loginOk
        visible: false
    }

    MessageDialog {
        id: errorStartingVPN
        modality: Qt.NonModal
        title: qsTr("Error starting VPN")
        text: ""
        detailedText: ""
        visible: false
    }

    MessageDialog {
        id: authAgent
        modality: Qt.NonModal
        title: qsTr("Missing authentication agent")
        text: qsTr("Could not find a polkit authentication agent. Please run one and try again.")
        visible: false
    }

    FailDialog {
        id: initFailure
        visible: false
    }

    function locationStr() {
        return ctx.currentLocation + ", " + ctx.currentCountry
    }

    function useBridges(value) {
        if (value==true) {
            backend.setTransport("obfs4")
        } else {
            backend.setTransport("openvpn")
        }
    }

    property alias brReconnect:bridgesTab.displayReconnect

}
