package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	//	"time"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

const (
	radioKind       = "radio"
	vfoKind         = "vfo"
	cwKind          = "cw"
	otherKind       = "other"
	vfoAddress byte = 0x81
	cwAddress  byte = 0x80
	ack        byte = 0xff
	nack       byte = 0x00
	remTest         = 0x01
	sendMsgLen      = 8
	tutor      byte = 0x02
	keyer      byte = 0x03
	upLimit         = 1
	noRadio         = "Radio is not connected or up"
	noVFO           = "VFO is not connected or up"
	noCW            = "CW is not connected or up"
	yesRadio        = "Radio is up"
	yesVFO          = "VFO is up"
	yesCW           = "CW is up"
	auditError      = "Error report from the monitor"
	baudRate        = 38400
	setTx      byte = 0x03
	setRx      byte = 0x04
	bandCmd    byte = 0x05

	//radio commands
	id         string = "ID" //identification the reply is 0761
	idResponse string = "0761"
	km         string = "KM" //keyer memory
	varMem     string = "5"  //variable memory
	keyCW      string = "KY" //key the radio

)

var vidList = []string{"10c4", "2341"}
var noPortMatch = errors.New("no port matched the vid list")

var kinds []string = []string{radioKind, vfoKind, cwKind}

type remote struct {
	vid          string
	port         serial.Port
	portName     string
	address      byte
	up           bool
	kind         string
	serialNumber string
	lastUp       bool
	nowUp        bool
}

type remotes map[string]*remote

func (app *application) classifyRemotes() error {
	app.rem = remotes{
		radioKind: &remote{kind: radioKind},
		vfoKind:   &remote{kind: vfoKind, address: vfoAddress},
		cwKind:    &remote{kind: cwKind, address: cwAddress},
	}

	ports, err := findPorts(vidList) //returns port details
	if err != nil {
		//log.Println(err)
	}
	for _, p := range ports {
		//fmt.Println("VID: ", p.VID, "\t", "Serial No: ", p.SerialNumber)

		rdo := false
		port, err := openPort(p.Name, baudRate) //115200)
		if err != nil {
			if err.Error() == "Serial port busy" {
				fmt.Println("Busy port error", err)
				continue
			} else {
				log.Println("did not open radio port", err)
			}
		}
		if port != nil {
			rdo, err = testRadio(port)
			if err != nil {
				log.Println(err)
			}
		} else {
			//fmt.Println("nil port error", p.Name)
		}
		if rdo {
			//	fmt.Println("r is true")
			app.rem[radioKind].port = port
			app.rem[radioKind].vid = p.VID + ":" + p.PID
			app.rem[radioKind].serialNumber = p.SerialNumber
			app.rem[radioKind].portName = p.Name
			app.rem[radioKind].nowUp = true
			app.rem[radioKind].lastUp = true
			app.rem[radioKind].up = true
			rdo = false
			continue
		} else {
			//fmt.Println("r is false")
		}
		for _, kind := range kinds[1:] {
			if port != nil {
				//fmt.Printf("testing %s kind at %x\n", kind, app.rem[kind].address)
				if remoteUp(port, app.rem[kind].address) {
					//fmt.Printf("%s poassed ok\n", kind)
					app.rem[kind].port = port
					app.rem[kind].vid = p.VID + ":" + p.PID
					app.rem[kind].serialNumber = p.SerialNumber
					app.rem[kind].portName = p.Name
					app.rem[kind].nowUp = true
					app.rem[kind].lastUp = true
					app.rem[kind].up = true
				}
			} else {
				//fmt.Println("Nil port error no 2", p.Name)
			}
		}
	}
	for _, kind := range kinds {
		fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++")
		fmt.Println("Kind: ", kind)
		fmt.Println("Other Kind: ", app.rem[kind].kind)
		fmt.Println("Port Name: ", app.rem[kind].portName)
		fmt.Println("VID: ", app.rem[kind].vid)
		fmt.Println("Address: ", app.rem[kind].address)
		fmt.Println("Up? :", app.rem[kind].nowUp)
		fmt.Println("Serial Number: ", app.rem[kind].serialNumber)
		fmt.Println()
	}
	return nil

}

