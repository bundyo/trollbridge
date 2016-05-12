package main

import (
	"encoding/json"
	"launchpad.net/xmlpath"
	"strings"
	"strconv"
	"errors"
	"sort"
	"fmt"
	"image"
    _ "image/gif"
    _ "image/jpeg"
    _ "image/png"
	"net/http"
	"io"
	"io/ioutil"
	"gopkg.in/qml.v1"
	"os"
	//"math/rand"
	//"os/signal"
	//"syscall"
	"runtime"
	"time"
)

// VERSION Returns the app version
const VERSION = "0.1.2"

var (
	config Config
	bridge BridgeControl
)

// File Single File record
type File struct {
	Index         string
	Path          string
	File          string
	TrollPath     string
	Type		  string
	Selected      bool
	Downloaded    bool
	Downloading   bool
	Quarter		  bool
	Size          int64
}

// Files Slice of File :)
type Files []File

func (slice Files) Len() int {
    return len(slice)
}

func (slice Files) Less(i, j int) bool {
    return slice[i].Index < slice[j].Index;
}

func (slice Files) Swap(i, j int) {
    slice[i], slice[j] = slice[j], slice[i]
}

// BridgeControl Main application struct
type BridgeControl struct {
	Root         qml.Object
	Model        string
	list         Files
	FileLen      int
	ticker       *time.Ticker
	Connected    bool
	Downloading  bool
	OPC			 bool
	//socket       Socket
}

// Config is config
type Config struct {
	DownloadPath string
	LastRun      time.Time
}

func main() {
	if err := qml.SailfishRun(run); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func getPath() (string, error) {
	path := os.Getenv("XDG_CONFIG_HOME")
	if len(path) == 0 {
		path = os.Getenv("HOME")
		if len(path) == 0 {
			return "", errors.New("No XDG_CONFIG or HOME env set!")
		}
	}
	return path, nil
}

// Load JSON serialized config data
func loadSettings() error {
	path, err := getPath()
	if err != nil {
		panic(err)
	}
	filename := fmt.Sprintf("%s/.config/harbour-trollbridge/settings_%s.json", path, VERSION)
	fmt.Printf("Trying to load settings from %s\n", filename)
	f, err := os.Open(filename)
	if err != nil {
		config = Config{LastRun: time.Now(), DownloadPath: "/home/nemo/Pictures/Olympus"}
		return nil
	}
	defer f.Close()
	jsondec := json.NewDecoder(f)
	err = jsondec.Decode(&config)
	if err != nil {
		return err
	}
	return nil
}

// Save serialized config data as JSON
func saveSettings() error {
	path, err := getPath()
	if err != nil {
		panic(err)
	}

	directory := fmt.Sprintf("%s/.config/harbour-trollbridge", path)
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.MkdirAll(directory, 0777)
	}
	filename := fmt.Sprintf("%s/settings_%s.json", directory, VERSION)
	f, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Can not create settings: %v\n", err)
		return errors.New("Can not create settings file!")
	}
	defer f.Close()

	jsondata, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Println("Can not marshal to json")
		return errors.New("Can not marshal to json!")
	}
	f.Write(jsondata)
	return nil
}

