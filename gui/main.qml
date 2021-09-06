

/*
 TODO (ui rewrite)
 - [x] add systray
 - [x] systray status
 - [x] splash screen
 - [x] splash delay/transitions
 - [x] font: monserrat
 - [x] nested states
 - [x] splash init errors
 - [.] gateway selector
 - [ ] bridges
 - [ ] minimize/hide from systray
 - [ ] control actions from systray
 - [ ] add gateway to systray
 - [ ] donation dialog
 - [ ] parse ctx flags (need dialog, etc)
 - [ ] udp support
*/
import QtQuick 2.0
import QtQuick.Controls 2.4
import QtQuick.Dialogs 1.2
import QtQuick.Controls.Material 2.1
import QtQuick.Layouts 1.14

import "./components"

ApplicationWindow {

    id: root

    visible: true
    width: 360
    height: 520
    minimumWidth: 300
    maximumWidth: 300
    minimumHeight: 500
    maximumHeight: 500

    title: ctx ? ctx.appName : "VPN"
    Material.accent: Material.Green

    property var ctx
    property var error: ""

    // TODO can move properties to some state sub-item to unclutter
    property bool isDonationService: false
    property bool showDonationReminder: false
    property var locationsModel: []
    // TODO get from persistance
    property var selectedGateway: "auto"

    property var icons: {
        "off": "qrc:/assets/icon/png/white/vpn_off.png",
        "on": "qrc:/assets/icon/png/white/vpn_on.png",
        "wait": "qrc:/assets/icon/png/white/vpn_wait_0.png",
        "blocked": "qrc:/assets/icon/png/white/vpn_blocked.png"
    }

    FontLoader {
        id: lightFont
        source: "qrc:/montserrat-light.ttf"
    }

    FontLoader {
        id: boldFont
        source: "qrc:/montserrat-bold.ttf"
    }

    font.family: lightFont.name

    Loader {
        id: loader
        asynchronous: true
        anchors.fill: parent
    }

    Systray {
        id: systray
    }


    Connections {
        target: jsonModel
        function onDataChanged() {
            let j = jsonModel.getJson()
            if (qmlDebug) {
                console.debug(j)
            }
            ctx = JSON.parse(j)
            if (ctx != undefined) {
                locationsModel = Object.keys(ctx.locations)
            }
            if (ctx.errors) {
                console.debug("errors, setting root.error")
                root.error = ctx.errors
            } else {
                root.error = ""
            }
            if (ctx.donateURL) {
                isDonationService = true;
            }
            if (ctx.donateDialog == 'true') {
                showDonationReminder = true;
            }

            // TODO check donation
            //if (needsDonate && !shownDonate) {
            //    donate.visible = true;
            //    shownDonate = true;
            //    // move this to onClick of "close" for widget
            //    backend.donateSeen();
            //}
            // TODO refactor donate widget into main view (with close window!)
            //if (ctx.status == "on") {
            //    gwNextConnectionText.visible = false
            //    gwReconnectText.visible = false
            // when: vpn.status == "on"
            //}

            /*
            TODO libraries need login 
            if (ctx.loginDialog == 'true') {
                login.visible = true
            }
            if (ctx.loginOk == 'true') {
                loginOk.visible = true
            }
            */
        }
    }

    onSceneGraphError: function (error, msg) {
        console.debug("ERROR while initializing scene")
        console.debug(msg)
    }

    Component.onCompleted: {
        loader.source = "components/Splash.qml"
    }
}
