import QtQuick
import "../themes/themes.js" as Theme

MouseArea {
    id: mouseArea

    property var targetEl: parent
    property var uiState: Theme.uiState
    property var onMouseAreaClicked: function () {
        parent.clicked()
    }

    //function changeState(stateName) {
    //    if (mouseArea.hoverEnabled)
    //       targetEl.state = stateName;
    //}
    anchors.fill: parent
    hoverEnabled: true
    cursorShape: !hoverEnabled ? Qt.ForbiddenCursor : Qt.PointingHandCursor
    //onPressed: {
    //    console.debug("button pressed")
    //changeState(uiState.statePressed)
    //}
    //onEntered: changeState(uiState.stateHovered)
    //onExited: changeState(uiState.stateDefault)
    //onCanceled: changeState(uiState.stateDefault)


    /*
    onReleased: {
        if (hoverEnabled) {
            changeState(uiState.stateDefault);
            onMouseAreaClicked();
        }
    }
*/
}
