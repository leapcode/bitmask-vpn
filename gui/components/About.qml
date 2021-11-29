import QtQuick 2.9
import QtQuick.Controls 2.2
import QtQuick.Controls.Material 2.1

import "../themes/themes.js" as Theme

ThemedPage {

    title: qsTr("About")

    Rectangle {
        anchors.horizontalCenter: parent.horizontalCenter
        width: root.appWidth * 0.9
        height: root.appHeight - 100
        radius: 10
        color: "white"

        anchors {
            fill: parent
            margins: 10
        }

        Column {
            id: aboutCol
            padding: 20
            width: parent.width 

            Text {
                textFormat: Text.RichText
                width: parent.width - 40
                color: Theme.fontColor
                font.pixelSize: 12
                wrapMode: Label.Wrap
                text: getText()
                anchors.horizontalCenter: parent.horizontalCenter
                onLinkActivated: Qt.openUrlExternally(link)
            }

            VerticalSpacer {
                visible: true
                height: 10
            }

            Image {
                id: aboutIcon
                height: 60
                source: "../resources/icon-noshield.svg"
                fillMode: Image.PreserveAspectFit
                anchors.horizontalCenter: parent.horizontalCenter
            }

            Rectangle {
                anchors.top: aboutIcon.bottom
                width: 100
                anchors.horizontalCenter: parent.horizontalCenter
                TextEdit {
                    width: 100
                    font.pixelSize: 10
                    readOnly: true
                    selectByMouse: true
                    text: getVersion()
                }
            }

            VerticalSpacer {
                visible: true
                height: 50
            }
        } // end column
    } // end rectangle

    function getText() {
        var _name = ctx ? ctx.appName : "vpn"
        var _provider = ctx ? ctx.provider : "unknown"
        var _donateURL = ctx ? ctx.donateURL : ""
        var _tosURL = ctx ? ctx.tosURL : "…"
        var _donateTXT = ""
        if (_donateURL) {
            //: donation text of the about dialog
            _donateTXT = qsTr(
                        "<p>This service is paid for entirely by donations from users like you. <a href=\"%1\">Please donate</a>.</p>").arg(_donateURL)
        }
        if (_provider == "Riseup") {
            // XXX this is a quick workaround. About-us should be parametrized too.
            _provider = "<a href=\"https://riseup.net/about-us\">" + ctx.provider + "</a>"
        }
        //: about dialog
        //: %1 -> application name
        //: %2 -> provider name
        //: %3 -> donation text if activated
        //: %4 -> terms of service
        let _txt = qsTr(
                    "<style>a:link {color: '" + Theme.green + "';}</style><p>%1 is an easy, fast, and secure VPN service from %2. %1 does not require a user account, keep logs, or track you in any way.</p> %3 <p>By using this application, you agree to the <a href=\"%4\">Terms of Service</a>. This service is provided as-is, without any warranty, and is intended for people who work to make the world a better place.</p>").arg(_name).arg(_provider).arg(_donateTXT).arg(_tosURL)
        return _txt
    }

    function getVersion() {
        var _name = ctx ? ctx.appName : "vpn"
        var _ver = ctx ? ctx.version : "unknown"
        //: %1 -> application name
        //: %2 -> version string
        var _txt = qsTr("%1 version: \n%2").arg(_name).arg(_ver)
        return _txt
    }
}
