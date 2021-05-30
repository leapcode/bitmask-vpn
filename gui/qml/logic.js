let status = 'off';
let needsReconnect = false;

function setStatus(st) {
    status = st;
}

function getStatus() { 
    return status;
}

function setNeedsReconnect(val) {
    needsReconnect = val;
}

function getNeedsReconnect() {
    return needsReconnect;
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
                    ctx.appName) // TODO failed is not handed yet
    }
}
