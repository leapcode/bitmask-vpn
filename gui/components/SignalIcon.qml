import QtQuick 2.9
import QtQuick.Controls 2.2
import QtQuick.Layouts 1.14
import Qt5Compat.GraphicalEffects

import "../themes/themes.js" as Theme

Image {
    id: icon
    height: 16
    width: 16
    // one of: good, medium, low
    property var quality: "good"

    ColorOverlay{
        anchors.fill: icon
        source: icon
        color: getQualityColor()
        antialiasing: true
    }

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
