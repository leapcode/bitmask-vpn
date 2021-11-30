

/*
 TODO (ui rewrite)
 See https://0xacab.org/leap/bitmask-vpn/-/issues/523
 - [ ] control actions from systray
 - [ ] add gateway to systray
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

    property int appHeight: 460
    property int appWidth: 280

    width: appWidth
    minimumWidth: appWidth
    maximumWidth: appWidth

    height: appHeight
    minimumHeight: appHeight
    maximumHeight: appHeight

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

    // FIXME get svg icons
    property var icons: {
        "off": "qrc:/assets/icon/png/white/vpn_off.png",
        "on": "qrc:/assets/icon/png/white/vpn_on.png",
        "wait": "qrc:/assets/icon/png/white/vpn_wait_0.png",
        "blocked": "qrc:/assets/icon/png/white/vpn_blocked.png"
    }

    signal openDonateDialog()

    FontLoader {
        id: lightFont
        source: "qrc:/poppins-regular.ttf"
    }

    FontLoader {
        id: boldFont
        source: "qrc:/poppins-bold.ttf"
    }

    FontLoader {
        id: boldFontMonserrat
        source: "qrc:/monserrat-bold.ttf"
    }

    FontLoader {
        id: robotoFont
        source: "qrc:/roboto.ttf"
    }

    FontLoader {
        id: robotoBoldFont
        source: "qrc:/roboto-bold.ttf"
    }

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
                locationsModel = getSortedLocations()
                //console.debug("Got sorted locations: " + locationsModel)
            }
            if (ctx.errors) {
                console.debug("errors, setting root.error")
                root.error = ctx.errors
            } else {
                root.error = ""
            }
            if (ctx.donateURL) {
                isDonationService = true
            }
            if (ctx.donateDialog == 'true') {
                showDonationReminder = true
            }

            // TODO check donation
            //if (needsDonate && !shownDonate) {
            //    donate.visible = true;
            //    shownDonate = true;
            //    // move this to onClick of "close" for widget
            //    backend.donateSeen();
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

    function getSortedLocations() {
        let obj = ctx.locations
        var arr = []
        for (var prop in obj) {
            if (obj.hasOwnProperty(prop)) {
                arr.push({
                             "key": prop,
                             "value": obj[prop]
                         })
            }
        }
        arr.sort(function (a, b) {
            return a.value - b.value
        }).reverse()
        return Array.from(arr, (k,_) => k.key);
    }

    function bringToFront() {
        // FIXME does not work properly, at least on linux 
        if (visibility == 3) {
            showNormal()
        } else {
            show() 
        }
        raise()
        requestActivate()
    }

    onSceneGraphError: function (error, msg) {
        console.debug("ERROR while initializing scene")
        console.debug(msg)
    }

    Component.onCompleted: {
        loader.source = "components/Splash.qml"
        // XXX workaround for custom font not working in osx
        /*
        if (Qt.platform.os === "osx") {
            root.font.family = robotoFont.name
            root.font.weight = Font.Light
        }
        */
    }
}
