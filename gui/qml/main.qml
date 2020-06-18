import QtQuick 2.9
import QtQuick.Controls 2.2
import QtQuick.Dialogs 1.2
import QtQuick.Extras 1.2
import Qt.labs.platform 1.1

ApplicationWindow {

    id: app
    visible: false

    property var ctx

    Connections {
        target: jsonModel
        onDataChanged: {
            ctx = JSON.parse(jsonModel.getJson());
            if (ctx.donateDialog == 'true') {
                console.debug(jsonModel.getJson())
                donate.visible = true
            }
            if (ctx.errors ) {
               // TODO consider disabling on/off buttons, or quit after closing the dialog
               if ( ctx.errors  == "nohelpers" ) {
                   showInitFailure(qsTr("Could not find helpers. Check your installation"))
               } else if ( ctx.errors == "nopolkit" ) {
                   showInitFailure(qsTr("Could not find polkit agent."))
               } else {
                   console.debug(ctx.errors)
               }
            }
        }
    }

    function showInitFailure(msg) {
          initFailure.text = msg
          initFailure.visible  = true
    }

    Component.onCompleted: {
        /* stupid as it sounds, windows doesn't like to have the systray icon
         not being attached to an actual application window.
         We can still use this quirk, and can use the AppWindow with deferred
         Loaders as a placeholder for all the many dialogs, or to load
         a nice splash screen etc...  */

        app.visible = true;
        show();
        hide();
    }

    function toHuman(st) {
        switch(st) {
            case "off":
                // TODO improve string interpolation, give context to translators etc
                return qsTr(ctx.appName + " off");
            case "on":
                return qsTr(ctx.appName + " on");
            case "connecting":
                return qsTr("Connecting to " + ctx.appName);
            case "stopping":
                return qsTr("Stopping " + ctx.appName);
            case "failed":
                return qsTr(ctx.appName + " blocking internet"); // TODO failed is not handed yet
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
        visible: true

        onActivated: {
            // this looks like a widget bug. middle click (reasons 3 or 4)
            // produce a segfault when trying to call menu.open()
            // left and right click seem to be working fine, so let's ignore this for now.
            switch (reason) {
            case SystemTrayIcon.Unknown:
                break
            case SystemTrayIcon.Context:
                break
            case SystemTrayIcon.DoubleClick:
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
            show();
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
                        PropertyChanges { target: systray; tooltip: toHuman("failed"); icon.source: icons["wait"] }
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
                text: qsTr("Donate...")
                visible: true
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

    MessageDialog {
        id: initFailure
        buttons: MessageDialog.Ok
        modality: Qt.NonModal
        title: qsTr("Initialization Error")
        text: ""
        visible: false
    }
}