func run() error {
	err := loadSettings()
	if err != nil {
		fmt.Printf("Error %s, stopping execution.", err)
		os.Exit(-2)
	}
	
	engine := qml.SailfishNewEngine()
	//engine.Translator("/usr/share/harbour-trollbridge/qml/i18n")
	
	engine.AddImageProvider("troll", func(id string, width, height int) image.Image {
		if bridge.Connected {
			img, _ := bridge.CameraGetFile("/" + id)
			
			if img != nil {
				return img
			}
		}
		
		return bridge.ReadDefaultImage()
	})

	bridge = BridgeControl{Model: "", Connected: false, Downloading: false, ticker: time.NewTicker(5 * time.Second)}
	
	go func() {
		for t := range bridge.ticker.C {
			if (!bridge.Downloading) {
				bridge.Connect()
				fmt.Println("Connect at", t)
			}
		}
	}()

	context := engine.Context()
	context.SetVar("bridge", &bridge)
	controls, err := engine.SailfishSetSource("qml/main.qml")
	if err != nil {
		return err
	}

	window := controls.SailfishCreateWindow()
	bridge.Root = window.Root()

	err = loadSettings()
	if err != nil {
		fmt.Printf("Error %s, stopping execution.", err)
		os.Exit(-2)
	}
	
	window.SailfishShow()
	window.Wait()

    // sigs := make(chan os.Signal, 1)
    // done := make(chan bool, 1)

    // signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

    // go func() {
    //     sig := <-sigs
    //     fmt.Println()
    //     fmt.Println(sig)
    //     done <- true
    // }()

    // <-done
    // fmt.Println("exiting")

	err = saveSettings()
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

// Connect Connects to the Camera
func (ctrl *BridgeControl) Connect() {
	go func() {
		model, err := ctrl.CameraGetValue("get_caminfo", "/caminfo/model")
		cameraType, _ := ctrl.CameraGetValue("get_connectmode", "/connectmode")

		if (cameraType == "OPC") {
			ctrl.CameraExecute("switch_commpath", "path=wifi")
			ctrl.OPC = true

			if !ctrl.Connected && err != nil {
				ctrl.SwitchMode("standalone")
			}
						
			if model == "" {
				model, err = ctrl.CameraGetValue("get_caminfo", "/caminfo/model")
			}
		}
		
		qml.Changed(ctrl, &ctrl.OPC)
		ctrl.Connected = err == nil
		ctrl.SetModel(model)
    }()
}

// ReadDefaultImage Returns the default image
func (ctrl *BridgeControl) ReadDefaultImage() image.Image {
	f, _ := os.Open("/usr/share/icons/hicolor/86x86/apps/harbour-trollbridge.png")
	defer f.Close()
			
	img, _, _ := image.Decode(f)

	return img
}

// SwitchState Switch the camera on or off
func (ctrl *BridgeControl) SwitchState(on bool) {
	go func ()  {
		if on {
			ctrl.CameraExecute("exec_pwon", "")
		} else {
			ctrl.CameraExecute("exec_pwoff", "")
		}
	
		time.AfterFunc(100 * time.Millisecond, ctrl.Connect)
	}()
}

// GetImage Get image at list index
func (ctrl *BridgeControl) GetImage(index int) *File {
	return &ctrl.list[index]
}

// SetSelection Set selection at list index
func (ctrl *BridgeControl) SetSelection(index string, value bool) {
	for idx := range ctrl.list {
		if ctrl.list[idx].Index == index {
			ctrl.SetSelectionItem(idx, value)
			return;
		}
	}
}

// SetSelectionItem Set selection at list index
func (ctrl *BridgeControl) SetSelectionItem(idx int, value bool) {
	ctrl.list[idx].Selected = value
	
	bridge.UpdateItem(idx)
}

// ClearAllSelection Clears the file list selection
func (ctrl *BridgeControl) ClearAllSelection() {
	go func() {
		for idx := range ctrl.list {
			ctrl.SetSelectionItem(idx, false)
		}
	}()
}

// Download Downloads the file at index
func (ctrl *BridgeControl) Download(idx int, quarterSize bool) {
	size := bridge.CameraDownloadFile(ctrl.list[idx].Path, ctrl.list[idx].File, quarterSize)
	
	ctrl.list[idx].Downloaded = size > -1
	ctrl.list[idx].Downloading = false
	ctrl.list[idx].Selected = false
	ctrl.list[idx].Quarter = quarterSize
	
	bridge.UpdateItem(idx)
}

// DownloadSelected Downloads all selected files
func (ctrl *BridgeControl) DownloadSelected(quarterSize bool) {
	go func() {
		ctrl.Downloading = true
		qml.Changed(ctrl, &ctrl.Downloading)
		
		for idx := range ctrl.list {
			if (ctrl.list[idx].Selected) {
				ctrl.list[idx].Downloading = true

				bridge.UpdateItem(idx)
			}
		}

		for idx := range ctrl.list {
			if (ctrl.list[idx].Selected) {
				ctrl.Download(idx, quarterSize)
			}
		}

		ctrl.Downloading = false
		qml.Changed(ctrl, &ctrl.Downloading)
	}()
}

// UpdateItem Downloads the file at index
func (ctrl *BridgeControl) UpdateItem(idx int) {
	photoModel := bridge.Root.ObjectByName("photoModel")
	photoModel.Set("item", ctrl.list[idx])
	photoModel.Set("setIndex", ctrl.list[idx].Index)
}

// // BindToEvents Bind to cmera events
// func (ctrl *BridgeControl) BindToEvents() {
// 	rand.Seed(time.Now().UnixNano())
	
// 	port := rand.Intn(5000) + 50000
	
// 	ctrl.CameraExecute("start_pushevent", "port=" + port)
	
// //        self.eventSocket = socket.socket(socket.AF_INET,socket.SOCK_STREAM)
// //        self.eventSocket.setsockopt(socket.SOL_SOCKET,socket.SO_REUSEADDR,1)
// //        self.eventSocket.settimeout(2)
// //        self.eventSocket.connect((OlympusAir.IP,evPort))

//   conn, _ := net.Dial("tcp", "192.168.0.10:" + port)
//   for {
//     // read in input from stdin
//     reader := bufio.NewReader(os.Stdin)
//     //fmt.Print("Text to send: ")
//     //text, _ := reader.ReadString('\n')
//     // send to socket
//     //fmt.Fprintf(conn, text + "\n")
//     // listen for reply
//     message, _ := bufio.NewReader(conn).ReadString('\n')
//     fmt.Print("Message from server: " + message)
//   }
// }

// }

// SwitchMode Switch the camera mode to rec/play/shutter
func (ctrl *BridgeControl) SwitchMode(mode string) {
	go func ()  {
		if (ctrl.OPC) {
			if mode == "shutter" {
				mode = "rec"
			}
			
			ctrl.CameraExecute("switch_cameramode", "mode=" + mode)
			return 
		}
		
		ctrl.CameraExecute("switch_cammode", "mode=" + mode)
	}()
}

// ShutterToggle Toggle the remote shutter
func (ctrl *BridgeControl) ShutterToggle(press bool) {
	go func ()  {
		if press {
			if (ctrl.OPC) {
				ctrl.CameraExecute("exec_takemotion", "com=newstarttake")
				return
			}
			
			ctrl.CameraExecute("exec_shutter", "com=1st2ndpush")
			return
		}

		if (ctrl.OPC) {
			ctrl.CameraExecute("exec_takemotion", "com=newstoptake")
			return
		}

		ctrl.CameraExecute("exec_shutter", "com=2nd1strelease")
	}()
}

// HalfWayToggle Toggle remote focusing
func (ctrl *BridgeControl) HalfWayToggle(press bool) {
	go func ()  {
		if press {
			ctrl.CameraExecute("exec_shutter", "com=1stpush")
		} else {
			ctrl.CameraExecute("exec_shutter", "com=1strelease")
		}
	}()
}

// GetFileList Check for files
func (ctrl *BridgeControl) GetFileList() {
	go func ()  {
		ctrl.CameraGetFolder("/DCIM/100OLYMP")
	}()
}

// SetModel BridgeControl Model setter 
func (ctrl *BridgeControl) SetModel(model string) {
	ctrl.Model = model
	fmt.Println("changed: " + ctrl.Model + " " + map[bool] string {true: "true", false: "false"} [ctrl.Connected])
	qml.Changed(ctrl, &ctrl.Model)
	qml.Changed(ctrl, &ctrl.Connected)
}

func (ctrl *BridgeControl) fireQuery(requestType string, query string, params []string) (*http.Response, error) {
	var (
		client *http.Client
		req    *http.Request
	)
	
	paramString := ""
	
	if len(params) > 0 {
		paramString = "?" + strings.Join(params[:], "&")
	}

	// Shorten the delay for camera detection
	if query == "get_caminfo" {
		client = &http.Client{
			Timeout: time.Duration(2 * time.Second),
		}
	} else {
		client = &http.Client{}
	}
	
	if requestType == "" {
		requestType = "GET"
	}

	if requestType == "file" {
		req, _ = http.NewRequest("GET", "http://192.168.0.10/" + query + paramString, nil)
	} else {
		req, _ = http.NewRequest(requestType, "http://192.168.0.10/" + query + ".cgi" + paramString, nil)
	}
	req.Header.Add("User-Agent", "OlympusCameraKit")
	req.Header.Add("Host", "192.168.0.10")
	resp, err := client.Do(req)
			
	if err != nil {
 		fmt.Println("Error: " + err.Error())
 		return nil, err
 	}
	
	return resp, nil
}

// CameraGetValue Get a value from camera
func (ctrl *BridgeControl) CameraGetValue(query string, path string, params ...string) (string, error) {
	resp, err := ctrl.fireQuery("", query, params)
	
	if err != nil {
		return "", err
	}
	
	defer resp.Body.Close()
	
	xpath := xmlpath.MustCompile(path)
	root, err := xmlpath.Parse(resp.Body)
	
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return "", err
	}
	
	if value, ok := xpath.String(root); ok {
		fmt.Println("returning " + value)
		return value, nil
	}
		
	return "", err
}

