import QtQuick 2.0
import QtQuick.Dialogs 1.2

MessageDialog {
    standardButtons: StandardButton.No | StandardButton.Yes
    title: qsTr("Donate")
    icon: StandardIcon.Warning
    text: getText()

    function getText() {
        var _name = ctx ? ctx.appName : "vpn"
	//: donate dialog
	//: %1 -> application name
        var _txt = qsTr(
            "The %1 service is expensive to run. Because we don't want to store personal information about you, there are no accounts or billing for this service. But if you want the service to continue, donate at least $5 each month.\n\nDo you want to donate now?").arg(_name)
        return _txt
    }

    onAccepted: {
        if (backend) {
            Qt.openUrlExternally(ctx.donateURL)
            backend.donateAccepted()
        }
    }
}
