import QtQuick 2.9
import QtQuick.Controls 2.2
import QtQuick.Controls.Material 2.12
import QtQuick.Controls.Material.impl 2.12

import "../themes/themes.js" as Theme

RadioButton {
   id: control
   width: parent.width
   property var location

   contentItem: Label {
       text: control.text
       font: control.font
       horizontalAlignment: Text.AlignLeft
       verticalAlignment: Text.AlignVCenter
       leftPadding: control.indicator.width + control.spacing
       wrapMode: Label.Wrap
   }
}
