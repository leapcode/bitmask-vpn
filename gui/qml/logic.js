var status = 'off';
var needsReconnect = false;

function setNeedsReconnect(val) {
    needsReconnect = val;
}

function getNeedsReconnect() {
    return needsReconnect;
}

function setStatus(st) {
    status = st;
}

function getStatus() { 
    return status;
}

function setNeedsDonate(val) {
    needsDonate = val;
}

function toHuman(st) {
    switch (st) {
    case "off":
        //: %1 -> application name
        return qsTr("%1 off").arg(ctx.appName)
    case "on":
        //: %1 -> application name
        return qsTr("%1 on").arg(ctx.appName)
    case "connecting":
        //: %1 -> application name
        return qsTr("Connecting to %1").arg(ctx.appName)
    case "stopping":
        //: %1 -> application name
        return qsTr("Stopping %1").arg(ctx.appName)
    case "failed":
        //: %1 -> application name
        return qsTr("%1 blocking internet").arg(
                    ctx.appName) // TODO failed is not handled yet
    }
}

// Helper to show notification messages
function showNotification(ctx, msg) {
    console.log("Going to show notification message: ", msg)
    if (supportsMessages) {
        let appname = ctx ? ctx.appName : "VPN"
        showMessage(appname, msg, null, 15000)
    } else {
        console.log("System doesn't support systray notifications")
    }
}

function shouldAllowEmptyPass(providers) {
    let obj = JSON.parse(providers.getJson())
    let active = obj['default']
    let allProviders = obj['providers']
    for (var i = 0; i < allProviders.length; i++) {
        if (allProviders[i]['name'] === active) {
            return (allProviders[i]['authEmptyPass'] === 'true')
        }
    }
    return false
}

function debugInit() {
    console.debug("Platform:", Qt.platform.os)
    console.debug("DEBUG: Pre-seeded providers:")
    console.debug(providers.getJson())
}
