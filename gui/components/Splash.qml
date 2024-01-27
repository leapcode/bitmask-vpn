import QtQuick 2.15
import QtQuick.Controls 2.2
import Qt5Compat.GraphicalEffects
import "../themes/themes.js" as Theme

Page {
    id: splash
    property int timeoutInterval: qmlDebug ? 600 : 1600
    property alias errors: splashErrorBox

    ToolButton {
        id: closeButton 
        visible: false
        anchors {
            right: parent.right
            //rightMargin: -10
        }
        icon.source: "../resources/close.svg"
        HoverHandler {
            cursorShape: Qt.PointingHandCursor
        }
        onClicked: {
            loader.source = "components/MainView.qml"
        }
    }

    Column {
        width: parent.width * 0.8
        anchors.horizontalCenter: parent.horizontalCenter
        anchors.topMargin: 24

        MotdBox {
            id: motd
            visible: false
            anchors {
                top: parent.top
                topMargin: 100
                bottomMargin: 30
            }
        }

        VerticalSpacer {
            id: motdSpacer 
            visible: false
            height: 100
        }

        VerticalSpacer {
            id: upperSpacer
            visible: true
            height: root.height * 0.25
        }

        Image {
            id: connectionImage
            height: 180
            anchors.horizontalCenter: parent.horizontalCenter
            source: customTheme.iconSplash
            fillMode: Image.PreserveAspectFit
        }

        VerticalSpacer {
            id: middleSpacer
            visible: true
            height: root.height * 0.05
        }

        ProgressBar {
            id: splashProgress
            width: appWidth * 0.8 - 60
            indeterminate: true
            anchors.horizontalCenter: parent.horizontalCenter
        }

        InitErrors {
            id: splashErrorBox
        }
    } // end Column

    Image {
        id: motdImage
        visible: false
        height: 100
        anchors.horizontalCenter: parent.horizontalCenter
        anchors.bottom: parent.bottom
        anchors.bottomMargin: 50
        source: customTheme.iconSplash
        fillMode: Image.PreserveAspectFit
    }

    Timer {
        id: splashTimer
    }

    function hasMotd() {
        return needsUpgrade() || (ctx && !isEmptyMotd(ctx.motd))
    }

    function getUpgradeText() {
        return qsTr("There is a newer version available. ") + qsTr("Make sure to <a href=\"https://0xacab.org/leap/bitmask-vpn/-/blob/main/docs/uninstall.md\">uninstall</a> the previous one before running the new installer.")
    }

    function getUpgradeLink() {
        return "<a href='" + getLinkURL() + "'>" + qsTr("UPGRADE NOW") + "</a>";
    }

    function getLinkURL() {
        return "https://downloads.leap.se/RiseupVPN/" + Qt.platform.os + "/"
    }

    function needsUpgrade() {
        if (ctx && isTrue(ctx.canUpgrade)) {
            if (qmlDebug) {
                return true
            }
            let platform = Qt.platform.os
            //DEBUG -------------------------------------------------------------------
            //if (platform == "windows" || platform == "osx" || platform == "linux" ) {
            //DEBUG -------------------------------------------------------------------
            if (platform == "windows" || platform == "osx") {
                    return true
            }
        }
        return false 
    }

    function showMotd() {
        // XXX this is not picking locales configured by LANG or LC_ALL
        // Need to fix this; probably also with allowing to select translation
        // manually on runtime.
        let isUpgrade = false
        let lang = Qt.locale().name.substring(0,2)
        let messages = JSON.parse(ctx.motd)
        let platform = Qt.platform.os
        let textEn = ""
        let textLocale = ""
        let link = ""

        if (needsUpgrade()) {
            isUpgrade = true;
            textLocale = getUpgradeText();
            link = getUpgradeLink();
        } else {
            // TODO fallback in case upgrade has no text
            console.debug("configured locale: " + lang)
            console.debug("platform: " + Qt.platform.os)
            for (let i=0; i < messages.length; i++) {
                let m = messages[i]
                if (m.platform == "all" || m.platform == platform) {
                    for (let k=0; k < m.text.length; k++) {
                        if (m.text[k].lang == lang) {
                            textLocale = m.text[k].str
                            break
                        } else if (m.text[k].lang == "en") {
                            textEn = m.text[k].str
                        }
                    }
                    break
                }
            }
        }
        if (isUpgrade) {
            upperSpacer.height = 100
        } else {
            // TODO get proportional to textLocale/textEn
            upperSpacer.height = 50
        }
        //connectionImage.height = 100
        connectionImage.visible = false
        motdImage.visible = true
        middleSpacer.visible = false
        splashProgress.visible = false
        motd.visible = true
        motdSpacer.visible = true
        motd.motdText = textLocale ? textLocale : textEn
        motd.motdLink = link
        motd.url = getLinkURL()
        // FIXME if no text, just skip to main view
        closeButton.visible = true
    }

    function delay(delayTime, cb) {
        splashTimer.interval = delayTime
        splashTimer.repeat = true
        splashTimer.triggered.connect(cb)
        splashTimer.start()
    }

    function loadMainViewWhenReady() {
        if (!isEmpty(root.error)) {
            return
        }
        if (ctx && isTrue(ctx.isReady) || qmlDebug) {
            splashTimer.stop()
            if (hasMotd()) {
                console.debug("show motd");
                showMotd();
            } else {
                loader.source = "components/MainView.qml"
            }
        } else {
            if (!splashTimer.running) {
              console.debug('delay...')
              delay(500, loadMainViewWhenReady)
            }
        }
    }

    Timer {
        interval: timeoutInterval
        running: true
        repeat: false
        onTriggered: {
            loadMainViewWhenReady()
        }
    }

    Component.onCompleted: {
    }

    function isTrue(val) {
        return val == "true";
    }

    function isEmpty(val) {
         return val==undefined ? true : val.length == 0;
    }

    function isEmptyMotd(motd) {
        let m = JSON.parse(motd)
        let first = m[0]
        if (first == undefined) {
            return true
        }
        return isEmpty(first.text)
    }

}