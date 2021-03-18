import QtQuick 2.9
import QtQuick.Window 2.2
import QtQuick.Dialogs 1.2
import QtQuick.Layouts 1.0
import QtQuick.Controls 2.12

import Qt.labs.platform 1.0

Window {
    id: app
    visible: true
    width: 300
    height: 600
    maximumWidth: 300
    minimumWidth: 300
    maximumHeight: 600
    minimumHeight: 600

    flags: Qt.WindowsStaysOnTopHint

    // TODO get a nice background color
    property var ctx
    property var loginDone
    property var allowEmptyPass

    GridLayout {

        visible: true
        columns: 3

        Item {
            Layout.column: 2
            Layout.topMargin: app.height * 0.15
            Layout.leftMargin: app.width * 0.10

            ColumnLayout {

                Layout.alignment: Qt.AlignHCenter

                Text {
                    id: mainStatus
                    text: "off"
                    font.pixelSize: 26
                    Layout.alignment: Text.AlignHCenter
                }

                Text {
                    id: mainCurrentGateway
                    text: ""
                    font.pixelSize: 20
                    Layout.alignment: Text.AlignHCenter
                }

                SwitchDelegate {

                    id: vpntoggle

                    text: qsTr("")
                    checked: false

                    Connections {
                        onCheckedChanged: {
                            if (vpntoggle.checked == true && ctx.status == "off") {
                                backend.switchOn()
                            }
                            if (vpntoggle.checked === false && ctx.status == "on") {
                                backend.switchOff()
                            }
                        }
                    }

                    contentItem: Text {
                        rightPadding: vpntoggle.indicator.width + control.spacing
                        text: vpntoggle.text
                        font: vpntoggle.font
                        opacity: enabled ? 1.0 : 0.3
                        color: vpntoggle.down ? "#17a81a" : "#21be2b"
                        elide: Text.ElideRight
                        verticalAlignment: Text.AlignVCenter
                    }

                    indicator: Rectangle {
                        implicitWidth: 48
                        implicitHeight: 26
                        x: vpntoggle.width - width - vpntoggle.rightPadding
                        y: parent.height / 2 - height / 2
                        radius: 13
                        color: vpntoggle.checked ? "#17a81a" : "transparent"
                        border.color: vpntoggle.checked ? "#17a81a" : "#cccccc"

                        Rectangle {
                            x: vpntoggle.checked ? parent.width - width : 0
                            width: 26
                            height: 26
                            radius: 13
                            color: vpntoggle.down ? "#cccccc" : "#ffffff"
                            border.color: vpntoggle.checked ? (vpntoggle.down ? "#17a81a" : "#21be2b") : "#999999"
                        }
                    }

                    background: Rectangle {
                        implicitWidth: 100
                        implicitHeight: 40
                        visible: vpntoggle.down || vpntoggle.highlighted
                        color: vpntoggle.down ? "#bdbebf" : "#eeeeee"
                    }
                }

                ComboBox {
                    id: gwSelector
                    editable: false
                    model: [qsTr("Automatic")]
                    onActivated: {
                        console.debug("Selected gateway:", currentText)
                        backend.useGateway(currentText.toString())
                    }
                }
            }
        }
    }

    Connections {
        target: jsonModel
        onDataChanged: {
            ctx = JSON.parse(jsonModel.getJson())
            // TODO pass QML_DEBUG variable to be hyper-verbose
            //console.debug(jsonModel.getJson())
            gwSelector.model = Object.keys(ctx.gateways)

            if (ctx.donateDialog == 'true') {
                console.debug(jsonModel.getJson())
                donate.visible = true
                backend.donateSeen()
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

    function shouldAllowEmptyPass() {
        let obj = JSON.parse(providers.getJson())
        let active = obj['default']
        let allProviders = obj['providers']
        for (var i = 0; i < allProviders.length; i++) {
            if (allProviders[i]['name'] === active) {
                return (allProviders[i]['authEmptyPass'] === 'true')
            }
        }
        return false
    }

    Component.onCompleted: {
        loginDone = false
        console.debug("Platform:", Qt.platform.os)
        console.debug("DEBUG: Pre-seeded providers:")
        console.debug(providers.getJson())
        allowEmptyPass = shouldAllowEmptyPass()

        /* TODO get appVisible flag from backend */
        app.visible = true
        app.raise()
    }

    function toHuman(st) {
        switch (st) {
        case "off":
            //: %1 -> application name
            return qsTr("%1 off").arg(ctx.appName)
        case "on":
            //: %1 -> application name
            return qsTr("%1 on").arg(ctx.appName)
        case "connecting":
            //: %1 -> application name
            return qsTr("Connecting to %1").arg(ctx.appName)
        case "stopping":
            //: %1 -> application name
            return qsTr("Stopping %1").arg(ctx.appName)
        case "failed":
            //: %1 -> application name
            return qsTr("%1 blocking internet").arg(
                        ctx.appName) // TODO failed is not handed yet
        }
    }

    /* TODO change this!!! automatic: Paris (FR) */

    function toHumanWithLocation(st) {
        switch (st) {
        case "off":
            //: %1 -> application name
            return qsTr("%1 off").arg(ctx.appName)
        case "on":
            //: %1 -> application name
            //: %2 -> current gateway
            return qsTr("%1 on - %2").arg(ctx.appName).arg(ctx.currentLocation)
        case "connecting":
            //: %1 -> application name
            //: %2 -> current gateway
            return qsTr("Connecting to %1 - %2").arg(ctx.appName).arg(
                        ctx.currentLocation)
        case "stopping":
            //: %1 -> application name
            return qsTr("Stopping %1").arg(ctx.appName)
        case "failed":
            //: %1 -> application name
            return qsTr("%1 blocking internet").arg(
                        ctx.appName) // TODO failed is not handed yet
        }
    }

    property var icons: {
        "off": "qrc:/assets/icon/png/black/vpn_off.png",
        "on": "qrc:/assets/icon/png/black/vpn_on.png",
        "wait": "qrc:/assets/icon/png/black/vpn_wait_0.png",
        "blocked": "qrc:/assets/icon/png/black/vpn_blocked.png"
    }

    VpnState {
        id: vpn
    }

    SystemTrayIcon {

        id: systray
        visible: systrayVisible
        signal activatedSignal

        onActivated: {
            systray.activatedSignal()
        }


        /* the systray menu cannot be buried in a child qml file because
         * otherwise the ids are not available
         * from other components
         */
        menu: Menu {

            id: systrayMenu

            Connections {
                target: systray
                onActivatedSignal: {
                    if (Qt.platform.os === "windows" || desktop === "LXQt") {
                        console.debug("open systray menu")
                        systrayMenu.open()
                    }
                }
            }

            MenuItem {
                id: statusItem
                text: qsTr("Checking status…")
                enabled: false
            }

            MenuItem {
                id: autoSelectionItem
                text: qsTr("automatic")
                checkable: true
                checked: true
                enabled: false
            }


            /* a minimal segfault for submenu */
            // Menu {}

            MenuSeparator {}

            MenuItem {
                id: manualSelectionItem
                text: qsTr("Pick gateway…")
                checkable: true
                checked: false
                enabled: true
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
                    app.focus = true
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
            icon.source = icons["off"]
            tooltip = qsTr("Checking status…")
            console.debug("systray init completed")
            hide()
            if (systrayVisible) {
                console.log("show systray")
                show()
                if (Qt.platform.os === "windows") {
                    let appname = ctx ? ctx.appName : "VPN"
                    showNotification(
                                appname
                                + " is up and running. Please use system tray icon to control it.")
                }
            }
        }

        // Helper to show notification messages
        function showNotification(msg) {
            console.log("Going to show notification message: ", msg)
            if (supportsMessages) {
                let appname = ctx ? ctx.appName : "VPN"
                showMessage(appname, msg, null, 15000)
            } else {
                console.log("System doesn't support systray notifications")
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
        //buttons: MessageDialog.Ok
        modality: Qt.NonModal
        title: qsTr("Error starting VPN")
        text: ""
        detailedText: ""
        visible: false
    }

    MessageDialog {
        id: authAgent
        //buttons: MessageDialog.Ok
        modality: Qt.NonModal
        title: qsTr("Missing authentication agent")
        text: qsTr("Could not find a polkit authentication agent. Please run one and try again.")
        visible: false
    }

    FailDialog {
        id: initFailure
        visible: false
    }
}