// CameraGetFolder Get file list from camera
func (ctrl *BridgeControl) CameraGetFolder(path string) error {
	resp, err := ctrl.fireQuery("", "get_imglist", []string{ "DIR=" + path })
	
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return err
	}

	defer resp.Body.Close()
	
	if resp.Header.Get("Content-Type") == "text/plain" {
		d, err := ioutil.ReadAll(resp.Body)
		data := string(d)
		
		if err != nil {
			fmt.Println("Error: " + err.Error())
			return err
		}
		
		if strings.HasPrefix(data, "VER_100") {
			var rowData []string

			data = strings.TrimSpace(strings.TrimLeft(data, "VER_100"))
			rows := strings.Split(data, "\r\n")
			
			if len(ctrl.list) > 0 {
				rowData = strings.Split(rows[len(rows)-1:][0], ",")
				dotIndex := strings.LastIndex(rowData[1], ".") + 1
				index := rowData[1][4:8] + map[bool] string {true: "", false: rowData[1][dotIndex:]} [dotIndex == 0]
				
				if err != nil {}
				
				if index == ctrl.list[0].Index {
					for _, row := range ctrl.list {
						bridge.Root.ObjectByName("photoModel").Set("addItem", row)
					}

					return nil
				}
				
				ctrl.list = nil
				bridge.Root.ObjectByName("photoModel").Call("clear")
			}
			
			go func() {
				for _, row := range rows {
					rowData = strings.Split(row, ",")

					if len(rowData) > 0 {
						size, err := strconv.ParseInt(rowData[2], 10, 64)
						stat, err := os.Stat(config.DownloadPath + "/" + ctrl.Model + "/" + rowData[1])
						dotIndex := strings.LastIndex(rowData[1], ".") + 1
						fileType := map[bool] string {true: "", false: rowData[1][dotIndex:]} [dotIndex == 0]
						
						ctrl.list = append(ctrl.list, 
							File { 
								Index: rowData[1][4:8] + fileType, 
								Path: rowData[0], 
								File: rowData[1],
								TrollPath: "image://troll" + rowData[0] + "/" + rowData[1],
								Type: fileType,
								Size: size,
								Downloading: false,
								Selected: false,
								Downloaded: err == nil,
								Quarter: err == nil && stat.Size() != size,
							})
					}
				}
				
				sort.Sort(sort.Reverse(ctrl.list))
				
				for _, row := range ctrl.list {
					bridge.Root.ObjectByName("photoModel").Set("addItem", row)
				}

				//fmt.Println(ctrl.list)
				ctrl.FileLen = len(ctrl.list)
				qml.Changed(ctrl, &ctrl.FileLen)
			}()
		}
	}

	return nil
}

