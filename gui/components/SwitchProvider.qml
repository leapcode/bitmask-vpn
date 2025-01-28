import QtQuick
import QtQuick.Controls
import QtQuick.Effects

ThemedPage {
    id: providerSetupPage

    property string providerName
    property bool useCircumvention: false
    property bool providerSetupInProgress: false

    signal configurationCompleted

    onConfigurationCompleted: {
        providerSetupPage.providerSetupInProgress = false;
        stackView.push("ProviderSetupComplete.qml");
        pageIndicatorAndNavigationButtonContainer.visible = false;
    }

    StackView {
        id: stackView
        anchors.fill: parent

        initialItem: ProviderSelection {}
    }

    footer: Rectangle {
        id: pageIndicatorAndNavigationButtonContainer
        color: "white"
        radius: 7
        anchors.horizontalCenter: parent.horizontalCenter

        width: 300
        height: 45
        Button {
            id: backButton
            height: 35
            width: 60
            text: {
                if (stackView.depth > 1) {
                    return "Back";
                } else {
                    return "Cancel";
                }
            }
            visible: stackView.depth > 1
            HoverHandler {
                cursorShape: Qt.PointingHandCursor
            }
            background: Rectangle {
                id: backButtonBackground
                anchors.fill: parent
                color: "pink"
                radius: 7
            }
            MultiEffect {
                source: backButtonBackground
                anchors.fill: backButtonBackground
                autoPaddingEnabled: true
                shadowEnabled: backButton.enabled
                shadowHorizontalOffset: 1
                shadowVerticalOffset: 1
                shadowColor: "gray"
                shadowBlur: 0.7
                opacity: backButton.pressed ? 0.75 : 1.0
            }
            enabled: stackView.depth > 1
            onClicked: stackView.pop()
            anchors {
                leftMargin: 4
                bottomMargin: 2
                left: parent.left
                verticalCenter: parent.verticalCenter
            }
        }

        PageIndicator {
            id: pageIndicator
            count: 4 // Total number of pages
            currentIndex: stackView.depth - 1 // Current page index (0-based)
            anchors {
                centerIn: parent
            }
        }

        Button {
            id: nextButton
            text: "Next"
            enabled: providerSetupPage.providerName !== "" && stackView.depth < 3
            onClicked: {
                if (stackView.depth === 1) {
                    stackView.push("SelectCircumventionTechnique.qml");
                } else if (stackView.depth === 2) {
                    stackView.push("ConfiguringProvider.qml");
                    providerSetupPage.providerSetupInProgress = true;
                }
            }
            HoverHandler {
                cursorShape: Qt.PointingHandCursor
            }
            background: Rectangle {
                id: nextButtonBackground
                anchors.fill: parent
                color: "pink"
                radius: 7
            }
            MultiEffect {
                source: nextButtonBackground
                anchors.fill: nextButtonBackground
                autoPaddingEnabled: true
                shadowEnabled: nextButton.enabled
                shadowHorizontalOffset: 1
                shadowVerticalOffset: 1
                shadowBlur: 0.7
                shadowColor: "gray"
                opacity: nextButton.pressed ? 0.75 : 1.0
            }
            anchors {
                rightMargin: 4
                bottomMargin: 2
                right: parent.right
                verticalCenter: parent.verticalCenter
            }
            height: 35
            width: 60
        }
    }
}
