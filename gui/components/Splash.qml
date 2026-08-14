import QtQuick
import QtQuick.Controls
import QtQuick.Effects
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

        VerticalSpacer {
            id: upperSpacer
            visible: true
            height: root.height * 0.25
        }

        MotdBox {
            id: motd
            visible: false
        }

        VerticalSpacer {
            id: motdSpacer
            visible: false
            height: 100
        }

        Image {
            id: connectionImage
            height: 135
            width: parent.width
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

    // Re-evaluate the MOTD text whenever a fresh ctx arrives from Go, so the
    // box is never left empty due to a timing gap between showMotd() running
    // and ctx.motd being fully populated. Mirrors the jsonModel onDataChanged
    // handler in main.qml.
    Connections {
        target: jsonModel
        function onDataChanged() {
            if (motd.visible) {
                updateMotdText()
            }
        }
    }

    function hasMotd() {
        return needsUpgrade() || (ctx && !isEmptyMotd(ctx.motd))
    }

    function getUpgradeText() {
        return qsTr("There is a newer version available.") + qsTr("Make sure to <a href=\"https://0xacab.org/leap/bitmask-vpn/-/blob/main/docs/uninstall.md\">uninstall</a> the previous one before running the new installer.")
    }

    function getUpgradeLink() {
        return "<a href='" + getLinkURL() + "'>" + qsTr("UPGRADE NOW") + "</a>";
    }

    function getLinkURL() {
        return "https://downloads.leap.se/RiseupVPN/" + Qt.platform.os + "/"
    }

    function needsUpgrade() {
        if (qmlDebug) {
            return true
        }
        return ctx && isTrue(ctx.canUpgrade)
    }

    function showMotd() {
        // Layout/visibility setup for the MOTD view. Text is selected in
        // updateMotdText(), which is also re-run whenever a fresh ctx arrives
        // (see Connections to jsonModel below) so the box is never left empty
        // due to a timing gap between showMotd() and ctx.motd being populated.
        if (needsUpgrade()) {
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
        closeButton.visible = true
        updateMotdText()
    }

    function updateMotdText() {
        // XXX this is not picking locales configured by LANG or LC_ALL
        // Need to fix this; probably also with allowing to select translation
        // manually on runtime.
        let lang = Qt.locale().name.substring(0,2)
        let platform = Qt.platform.os
        let textEn = ""
        let textLocale = ""
        let link = ""

        if (needsUpgrade()) {
            textLocale = getUpgradeText();
            link = getUpgradeLink();
        } else if (ctx && !isEmptyMotd(ctx.motd)) {
            let messages = JSON.parse(ctx.motd)
            console.debug("configured locale: " + lang)
            console.debug("platform: " + Qt.platform.os)
            // First pass: pick a message that targets this platform.
            let chosen = null
            for (let i=0; i < messages.length; i++) {
                if (messages[i].platform == platform) {
                    chosen = messages[i]
                    break
                }
            }
            // Fallback 1: any "all" message.
            if (!chosen) {
                for (let i=0; i < messages.length; i++) {
                    if (messages[i].platform == "all") {
                        chosen = messages[i]
                        break
                    }
                }
            }
            // Fallback 2: first message in the array, if any.
            if (!chosen && messages.length > 0) {
                chosen = messages[0]
            }
            if (chosen) {
                for (let k=0; k < chosen.text.length; k++) {
                    if (chosen.text[k].lang == lang) {
                        textLocale = chosen.text[k].str
                        break
                    } else if (chosen.text[k].lang == "en") {
                        textEn = chosen.text[k].str
                    }
                }
            }
        }

        let finalText = textLocale ? textLocale : textEn
        if (!finalText) {
            // No text to show: skip the MOTD view and go straight to main.
            console.debug("no motd text, skipping to main view")
            motd.visible = false
            motdSpacer.visible = false
            loader.source = "components/MainView.qml"
            return
        }
        motd.motdText = finalText
        motd.motdLink = link
        motd.url = getLinkURL()
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
        
        // Debug logging to trace the issue
        console.debug("loadMainViewWhenReady called");
        console.debug("hasNoProvider: " + hasNoProvider);
        console.debug("ctx: " + ctx);
        if (ctx) {
            console.debug("ctx.providers: " + ctx.providers);
            if (ctx.providers) {
                console.debug("ctx.providers.length: " + ctx.providers.length);
            }
        }

        // If no provider configured, show provider selection immediately
        // Check hasNoProvider first (from C++ context property) - this is the most reliable check
        // Use truthy check since hasNoProvider from C++ might be a variant/bool
        if (hasNoProvider) {
            console.debug("No provider configured (hasNoProvider=true), loading SwitchProvider");
            splashTimer.stop()
            loader.setSource("SwitchProvider.qml", {
                "isInitialSetup": true
            })
            return
        }

        // Fallback: check ctx.providers if ctx is available
        if (ctx && ctx.providers && ctx.providers.length === 0) {
            console.debug("No provider configured (ctx.providers is empty), loading SwitchProvider");
            splashTimer.stop()
            loader.setSource("SwitchProvider.qml", {
                "isInitialSetup": true
            })
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
        if (!motd || motd === "") {
            return true
        }
        try {
            let m = JSON.parse(motd)
            let first = m[0]
            if (first == undefined) {
                return true
            }
            return isEmpty(first.text)
        } catch (e) {
            return true
        }
    }
}