// CameraGetFile Gets a file from camera
func (ctrl *BridgeControl) CameraGetFile(file string) (image.Image, error) {
	resp, err := ctrl.fireQuery("", "get_thumbnail", []string{ "DIR=" + file })
	
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return nil, err
	}
	
	defer resp.Body.Close()
	
	//fmt.Println(resp.Header)
	
	img, _, err := image.Decode(resp.Body)

	if err != nil {
		fmt.Println("Error: " + err.Error())
		return nil, err
	}

	return img, nil
}

// CameraDownloadFile Download a file from the camera
func (ctrl *BridgeControl) CameraDownloadFile(path string, file string, quarterSize bool) int64 { 
	downloadPath := config.DownloadPath + "/" + ctrl.Model
	
	stat, err := os.Stat(downloadPath)
	if os.IsNotExist(err) {
		os.MkdirAll(downloadPath, 0777)
	}
	
	if stat != nil && !stat.IsDir() {
		fmt.Println("The download path " + downloadPath + " is not a folder")
		return -1
	}
	
	output, err := os.Create(downloadPath + "/" + file)
	
	if err != nil {
		fmt.Println("Error while creating", file, "-", err)
		return -1
	}
	
	defer output.Close()
	
	var resp *http.Response
	
	if (quarterSize) {
		resp, err = ctrl.fireQuery("", "get_resizeimg", []string{ "DIR=" + path + "/" + file, "size=2048" })
	} else {
		resp, err = ctrl.fireQuery("file", path[1:] + "/" + file, nil)
	}
	
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return -1
	}
	
	defer resp.Body.Close()
	
	size, err := io.Copy(output, resp.Body)
	
	if err != nil {
		fmt.Println("Error while downloading", path + "/" + file, "-", err)
		return -1
	}

	fmt.Println(size, "bytes downloaded.")
	
	return size
}

// CameraExecute Fire GET request to camera
func (ctrl *BridgeControl) CameraExecute(query string, params ...string) (string, error) {
	resp, err := ctrl.fireQuery("", query, params)
	
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return "", err
	}

	defer resp.Body.Close()
	
	//htmlData, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(query, params)
	//fmt.Println(string(htmlData), resp.Header)
	return "", nil
}

// RuntimeVersion Returns the GO runtime version used for building the application
func (ctrl *BridgeControl) RuntimeVersion() string {
	return runtime.Version()
}

// Version Returns the dewpoint calculator application version
func (ctrl *BridgeControl) Version() string {
	return VERSION
}