func findPorts(vids []string) ([]*enumerator.PortDetails, error) {
	portDetails := []*enumerator.PortDetails{}
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return portDetails, err
	}
	if len(ports) == 0 {
		return portDetails, fmt.Errorf("no ports were found")
	}
	//for _, p := range vids {
	//	vu := strings.ToUpper(v)
	//	vl := strings.ToLower(v)
	for _, port := range ports {
		if port.IsUSB {
			//			if port.VID == v || port.VID == vl || port.VID == vu {
			portDetails = append(portDetails, port)
			//}
		}
	}
	//}
	if len(portDetails) == 0 {
		return portDetails, noPortMatch
	}
	return portDetails, nil
}

func openPort(p string, b int) (serial.Port, error) {
	mode := &serial.Mode{
		BaudRate: b,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(p, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open the usb connection: %s: %v", p, err)
	}
	port.SetReadTimeout(time.Duration(150) * time.Millisecond)
	time.Sleep(time.Duration(10) * time.Millisecond)
	return port, nil
}

func remoteUp(port serial.Port, address byte) bool {
	if port == nil {
		return false
	}
	for i := 0; i < upLimit; i++ {
		err := testRemote(port, address)
		if err == nil {
			return true
		}
		time.Sleep(time.Duration(500) * time.Millisecond)
	}
	return false
}

// testRemote takes the port variable and the address of the remote device
// and sents a remTest command to the device attached to the port.  If it
// does not return an error, it means the device at the remote address
// responded with a satisfactory ack.
func testRemote(port serial.Port, address byte) error {
	wBuff := make([]byte, sendMsgLen)
	rBuff := make([]byte, 1)

	wBuff[0] = address
	wBuff[1] = remTest
	crc := crc16(wBuff[:sendMsgLen-2])
	wBuff[sendMsgLen-2] = byte(crc >> 8)
	wBuff[sendMsgLen-1] = byte(crc)

	n, err := port.Write(wBuff)
	if err != nil {
		return fmt.Errorf("failed write to usb port %v", err)
	}
	if n != sendMsgLen {
		return fmt.Errorf("did not write %d bytes, it wrote %d", sendMsgLen, n)
	}
	//	time.Sleep(time.Duration(5) * time.Millisecond)
	n, err = port.Read(rBuff)
	if err != nil {
		return fmt.Errorf("failed to read from usb port: %v", err)
	}
	if n != 1 {
		return fmt.Errorf("did not read one byte, read %d:", n)
	}
	if rBuff[0] != ack {
		return fmt.Errorf("did not get a %X in return, got %v", ack, rBuff)
	}
	return nil
}

// ------------------------  VFO  -----------------------------------
func (app *application) setFrequency(f uint32, txRx string) error {
	wBuff := make([]byte, sendMsgLen)
	rBuff := make([]byte, 1)

	wBuff[0] = vfoAddress
	switch txRx {
	case "tx":
		wBuff[1] = setTx
	case "rx":
		wBuff[1] = setRx
	default:
		wBuff[1] = setTx
	}
	wBuff[2] = byte(f >> 24)
	wBuff[3] = byte(f >> 16)
	wBuff[4] = byte(f >> 8)
	wBuff[5] = byte(f)
	crc := crc16(wBuff[:6])
	wBuff[6] = byte(crc >> 8)
	wBuff[7] = byte(crc)

	n, err := app.writeRemote(wBuff, vfoKind) //rem[vfoKind].port.Write(wBuff)
	if err != nil {
		return fmt.Errorf("failed write to usb port %v", err)
	}
	if n != 8 {
		return fmt.Errorf("did not write 8 bytes, it wrote %d", n)
	}
	n, err = app.readRemote(rBuff, vfoKind) //rem[vfoKind].port.Read(rBuff)
	if err != nil {
		return fmt.Errorf("failed to read from usb port: %v", err)
	}
	if n != 1 {
		return fmt.Errorf("did not read one byte, read %d:", n)
	}
	if rBuff[0] != ack {
		return fmt.Errorf("did not get a %X in return, got %v", ack, rBuff)
	}
	return nil
}

func (app *application) readBand() (int, error) {
	wBuff := make([]byte, sendMsgLen)
	rBuff := make([]byte, 3)
	wBuff[0] = vfoAddress
	wBuff[1] = bandCmd
	//if app.rem[vfoKind].port == nil {
	//	return 0, fmt.Errorf("Port is not open %v", wBuff)
	//}
	crc := crc16(wBuff[:sendMsgLen-2])
	wBuff[sendMsgLen-2] = byte(crc >> 8)
	wBuff[sendMsgLen-1] = byte(crc)
	n, err := app.writeRemote(wBuff, vfoKind) //rem[vfoKind].port.Write(wBuff)
	if err != nil {
		return 0, errors.Join(fmt.Errorf("failed write to vfo usb port"), err)
	}
	if n != sendMsgLen {
		return 0, fmt.Errorf("did not write %d bytes, it wrote %d", sendMsgLen, n)
	}
	time.Sleep(time.Duration(120) * time.Millisecond)
	n, err = app.readRemote(rBuff, vfoKind) //rem[vfoKind].port.Read(rBuff)
	if err != nil {
		return 0, errors.Join(fmt.Errorf("failed to read from vfo usb port"), err)
	}
	if n != 3 {
		return 0, fmt.Errorf("did not read 3 bytes from vfo, read %d:", n)
	}
	crc = crc16(rBuff[:1])
	highByte := byte(crc >> 8)
	lowByte := byte(crc)
	if highByte != rBuff[1] && lowByte != rBuff[2] {
		return 0, fmt.Errorf("CRC mismtach on read band %v\t%v\t%v", rBuff, highByte, lowByte)
	}
	return int(rBuff[0]), nil
}

//<-----------------------------  CW  -------------------------------------

func (app *application) issueCWCmd(cmd *cwData) error {
	wBuff := make([]byte, 8)
	rBuff := make([]byte, 1)

	wBuff[0] = cwAddress
	wBuff[1] = cmd.cmd
	wBuff[2] = byte(cmd.tone >> 8)
	wBuff[3] = byte(cmd.tone)
	wBuff[4] = byte(cmd.volume)
	wBuff[5] = byte(cmd.speed)
	crc := crc16(wBuff[:6])
	wBuff[6] = byte(crc >> 8)
	wBuff[7] = byte(crc)

	n, err := app.writeRemote(wBuff, cwKind)
	//n, err := app.rem[cwKind].port.Write(wBuff)
	if err != nil {
		return fmt.Errorf("failed write to usb port %v", err)
	}
	if n != 8 {
		return fmt.Errorf("did not write 8 bytes, it wrote %d", n)
	}
	n, err = app.readRemote(rBuff, cwKind)
	//n, err = app.rem[cwKind].port.Read(rBuff)
	if err != nil {
		return fmt.Errorf("failed to read from usb port: %v", err)
	}
	if n != 1 {
		return fmt.Errorf("did not read one byte, read %d:", n)
	}
	if rBuff[0] != ack {
		return fmt.Errorf("did not get a %X in return, got %v", ack, rBuff)
	}
	return nil
}

//<--------------------- READ and WRITE Remote ----------------------------

func (app *application) writeRemote(msg []byte, kind string) (int, error) {
	app.remLock.Lock()
	defer app.remLock.Unlock()
	r, ok := app.rem[kind]
	if !ok {
		return 0, fmt.Errorf("bad index into remotes with %s", kind)

	}
	if r.port == nil || !r.up {
		return 0, fmt.Errorf("remote %s is down", kind)
	}
	n, err := app.rem[kind].port.Write(msg)
	return n, err
}

func (app *application) readRemote(msg []byte, kind string) (int, error) {
	app.remLock.Lock()
	defer app.remLock.Unlock()
	r, ok := app.rem[kind]
	if !ok {
		return 0, fmt.Errorf("bad index into remotes with %s", kind)
	}
	if r.port == nil || !r.up {

		return 0, fmt.Errorf("remote %s is down", kind)

	}
	n, err := app.rem[kind].port.Read(msg)
	return n, err
}

func crc16(message []byte) uint16 {

	crc := uint16(0x0000) // Initial value
	for _, b := range message {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

//<----------------------------------  Radio ------------------------------------->

func testRadio(port serial.Port) (bool, error) {
	cmd := id + ";"
	resp := id + idResponse + ";"
	wBuff := bytesBuilder(cmd)
	tBuff := bytesBuilder(resp)
	rBuff := make([]byte, 7)

	n, err := port.Write(wBuff)
	if err != nil {
		return false, fmt.Errorf("failed write to usb port %v", err)
	}
	if n != len(wBuff) {
		return false, fmt.Errorf("did not write %d bytes, it wrote %d", len(wBuff), n)
	}
	time.Sleep(time.Duration(5) * time.Millisecond)
	n, err = port.Read(rBuff)
	if err != nil {
		return false, fmt.Errorf("failed to read from usb port: %v", err)
	}
	if n != len(rBuff) {
		return false, fmt.Errorf("did not read %d byte, read %d:", len(rBuff), n)
	}
	for i, item := range tBuff {
		if rBuff[i] != item {
			return false, nil
		}
	}
	return true, nil
}

func (app *application) initRadio() error {
	contest, err := app.otherModel.getDefault("contest")
	if err != nil {
		return err
	}
	if contest == "No" {
		return nil
	}
	for i := 0; i < 5; i++ {
		ipOne := strconv.Itoa(i + 1)
		f, err := app.otherModel.getDefault("F" + ipOne)
		if err != nil {
			if errors.Is(err, errNoRecord) {
				return nil
			}
			return err
		}
		wBuff := bytesBuilder(km + ipOne + f + "}" + ";")
		n, err := app.writeRemote(wBuff, radioKind)
		if err != nil {
			return err
		}
		if n != len(wBuff) {
			return fmt.Errorf("did not write %d, wrote %d", len(wBuff), n)
		}
	}
	return nil
}

func (app *application) tickleRadio(v *radioMsg) error {
	f, err := numFun(v.Key)
	if err != nil {
		//	if f == "0" {
		//		return nil
		//	}
		return err
	}
	fnKey, err := app.otherModel.getDefault("F" + f)
	if err != nil {
		return err
	}
	switch strings.ToUpper(fnKey) {
	case "HIS CALL":
		wBuff := bytesBuilder(km + varMem + strings.ToUpper(v.Call) + "}" + ";")
		app.writeRemote(wBuff, radioKind)
		f = varMem
	case seq:
		wBuff := bytesBuilder(km + varMem + v.Seq + "}" + ";")
		app.writeRemote(wBuff, radioKind)
		f = varMem
	default:
		if f == "5" || f == "6" || f == "7" || f == "8" || f == "9" || f == "10" {
			wBuff := bytesBuilder(km + varMem + strings.ToUpper(fnKey) + "}" + ";")
			app.writeRemote(wBuff, radioKind)
		}
	}
	stupid := map[string]string{"1": "6", "2": "7", "3": "8", "4": "9", "5": "A",
		"6": "A", "7": "A", "8": "A", "9": "A", "10": "A"}
	s := stupid[f]
	wBuff := bytesBuilder(keyCW + s + ";")
	n, err := app.writeRemote(wBuff, radioKind)
	if err != nil {
		return err
	}
	if n != len(wBuff) {
		return fmt.Errorf("did not write %d, wrote %d", len(wBuff), n)
	}
	return nil
}

func bytesBuilder(s string) []byte {
	buff := []byte{}
	for _, b := range []rune(s) {
		buff = append(buff, byte(b))
	}
	return buff
}

func numFun(n int) (string, error) {
	if n < 112 || n > 121 {
		return "", fmt.Errorf("function key number %d is out of range", n)
	}
	//if n > 116 {
	//	return "0", fmt.Errorf("function key % d not built yet", n)
	//}
	return strconv.Itoa(n - 111), nil
}
