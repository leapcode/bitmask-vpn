import QtQuick 2.3
import QtTest 1.0


TestCase {
    name: "SmokeTests"

    property var ctx

    function refresh() {
        ctx = JSON.parse(helper.refreshContext())
    }

    function test_helper() {
        compare(Boolean(helper), true, "does helper exist?")
    }

    function test_model() {
        compare(Boolean(jsonModel), true, "does model exist?")
    }

    function test_loadCtx() {
        refresh()
        compare(ctx.appName, "DemoLibVPN", "can read appName?")
        compare(ctx.tosURL, "https://libraryvpn.org/", "can read tosURL?")
        compare(ctx.status, "off", "is initial status off?")
    }
}
