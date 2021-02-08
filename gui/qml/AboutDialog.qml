import QtQuick 2.0
import QtQuick.Dialogs 1.2

MessageDialog {
    title: qsTr("About")
    text: getText()
    informativeText: getVersion()
    icon: StandardIcon.Information

    function getText() {
        var _name = ctx ? ctx.appName : "vpn"
        var _provider = ctx ? ctx.provider : "unknown"
        var _donateURL = ctx ? ctx.donateURL : ""
        var _tosURL = ctx ? ctx.tosURL : "..."
        var _donateTXT = ""
        if (_donateURL) {
            //: donation text of the about dialog
            _donateTXT = qsTr(
                        "<p>This service is paid for entirely by donations from users like you. <a href=\"%1\">Please donate</a>.</p>").arg(
                        _donateURL)
        }
        //: about dialog
        //: %1 -> application name
        //: %2 -> provider name
        //: %3 -> donation text if activated
        var _txt = qsTr(
                    "<p>%1 is an easy, fast, and secure VPN service from %2. %1 does not require a user account, keep logs, or track you in any way.</p> %3 <p>By using this application, you agree to the <a href=\"%4\">Terms of Service</a>. This service is provided as-is, without any warranty, and is intended for people who work to make the world a better place.</p>").arg(
                    _name).arg(_provider).arg(_donateTXT).arg(_tosURL)
        return _txt
    }

    function getVersion() {
        var _name = ctx ? ctx.appName : "vpn"
        var _ver = ctx ? ctx.version : "unknown"
        //: %1 -> application name
        //: %2 -> version string
        var _txt = qsTr("%1 version: %2").arg(_name).arg(_ver)
        return _txt
    }
}
