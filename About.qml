import QtQuick 2.0
import Sailfish.Silica 1.0

Page {
	id: aboutpage
	SilicaFlickable {
		anchors.fill: parent
		contentWidth: parent.width
		contentHeight: col.height + Theme.paddingLarge

		VerticalScrollDecorator {}

		Column {
			id: col
			spacing: Theme.paddingLarge
			width: parent.width
			PageHeader {
				title: qsTr("About Troll Bridge")
			}

			Image {
				anchors.horizontalCenter: parent.horizontalCenter
				source: "/usr/share/icons/hicolor/86x86/apps/harbour-trollbridge.png"
			}

			SectionHeader {
				text: qsTr("Information")
			}
			Label {
				text: qsTr("TRaveller's OLympus Bridge is an Olympus cameras control application<br><br>\
				Troll Bridge supports both OM-D/PEN WiFi cameras and Olympus Air.<br>\
				The name of the application was chosen in memory of Terry Pratchett,<br>\
				who died on March 12th, 2015.<br>\
				This application has been build using GO language and QML bindings.<br>\
				(C)2016 Bundyo, released under the MIT license.")
				anchors.horizontalCenter: parent.horizontalCenter
				wrapMode: Text.WrapAtWordBoundaryOrAnywhere
				width: (parent ? parent.width : Screen.width) - Theme.paddingLarge * 2
				verticalAlignment: Text.AlignVCenter
				horizontalAlignment: Text.AlignLeft
				x: Theme.paddingLarge
			}
			SectionHeader {
				text: qsTr("Additional Copyright")
			}

			Label {
				text: qsTr("<a href='https://together.jolla.com/question/105098/how-to-setup-go-142-15-16-runtime-and-go-qml-pkg-for-mersdk/'>GO-QML port to sailfish OS</a> (C) Nekron.")
				anchors.horizontalCenter: parent.horizontalCenter
				wrapMode: Text.WrapAtWordBoundaryOrAnywhere
				width: (parent ? parent.width : Screen.width) - Theme.paddingLarge * 2
				verticalAlignment: Text.AlignVCenter
				horizontalAlignment: Text.AlignLeft
				linkColor: "lightsteelblue"
				x: Theme.paddingLarge
			}
			Label {
				text: qsTr("<a href='https://github.com/go-qml/qml'>GO-QML package</a> (C) Gustavo Niemeyer.")
				anchors.horizontalCenter: parent.horizontalCenter
				wrapMode: Text.WrapAtWordBoundaryOrAnywhere
				width: (parent ? parent.width : Screen.width) - Theme.paddingLarge * 2
				verticalAlignment: Text.AlignVCenter
				horizontalAlignment: Text.AlignLeft
				linkColor: "lightsteelblue"
				x: Theme.paddingLarge
			}
			Label {
				text: qsTr("<a href='https://golang.org/'>GO</a> Copyright (c) 2012 The Go Authors. All rights reserved.")
				anchors.horizontalCenter: parent.horizontalCenter
				wrapMode: Text.WrapAtWordBoundaryOrAnywhere
				width: (parent ? parent.width : Screen.width) - Theme.paddingLarge * 2
				verticalAlignment: Text.AlignVCenter
				horizontalAlignment: Text.AlignLeft
				linkColor: "lightsteelblue"
				x: Theme.paddingLarge
			}

			Label {
				text: qsTr("Compiled using GO Runtime %1<br>Application version %2").arg(bridge.runtimeVersion()).arg(bridge.version())
				anchors.horizontalCenter: parent.horizontalCenter
				wrapMode: Text.WrapAtWordBoundaryOrAnywhere
				width: (parent ? parent.width : Screen.width) - Theme.paddingLarge * 2
				verticalAlignment: Text.AlignVCenter
				horizontalAlignment: Text.AlignLeft
				x: Theme.paddingLarge
			}
		}
	}
}

