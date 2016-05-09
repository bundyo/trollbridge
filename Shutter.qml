import QtQuick 2.0
import Sailfish.Silica 1.0

Page {
    id: shutterPage
    
    onStatusChanged: {
        if (status == PageStatus.Deactivating && _navigation == PageNavigation.Back) {
             bridge.switchMode(bridge.opc ? "standalone" : "play")
        }
    }

	SilicaFlickable {
		anchors.fill: parent
		contentHeight: column.height + Theme.paddingLarge

		// PullDownMenu {
		// 	id: pullDownMenu
		// 	MenuItem {
		// 		text: qsTr("")
		// 		onClicked: pageStack.push(Qt.resolvedUrl(""))
		// 	}
		// }

		VerticalScrollDecorator {}
		Column {
			id: column
			spacing: Theme.horizontalPageMargin
			width: parent.width
			PageHeader { title: qsTr("Remote Shutter") }
            
            Button {
                text: "Shutter"
                anchors.horizontalCenter: parent.horizontalCenter
                onPressedChanged: bridge.shutterToggle(pressed)
            }

            Button {
                text: "Half-way"
                anchors.horizontalCenter: parent.horizontalCenter
                onPressedChanged: bridge.halfWayToggle(pressed)
            }
        }
    }
}