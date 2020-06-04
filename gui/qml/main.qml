import QtQuick 2.9
import QtQuick.Controls 2.2
import QtQuick.Dialogs 1.2
import QtQuick.Extras 1.2
import Qt.labs.platform 1.1

ApplicationWindow {

    id: app
    visible: false

    property var     ctx

    Connections {
        target: jsonModel
        onDataChanged: {
            ctx = JSON.parse(jsonModel.getJson());
        }
    }

    Component.onCompleted: {
        /* stupid as it sounds, windows doesn't like to have the systray icon
         not being attached to an actual application window.
         We can still use this quirk, and can use the AppWindow with deferred
         Loaders as a placeholder for all the many dialogs, or to load
         a nice splash screen etc...  */
        app.visible = true;
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
            console.debug("app is", ctx.appName)
            menu.open()
        }

        Component.onCompleted: {
            icon.source = icons["off"]
            tooltip = qsTr("Checking status...")
            console.debug("systray init completed")
            show();
        }


        menu: Menu {
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

            /*
            LoginDialog {
                id: login
            }
            DonateDialog {
                id: donate
            }
            MessageDialog {
                id: about
                buttons: MessageDialog.Ok
                title: "About"
                text: "<p>%1 is an easy, fast, and secure VPN service from %2. %1 does not require a user account, keep logs, or track you in any way.</p>
    <p>This service is paid for entirely by donations from users like you. <a href=\"%3\">Please donate</a>.</p>
    <p>By using this application, you agree to the <a href=\"%4\">Terms of Service</a>. This service is provided as-is, without any warranty, and is intended for people who work to make the world a better place.</p>".arg(ctxSystray.applicationName).arg(ctxSystray.provider).arg(ctxSystray.donateURL).arg(ctxSystray.tosURL)
                informativeText: "%1 version: %2".arg(ctxSystray.applicationName).arg(ctxSystray.version)
            }
            MessageDialog {
                id: errorStartingVPN
                buttons: MessageDialog.Ok
                modality: Qt.NonModal
                title: "Error starting VPN"
                text: "Can't connect to %1".arg(ctxSystray.applicationName)
                detailedText: ctxSystray.errorStartingMsg
                visible: ctxSystray.errorStartingMsg != ""
            }
            MessageDialog {
                id: authAgent
                buttons: MessageDialog.Ok
                modality: Qt.NonModal
                title: "Missing authentication agent"
                text: "Could not find a polkit authentication agent. Please run one and try again."
                visible: ctxSystray.authAgent == true
            }
            MessageDialog {
                id: initFailure
                buttons: MessageDialog.Ok
                modality: Qt.NonModal
                title: "Initialization Error"
                text: ctxSystray.errorInitMsg
                visible: ctxSystray.errorInitMsg != ""
            }
            */

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
                //onTriggered: ctxSystray.help()
            }

            MenuItem {
                text: qsTr("Donate...")
                //onTriggered: ctxSystray.donate()
                visible: true
                //visible: ctx.showDonate
            }

            MenuItem {
                text: qsTr("About...")
                //onTriggered: about.open()
            }

            MenuSeparator {}

            MenuItem {
                text: qsTr("Quit")
                onTriggered: backend.quit()
            }
        }
    }
}
