import QtQuick
import QtQuick.Controls
import QtQuick.Layouts
import QtQuick.Effects

import "../themes/themes.js" as Theme

Item {
    id: signalIcons
    property var quality: "good"
    Image {
        id: icon
        height: 16
        width: 16
        // one of: good, medium, low

        StateGroup {
            state: quality
            states: [
                State {
                    name: "good"
                    PropertyChanges {
                        target: icon
                        source: "../resources/reception-4.svg"
                    }
                },
                State {
                    name: "medium"
                    PropertyChanges {
                        target: icon
                        source: "../resources/reception-2.svg"
                    }
                },
                State {
                    name: "low"
                    PropertyChanges {
                        target: icon
                        source: "../resources/reception-0.svg"
                    }
                }
            ]
        }
    }
    MultiEffect {
        anchors.fill: icon
        source: icon
        colorizationColor: getQualityColor()
        colorization: 1.0
        antialiasing: true
    }

    function getQualityColor() {
        // I like this better than with states
        switch (quality) {
        case "good":
            return Theme.signalGood
        case "medium":
            return Theme.signalMedium
        case "low":
            return Theme.signalLow
        default:
            return Theme.signalGood
        }
    }
}


