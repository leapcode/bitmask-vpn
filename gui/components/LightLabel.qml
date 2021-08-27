import QtQuick 2.5
import QtQuick.Controls 2.14
import "../themes/themes.js" as Theme

Text {
    font.pixelSize: Theme.fontSize - 2
    font.family: Theme.fontFamily
    color: Theme.fontColor
    width: parent.width * 0.80
    text: parent.text

    horizontalAlignment: Text.AlignHCenter
    verticalAlignment: Text.AlignVCenter
    //lineHeightMode: Text.FixedHeight
    wrapMode: Text.Wrap

    Accessible.role: Accessible.StaticText
    Accessible.name: text
}
