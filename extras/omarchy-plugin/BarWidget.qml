import QtQuick
import Quickshell
import qs.Commons
import qs.Ui

BarWidget {
  id: root
  moduleName: "jp.taxiprijs"

  implicitWidth: button.implicitWidth
  implicitHeight: button.implicitHeight

  property bool menuOpen: false

  BarIconButton {
    id: button
    anchors.fill: parent
    bar: root.bar
    text: "\uf1ba"
    fontFamily: "JetBrainsMono Nerd Font"
    tooltipText: "Open TaxiCheck"
    onPressed: function(btn) {
      if (!root.bar) return
      if (btn === Qt.LeftButton) {
        root.bar.run("alacritty --class TaxiCheck -o 'window.dimensions.columns=80' -o 'window.dimensions.lines=25' -e /home/jp/Code/taxiprijs/taxiprijs")
      } else if (btn === Qt.RightButton) {
        root.menuOpen = !root.menuOpen
      }
    }
  }

  PopupCard {
    id: menuPopup
    anchorItem: button
    owner: root
    bar: root.bar
    open: root.menuOpen
    padding: Style.space(8)
    contentWidth: Style.space(200)
    contentHeight: menuColumn.implicitHeight + Style.space(16)
    onVisibleChanged: if (!visible) root.menuOpen = false

    MouseArea {
      anchors.fill: parent
      onClicked: {}
    }

    Column {
      id: menuColumn
      anchors.fill: parent
      anchors.margins: Style.space(4)
      spacing: Style.space(2)

      Repeater {
        model: [
          { label: "Open TaxiCheck", cmd: "alacritty --class TaxiCheck -o 'window.dimensions.columns=80' -o 'window.dimensions.lines=25' -e /home/jp/Code/taxiprijs/taxiprijs" },
          { label: "Handleiding", cmd: "alacritty -e man taxiprijs" },
          { label: "Config map", cmd: "xdg-open ~/.taxiprijs" },
          { isSeparator: true },
          { label: "Quit", cmd: "killall -q taxiprijs" }
        ]

        Item {
          width: menuColumn.width
          implicitHeight: modelData.isSeparator ? Style.space(9) : Style.space(28)

          Rectangle {
            visible: !modelData.isSeparator
            anchors.fill: parent
            radius: Math.max(2, Style.cornerRadius)
            color: rowMouse.containsMouse ? Style.hoverFillFor(root.foreground, root.foreground) : "transparent"
          }

          Rectangle {
            visible: modelData.isSeparator
            anchors.left: parent.left
            anchors.leftMargin: Style.space(8)
            anchors.right: parent.right
            anchors.rightMargin: Style.space(8)
            anchors.verticalCenter: parent.verticalCenter
            height: 1
            color: Color.popups.border
            opacity: 0.5
          }

          Text {
            visible: !modelData.isSeparator
            anchors.verticalCenter: parent.verticalCenter
            anchors.left: parent.left
            anchors.leftMargin: Style.space(8)
            text: modelData.label
            color: root.foreground
            font.family: root.bar ? root.bar.fontFamily : Style.font.family
            font.pixelSize: Style.font.body
          }

          MouseArea {
            id: rowMouse
            visible: !modelData.isSeparator
            anchors.fill: parent
            hoverEnabled: true
            cursorShape: Qt.PointingHandCursor
            onClicked: {
              root.menuOpen = false
              root.bar.run(modelData.cmd)
            }
          }
        }
      }
    }
  }
}
