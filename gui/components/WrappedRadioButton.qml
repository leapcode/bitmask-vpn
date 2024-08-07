import QtQuick
import QtQuick.Controls
import QtQuick.Controls.Material
import QtQuick.Controls.Material.impl

import "../themes/themes.js" as Theme

MaterialRadioButton {
   id: control
   width: parent.width
   property var location

   /* this works for the pointer, but breaks the onClick connection
      XXX need to dig into RadioButton implementation.
   MouseArea {
       anchors.fill: parent
       cursorShape: Qt.PointingHandCursor
   }
   */

   contentItem: Label {
       text: control.text
       font: control.font
       horizontalAlignment: Text.AlignLeft
       verticalAlignment: Text.AlignVCenter
       leftPadding: control.indicator.width + control.spacing
       wrapMode: Label.Wrap
   }
}
