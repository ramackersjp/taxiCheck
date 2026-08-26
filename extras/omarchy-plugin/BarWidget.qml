import QtQuick
import Quickshell
import qs.Ui

BarWidget {
  id: root
  moduleName: "jp.taxiprijs"

  implicitWidth: button.implicitWidth
  implicitHeight: button.implicitHeight

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
        root.bar.run("alacritty -e taxiprijs")
      }
    }
  }
}
