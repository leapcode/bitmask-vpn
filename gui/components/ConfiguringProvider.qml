pragma ComponentBehavior: Bound

import QtQuick
import QtQuick.Controls
import QtQuick.Controls.Material

Item {
    id: configuringProvider
    property int setupPercentage: 0
    property int waitCount: 0

    signal startProviderSetupPercentageCount
    signal startProviderSetup

    onStartProviderSetupPercentageCount: {
        delay(1000, function () {
            if (configuringProvider.setupPercentage < 100) {
                configuringProvider.setupPercentage = configuringProvider.setupPercentage + 20;
            } else {
                timer.stop();
            }
        });
    }

    onStartProviderSetup: {
        backend.switchProvider(providerSetupPage.providerName, function () {
            timer.stop();
            configuringProvider.setupPercentage = 100;
            providerSetupPage.configurationCompleted();
        });
    }

    Rectangle {
        id: pageHeader
        color: "transparent"
        anchors.top: parent.top
        height: needCircumventionLabel.height + needCircumventionMsg.height + 5
        width: parent.width

        Label {
            id: needCircumventionLabel
            text: qsTr("Configuring provider")
            font.bold: true
            font.pixelSize: 14
            anchors.top: parent.top
            wrapMode: Text.WordWrap
            leftPadding: 20
            rightPadding: 20
            width: parent.width
            clip: false
        }

        Label {
            id: needCircumventionMsg
            /* TODO: make the "Bitmask" string interpolatable */
            text: qsTr("To connect to your provider Bitmask is fetching all the required configuration information. This only happens during first setup.")
            font.pixelSize: 12
            wrapMode: Text.WordWrap
            anchors.top: needCircumventionLabel.bottom
            anchors.topMargin: 10
            width: parent.width
            leftPadding: 20
            rightPadding: 20
            clip: false
        }

        Image {
            id: setupProgressImage
            source: "../resources/setup_progress_image.png"
            anchors.top: needCircumventionMsg.bottom
            anchors.topMargin: 10
            anchors.leftMargin: 10
            anchors.left: needCircumventionMsg.left
            height: 80
            width: 80

            // animation stops as soon as root.ctx.provider == providerSetupPage.providerName
            RotationAnimation on rotation {
                duration: 1000
                direction: RotationAnimation.Clockwise
                running: providerSetupPage.providerSetupInProgress
                from: 0
                to: 360
                loops: Animation.Infinite
            }
        }

        Rectangle {
            id: setupProgressPercentage
            height: 65
            width: 65
            anchors.centerIn: setupProgressImage
            radius: 360

            Label {
                text: qsTr(configuringProvider.setupPercentage + " %")
                anchors.centerIn: parent
                font.pixelSize: 14
            }
        }
    }

    Timer {
        id: timer
    }

    function delay(delayTime, cb) {
        timer.interval = delayTime;
        timer.repeat = true;
        timer.triggered.connect(cb);
        timer.start();
    }

    Timer {
        id: nonRepeatTimer
    }

    function delayNonRepeat(delayTime, cb) {
        nonRepeatTimer.interval = delayTime;
        nonRepeatTimer.triggered.connect(cb);
        nonRepeatTimer.repeat = false;
        nonRepeatTimer.start();
    }

    Component.onCompleted: {
        configuringProvider.startProviderSetupPercentageCount()
        delayNonRepeat(1000, function() {
            configuringProvider.startProviderSetup();
        })
    }
}
