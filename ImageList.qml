import QtQuick 2.0
import Sailfish.Silica 1.0

Page {
    id: imageList
    property string mode: ""
    
    onStatusChanged: {
        if (status == PageStatus.Deactivating && _navigation == PageNavigation.Back) {
            if (bridge.OPC) {
                bridge.switchMode("standalone");
            }
        }
    }
    
    ListModel {
        id: photoModel
        objectName: "photoModel"
        
        property var addItem: null
        onAddItemChanged: { 
            if (addItem !== null) {
                this.append(addItem);
            }
        }
        
        property var item: null
        property var setIndex: null
        onSetIndexChanged: {
            if (setIndex !== null && item !== null) {
                console.log("setIndex: " + item.index, item.downloading, item.downloaded)
                this.set(findInModel(setIndex), item);
                item = null;
                setIndex = null;
            }
        }
        
        function findInModel(idx) {
            for (var i = 0, len = photoModel.count; i < len; i++) {
                if (photoModel.get(i).index === idx) {
                    return i;
                }
            }
        }
    }
    
    SilicaGridView {
        id: thumbGridView

        anchors.fill: parent
        cellWidth: parent.width / Math.round(parent.width / 200)
        cellHeight: cellWidth
        model: photoModel
        height: parent.height
        
        PullDownMenu {
            id: pdMain
            
            property string action: "" 
            property var subAction: null
            enabled: !bridge.downloading
            
            MenuItem {
                id: pullDownDeleteShow
                
                visible: false
                text: qsTr("Select and DELETE!")
                onClicked: {
                    pdMain.action = "pullDownDelete";
                }
            }

            MenuItem {
                id: pullDownDownloadShow
                
                text: qsTr("Select and Download")
                onClicked: {
                    pdMain.action = "pullDownDownload";
                }
            }
            
            MenuItem {
                id: pullDownDelete
                visible: false

                text: qsTr("DELETE All Selected!")
                onClicked: {
                    bridge.clearAllSelection();
                    pdMain.action = "main";
                }
            }

            MenuItem {
                id: pullDownDownloadHalf
                visible: false

                text: qsTr("Download Selected Quarter Sized")
                onClicked: {
                    pdMain.action = "main";
                    pdMain.subAction = function () {
                        bridge.downloadSelected(true);
                    } 
                }
            }

            MenuItem {
                id: pullDownDownload
                visible: false

                text: qsTr("Download Selected")
                onClicked: {
                    pdMain.action = "main";
                    pdMain.subAction = function () {
                        bridge.downloadSelected(false);
                    } 
                }
            }
            
            onActiveChanged: {
                if (active === false && pdMain.action !== "") {
                    //pullDownDeleteShow.visible = pdMain.action === "main";
                    pullDownDownloadShow.visible = pdMain.action === "main";
                    pullDownDelete.visible = pdMain.action === "pullDownDelete";
                    pullDownDownload.visible = pdMain.action === "pullDownDownload";
                    pullDownDownloadHalf.visible = pdMain.action === "pullDownDownload";
                    imageList.mode = pdMain.action === "main" ? "" : pdMain.action.replace("pullDown", "");
                    pdMain.action = "";
                    
                    if (pdMain.subAction) {
                        pdMain.subAction();
                        thumbGridView.forceLayout();
                    }
                }
            }
        }

        header: Column {
            id: customHeader
        
            width: ( parent.width - ( 2.0 * Theme.paddingLarge ))
            height: 110
            spacing: 0
            x: Theme.paddingLarge
            
            // Aargh!
            Rectangle { height: Theme.paddingMedium; width: parent.width; color: "transparent" }
        
            Label {
                id: firstHeader
                width: parent.width
                height: Theme.fontSizeExtraLarge
        
                text: (imageList.mode !== "" ? imageList.mode : "Images")
                font.pixelSize: Theme.fontSizeLarge
                color: Theme.highlightColor
                horizontalAlignment: Text.AlignRight
            }
        
            Label {
                id: subHeader
                width: parent.width
        
                text: (imageList.mode !== "" ? "from " : "on ") + bridge.model
                font.pixelSize: Theme.fontSizeMedium
                color: Theme.secondaryHighlightColor
                horizontalAlignment: Text.AlignRight
            }
        }
        
        delegate: Item {
            id: thumbDelegate
            
            width: parent.width / Math.round(parent.width / 200)
            height: width

            Image {
                id: thumbImage
                anchors.fill: parent
                asynchronous: true
                cache: true
                smooth: !thumbDelegate.GridView.view.moving
                fillMode: Image.PreserveAspectCrop
                source: trollPath
            }

            Loader {
                anchors.centerIn: parent
                visible: downloading || thumbImage.status == Image.Loading
                sourceComponent: thumbBusy

                Component { 
                    id: thumbBusy 
                    
                    BusyIndicator { 
                        running: true 
                        visible: !downloading
                    } 
                }
                
                Image {
                    width: parent.width - 4
                    height: width
                    anchors.centerIn: parent
                    visible: downloading
                    source: "image://theme/icon-m-sync"
                    
                    RotationAnimation on rotation {
                        loops: Animation.Infinite
                        duration: 1000
                        from: 0
                        to: 360
                    }
                }
            }
            
            Rectangle {
                id: checkboxDownload
                width: 48
                height: width
                color: "#FFF"
                radius: width
                visible: selected && imageList.mode === "Download"
                anchors { right: parent.right; top: parent.top; margins: 5 }

                Image {
                    width: parent.width - 4
                    height: width
                    anchors.centerIn: parent
                    source: "image://theme/icon-lock-application-update"
                }
            }
            
            Rectangle {
                id: checkboxDelete
                width: 48
                height: width
                color: "red"
                radius: width
                visible: selected && imageList.mode === "Delete"
                anchors { right: parent.right; top: parent.top; margins: 5 }
                border.color: "white"
                border.width: 2

                Image {
                    width: parent.width - 4
                    height: width
                    anchors.centerIn: parent
                    source: "image://theme/icon-cover-cancel?#FFF"
                }
            }

            Rectangle {
                id: checkboxDownloaded
                width: 48
                height: width
                color: "#369"
                radius: width
                visible: downloaded
                anchors { right: parent.right; bottom: parent.bottom; margins: 5 }
                border.color: "white"
                border.width: 2

                Image {
                    width: parent.width - 4
                    height: width
                    anchors.centerIn: parent
                    source: "image://theme/" + (quarter ? "icon-m-scale" : "icon-m-tabs") + "?#FFF"
                }
            }
            
            Rectangle {
                id: checkboxType
                width: 48
                height: width
                color: type === "JPG" ? "darkorange" : "tomato"
                radius: width
                visible: type
                anchors { left: parent.left; bottom: parent.bottom; margins: 5 }
                border.color: "white"
                border.width: 2

                Label {
                    width: parent.width - 2
                    height: width
                    color: "white"
                    font.bold: true
                    font.pixelSize: Theme.fontSizeTiny
                    verticalAlignment: Text.AlignVCenter
                    horizontalAlignment: Text.AlignHCenter
                    text: type.toUpperCase()
                }
            }

            Rectangle {
                id: selectedIndicator
                anchors.fill: parent
                color: "transparent"
                border.color: "#333"
                border.width: 1
            }

            MouseArea {
                anchors.fill: parent
                onClicked: {
                    if (imageList.mode !== "") {
                        bridge.setSelection(index, !selected);
                    }
                }
            }
        }
    }
}