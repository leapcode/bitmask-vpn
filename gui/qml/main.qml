import QtQuick 2.9
import QtQuick.Controls 1.4
import QtQuick.Dialogs 1.2
import QtQuick.Extras 1.2
import Qt.labs.platform 1.1

ApplicationWindow {

    id: app
    visible: false

    property var ctx
    property var loginDone 
    property var allowEmptyPass

    Connections {
        target: jsonModel
        onDataChanged: {
            ctx = JSON.parse(jsonModel.getJson());

            // FIXME -- we need to inform the backend that we've already seen
            // this. Otherwise this keeps popping randonmly on state changes.
            if (ctx.donateDialog == 'true') {
                console.debug(jsonModel.getJson())
                donate.visible = true
            }
            if (ctx.loginDialog == 'true') {
                console.debug(jsonModel.getJson())
                console.debug("DEBUG: should display login")
                login.visible = true
            }
            if (ctx.loginOk == 'true') {
                loginOk.visible = true
            }
            if (ctx.errors ) {
               login.visible = false
               if ( ctx.errors  == "nohelpers" ) {
                   showInitFailure(qsTr("Could not find helpers. Check your installation"))
               } else if ( ctx.errors == "nopolkit" ) {
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
          if (ctx.errors == 'bad_auth_502' || ctx.errors == 'bad_auth_timeout') {
                  msg = qsTr("Oops! The authentication service seems down. Please try again later")
              initFailure.title = qsTr("Service Error")
          }
          else if (ctx.errors == 'bad_auth') {
              if (allowEmptyPass) {
                  // For now, this is a libraryVPN, so we can be explicit about what credentials are here.
                  // Another option to consider is to customize the error strings while vendoring.
                  msg = qsTr("Please check your Patron ID")
              } else {
                  msg = qsTr("Could not log in with those credentials, please retry")
              }
              initFailure.title = qsTr("Login Error")
          } else {
              //: %1 -> application name
              //: %2 -> error string
              msg = qsTr("Got an error starting %1: %2").arg(ctx.appName).arg(ctx.errors)
          }
      }
      initFailure.text = msg
      initFailure.visible  = true
    }

    function shouldAllowEmptyPass() {
        let obj = JSON.parse(providers.getJson())
        let active = obj['default']
        let allProviders = obj['providers']
        for (let i = 0; i < allProviders.length; i++) {
            if (allProviders[i]['name'] === active) {
                return (allProviders[i]['authEmptyPass'] === 'true')
            }
        }
        return false
    }

    Component.onCompleted: {
        loginDone = false;

        /* stupid as it sounds, windows doesn't like to have the systray icon
         not being attached to an actual application window.
         We can still use this quirk, and can use the AppWindow with deferred
         Loaders as a placeholder for all the many dialogs, or to load
         a nice splash screen etc...  */

        console.debug("DEBUG: Pre-seeded providers:");
        console.debug(providers.getJson());

        allowEmptyPass = shouldAllowEmptyPass()

        app.visible = true;
        //show();
        //hide();
    }

    function toHuman(st) {
        switch(st) {
            case "off":
		//: %1 -> application name
                return qsTr("%1 off").arg(ctx.appName);
            case "on":
		//: %1 -> application name
                return qsTr("%1 on").arg(ctx.appName);
            case "connecting":
		//: %1 -> application name
                return qsTr("Connecting to %1").arg(ctx.appName);
            case "stopping":
		//: %1 -> application name
                return qsTr("Stopping %1").arg(ctx.appName);
            case "failed":
		//: %1 -> application name
                return qsTr("%1 blocking internet").arg(ctx.appName); // TODO failed is not handed yet
        }
    }

    property var icons: {
        "off":     "qrc:/assets/icon/png/black/vpn_off.png",
        "on":      "qrc:/assets/icon/png/black/vpn_on.png",
        "wait":    "qrc:/assets/icon/png/black/vpn_wait_0.png",
        "blocked": "qrc:/assets/icon/png/black/vpn_blocked.png",
    }
    

    SystemTrayIcon {

        id: systray
        visible: systrayVisible

        onActivated: {
            // this looks like a widget bug. middle click (reasons 3 or 4)
            // produce a segfault when trying to call menu.open()
            // left and right click seem to be working fine, so let's ignore this for now.
            switch (reason) {
            case SystemTrayIcon.Unknown:
	        console.debug("reason: unknown")
                menu.open()
		break
            case SystemTrayIcon.Context:
	        console.debug("activated: context")
		if (Qt.platform.os !== "linux") {
			menu.open()
		}
                break
            case SystemTrayIcon.DoubleClick:
	        console.debug("activated: double click")
		if (Qt.platform.os !== "linux") {
			menu.open()
		}
                break
            case SystemTrayIcon.Trigger:
                break
            case SystemTrayIcon.MiddleClick:
                break
            }
        }

        Component.onCompleted: {
            icon.source = icons["off"]
            tooltip = qsTr("Checking status...")
            console.debug("systray init completed")
            hide();
            if (systrayVisible) {
                show();
            }
        }

        menu: Menu {

            id: systrayMenu

            StateGroup {
                id: vpn
                state: ctx ? ctx.status : ""
            
                states: [
                    State { name: "initializing" },
                    State {
                        name: "off"
                        PropertyChanges { target: systray; tooltip: toHuman("off"); icon.source: icons["off"] }
                        PropertyChanges { target: statusItem; text: toHuman("off") }
                    },
                    State {
                        name: "on"
                        PropertyChanges { target: systray; tooltip: toHuman("on"); icon.source: icons["on"] }
                        PropertyChanges { target: statusItem; text: toHuman("on") }
                    },
                    State {
                        name: "starting"
                        PropertyChanges { target: systray; tooltip: toHuman("connecting"); icon.source: icons["wait"] }
                        PropertyChanges { target: statusItem; text: toHuman("connecting") }
                    },
                    State {
                        name: "stopping"
                        PropertyChanges { target: systray; tooltip: toHuman("stopping"); icon.source: icons["wait"] }
                        PropertyChanges { target: statusItem; text: toHuman("stopping") }
                    },
                    State {
                        name: "failed"
                        PropertyChanges { target: systray; tooltip: toHuman("failed"); icon.source: icons["blocked"] }
                        PropertyChanges { target: statusItem; text: toHuman("failed") }
                    }
                ]
            }

            MenuItem {
                id: statusItem
                text: qsTr("Checking status...")
                enabled: false
            }

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
                visible: ctx ? (ctx.status == "off" || ctx.status == "failed") : false
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
                visible: ctx ? (ctx.status == "on" || ctx.status == "starting" || ctx.status == "failed") : false
            }

            MenuSeparator {}

            MenuItem {
                text: qsTr("Help...")
                onTriggered: Qt.openUrlExternally(ctx.helpURL)
            }

            MenuItem {
                id: donateItem
                text: qsTr("Donate...")
                visible: ctx ? ctx.donateURL : false
                onTriggered: { donate.visible = true }
            }

            MenuItem {
                text: qsTr("About...")
                onTriggered: { about.visible = true }
            }

            MenuSeparator {}

            MenuItem {
                text: qsTr("Quit")
                onTriggered: backend.quit()
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

    LoginOKDialog{
        id: loginOk
        visible: false
    }

    MessageDialog {
        id: errorStartingVPN
        buttons: MessageDialog.Ok
        modality: Qt.NonModal
        title: qsTr("Error starting VPN")
        text: ""
        detailedText: ""
        visible: false
    }

    MessageDialog {
        id: authAgent
        buttons: MessageDialog.Ok
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
