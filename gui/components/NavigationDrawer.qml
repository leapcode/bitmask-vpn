// (c) Evgenij Legotskoj https://evileg.com/en/post/192/
// a reimplementation of Drawer to workaround a problem with overlays
// in the Qt version packaged with the snaps.
import QtQuick
import QtQuick.Window

Rectangle {
    id: panel

    function show() {
        open = true;
        drawerOn = true;
    }
    function hide() {
        open = false;
        drawerOn = false;
    }
    function toggle() {
        open = open ? false : true;
        drawerOn = open;
    }

    property bool open: false
    property int position: Qt.LeftEdge

    readonly property bool _rightEdge: position === Qt.RightEdge
    readonly property int _closeX: _rightEdge ? _rootItem.width : - panel.width
    readonly property int _openX: _rightEdge ? _rootItem.width - width : 0
    readonly property int _minimumX: _rightEdge ? _rootItem.width - panel.width : -panel.width
    readonly property int _maximumX: _rightEdge ? _rootItem.width : 0
    readonly property int _pullThreshold: panel.width/2
    readonly property int _slideDuration: 260
    readonly property int _openMarginSize: 20

    property real _velocity: 0
    property real _oldMouseX: -1

    property Item _rootItem: parent

    on_RightEdgeChanged: _setupAnchors()
    onOpenChanged: completeSlideDirection()

    width: Math.min(root.width, root.height) / 3 * 2
    height: root.height
    x: _closeX
    z: 10

    function _setupAnchors() {
        _rootItem = parent;

        shadow.anchors.right = undefined;
        shadow.anchors.left = undefined;

        mouse.anchors.left = undefined;
        mouse.anchors.right = undefined;

        if (_rightEdge) {
            mouse.anchors.right = mouse.parent.right;
            shadow.anchors.right = panel.left;
        } else {
            mouse.anchors.left = mouse.parent.left;
            shadow.anchors.left = panel.right;
        }

        slideAnimation.enabled = false;
        panel.x = _rightEdge ? _rootItem.width :  - panel.width;
        slideAnimation.enabled = true;
    }

    function completeSlideDirection() {
        if (open) {
            panel.x = _openX;
        } else {
            panel.x = _closeX;
            Qt.inputMethod.hide();
        }
    }

    function handleRelease() {
        var velocityThreshold = 5
        if ((_rightEdge && _velocity > velocityThreshold) ||
                (!_rightEdge && _velocity < -velocityThreshold)) {
            panel.open = false;
            completeSlideDirection()
        } else if ((_rightEdge && _velocity < -velocityThreshold) ||
                   (!_rightEdge && _velocity > velocityThreshold)) {
            panel.open = true;
            completeSlideDirection()
        } else if ((_rightEdge && panel.x < _openX + _pullThreshold) ||
                   (!_rightEdge && panel.x > _openX - _pullThreshold) ) {
            panel.open = true;
            panel.x = _openX;
        } else {
            panel.open = false;
            panel.x = _closeX;
        }
    }

    function handleClick(mouse) {
        if ((_rightEdge && mouse.x < panel.x ) || mouse.x > panel.width) {
            open = false;
            drawerOn = false;
        }
    }

    onPositionChanged: {
        if (!(position === Qt.RightEdge || position === Qt.LeftEdge )) {
            console.warn("SlidePanel: Unsupported position.")
        }
    }

    Behavior on x {
        id: slideAnimation
        enabled: !mouse.drag.active
        NumberAnimation {
            duration: _slideDuration
            easing.type: Easing.OutCubic
        }
    }

    NumberAnimation on x {
        id: holdAnimation
        to: _closeX + (_openMarginSize * (_rightEdge ? -1 : 1))
        running : false
        easing.type: Easing.OutCubic
        duration: 200
    }

    MouseArea {
        id: mouse
        parent: _rootItem

        y: _rootItem.y
        width: open ? _rootItem.width : _openMarginSize
        height: _rootItem.height
        onPressed:  if (!open) holdAnimation.restart();
        onClicked: handleClick(mouse)
        drag.target: panel
        drag.minimumX: _minimumX
        drag.maximumX: _maximumX
        drag.axis: Qt.Horizontal
        drag.onActiveChanged: if (active) holdAnimation.stop()
        onReleased: handleRelease()
        z: open ? 1 : 0
        onMouseXChanged: {
            _velocity = (mouse.x - _oldMouseX);
            _oldMouseX = mouse.x;
        }
    }

    Rectangle {
        id: backgroundBlackout
        parent: _rootItem
        anchors.fill: parent
        opacity: 0.5 * Math.min(1, Math.abs(panel.x - _closeX) / _rootItem.width/2)
        color: "black"
    }

    Item {
        id: shadow
        anchors.left: panel.right
        anchors.leftMargin: _rightEdge ? 0 : 10
        height: parent.height

        Rectangle {
            height: 10
            width: panel.height
            rotation: 90
            opacity: Math.min(1, Math.abs(panel.x - _closeX)/ _openMarginSize)
            transformOrigin: Item.TopLeft
            gradient: Gradient{
                GradientStop { position: _rightEdge ? 1 : 0 ; color: "#00000000"}
                GradientStop { position: _rightEdge ? 0 : 1 ; color: "#2c000000"}
            }
        }
    }
}
