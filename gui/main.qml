/*
 TODO (ui rewrite)
 - [x] add systray
 - [x] systray status
 - [x] splash screen
 - [ ] splash delay/transitions
 - [ ] nested states
 - [ ] splash init errors
 - [ ] font: monserrat
 - [ ] donation dialog
 - [ ] add gateway to systray
 - [ ] control actions from systray
 - [ ] minimize/hide from systray
 - [ ] parse ctx flags (need dialog, etc)
 - [ ] gateway selector
 - [ ] bridges
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

    property var icons: {
        "off": "qrc:/assets/icon/png/white/vpn_off.png",
        "on": "qrc:/assets/icon/png/white/vpn_on.png",
        "wait": "qrc:/assets/icon/png/white/vpn_wait_0.png",
        "blocked": "qrc:/assets/icon/png/white/vpn_blocked.png"
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
	    ctx = JSON.parse(jsonModel.getJson())
            if (qmlDebug) {
                console.debug(jsonModel.getJson())
            }

	    // FIXME -- use nested state machines for all these cases.

	    //gwSelector.model = Object.keys(ctx.locations)

	    /*
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

	    if (ctx.status == "on") {
		gwNextConnectionText.visible = false
		gwReconnectText.visible = false
	    }
	    */
	}
    }

    onSceneGraphError: function(error, msg) {
        console.debug("ERROR while initializing scene")
        console.debug(msg)
    }

    Component.onCompleted: {
        loader.source = "components/Splash.qml"
    }
}
