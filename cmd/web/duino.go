package main

/*
My plan is to move from Raspberry Pi to Mac Mini.  That will allow me to share data with
all my other Macs using iCloud.  My initial implementation was using the Raspberry Pi
GPIO pins and the Gobot Raspi library.  Since then, I have built other solutons using
the go.bug.st/serial cross platform library.  My initial idea was to use USB, so it is
time to switch.

The functions in this (duino.go) scan the ports and identify the ports matching the
specified vendor IDs (I might have to augment this search with product ID also in
the future).  Once the ports are identified and opened, they are tested to see if one
of the required peripherals are attached to them (currently cw and vfo, in the future
500 watt tuner and 500 watt linear).

Once a port is opened and identified, it is added to the structure "remotes" that is
pointed to by app.rem field.

It is neccessary to allow for random plugging in and powering of the various peripherals.
Hence these functions will be constructed such that they can be called at sttartup and
also while this applicatio is already running.  Furthermore, before accessing the port
or the remote, their proper state will be tested.

Functions writeRemote and readRemote wrap the port.Write and port.Read to allow for the
random powering up, powering down of the remotes and starting the application in any
order.

All messages are specified to be 8 bytes since I have not figured out how to read
variable length messages on the Arduino side.
*/
import (
	"fmt"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

const (
	upLimit         = 1
	remTest    byte = 0x01
	ack        byte = 0xff
	cwKind          = "cw"
	vfoKind         = "vfo"
	portBusy        = "Serial port busy"
	sendMsgLen      = 8
)

type remote struct {
	vid      string
	port     serial.Port
	portName string
	address  byte
	up       bool
	kind     string
}

type remotes map[string]*remote

var remoteKind = map[string]byte{cwKind: cwAddress, vfoKind: vfoAddress}
var vidList = []string{"10c4"}

// classifyPorts takes a list of founded and opened port names and
// returns the structure remotes ready for communicating with remotes.
// the key to remotes is the port type and the value is the structure
// remote.  In that way, each functon (cw, vfo, etc.) can pick out the
// remote info they need.
//
//Variable app.rem is initialized by calling make.  It is a map of
//of pointers to structurs each modeling one remote peripheral (e.g. cw).
//to be able to call this function anytime, the first thing we will do
//is to check and see if a given peripheral pointer exists.

//to do that, first we check to see if a given kind is in the map.
//if it is not, the implication is that this port has not been discovered.

func (app *application) classifyRemotes() error {

	mode := &serial.Mode{
		BaudRate: 115200,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	portNames, err := findPorts(vidList)
	if err != nil {
		return err
	}

	//remote kind is a map with key being the kind (cw, vfo etc.) and value
	for kind, address := range remoteKind {
		r, ok := app.rem[kind]
		if !ok {
			for _, portName := range portNames {
				//Assuming we can open an already open port
				port, err := serial.Open(portName, mode)
				if err != nil {
					//log.Printf("failed to open the usb connection: %s: %v\n", portName, err)
				}
				port.SetReadTimeout(time.Duration(2) * time.Second)
				if remoteUp(port, address) {
					//app.remLock.Lock()
					app.rem[kind] = &remote{
						//		vid:      vid, //todo what should this be
						port:     port,
						portName: portName,
						address:  address,
						up:       true,
						kind:     kind,
					}
					//app.remLock.Unlock()
				} else {
					port.Close()
					log.Printf("in no struct, port %s of type %s was not up or attached", portName, kind)
				}
			}
		} else {
			if r.port == nil {
				//assuming we can open an already open port
				for _, portName := range portNames {
					port, err := serial.Open(portName, mode)
					if err != nil {
						//for now, just log errors like this
						log.Printf("failed to open the usb connection: %s: %v\n", portName, err)
					}
					port.SetReadTimeout(time.Duration(2) * time.Second)
					if remoteUp(port, address) {
						//app.remLock.Lock()
						app.rem[kind] = &remote{
							//		vid:      vid, //todo what should this be
							port:     port,
							portName: portName,
							address:  address,
							up:       true,
							kind:     kind,
						}
						//app.remLock.Unlock()
					} else {
						port.Close()
						log.Printf("in port nil, port %s of type %s was not up or attached", portName, kind)
					}

				}
			} else {
				if remoteUp(r.port, address) {
					app.rem[kind].up = true
					log.Printf("Remote %s at %v is up\n", kind, address)
					return nil
				}
			}
			log.Println("Length of remotes", len(app.rem))
			return nil
		}
	}
	return nil
}

// findPorts takes in a list of vendor IDs and returns a list of valid port names
// the key of the returned map is the vendor ID and the value is the port name.
// vendor ID in the map is likely to have future uses.
func findPorts(vids []string) ([]string, error) {
	portNames := []string{}
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return portNames, err
	}
	if len(ports) == 0 {
		return portNames, fmt.Errorf("no ports were found")
	}
	for _, v := range vids {
		vu := strings.ToUpper(v)
		vl := strings.ToLower(v)
		for _, port := range ports {
			if port.IsUSB {
				if port.VID == v || port.VID == vl || port.VID == vu {
					portNames = append(portNames, port.Name)
				}
			}
		}
	}
	if len(portNames) == 0 {
		return portNames, fmt.Errorf("no port matched the vid list")
	}
	return portNames, nil
}

func remoteUp(port serial.Port, address byte) bool {
	if port == nil {
		//log.Printf("Port %v is closed\n", address)
		return false
	}
	for i := 0; i < upLimit; i++ {
		err := testRemote(port, address)
		if err == nil {
			return true
		}
		time.Sleep(time.Duration(500) * time.Millisecond)
	}
	//log.Printf("Remote %v tested not up", address)
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
	fmt.Printf("Wrote: %v to port %v\n", wBuff, port)
	if n != sendMsgLen {
		fmt.Println("Error 0")
		return fmt.Errorf("did not write %d bytes, it wrote %d", sendMsgLen, n)
	}
	//time.Sleep(time.Duration(5) * time.Millisecond)
	n, err = port.Read(rBuff)
	if err != nil {
		fmt.Println("Error 1")
		return fmt.Errorf("failed to read from usb port: %v", err)
	}
	if n != 1 {
		fmt.Println("Error 2")
		return fmt.Errorf("did not read one byte, read %d:", n)
	}
	if rBuff[0] != ack {
		fmt.Println("Error 3")
		return fmt.Errorf("did not get a %X in return, got %v", ack, rBuff)
	}
	return nil
}

func (app *application) writeRemote(msg []byte, kind string) (int, error) {
	//log.Printf("insite writeRemote\n")
	app.remLock.Lock()
	defer app.remLock.Unlock()
	r, ok := app.rem[kind]
	if !ok {
		//log.Printf("failed map lookup\n")
		err := app.classifyRemotes()
		if err != nil {
			//log.Println("XXXX")
			return 0, err
		}
	}
	if r.port == nil || !r.up {
		//log.Println("Failed port and up test")
		err := app.classifyRemotes()
		if err != nil {
			//log.Println("YYYY")
			return 0, err
		}
	}
	//log.Printf("writing to remote %s\n", kind)
	n, err := app.rem[kind].port.Write(msg)
	return n, err
}

func (app *application) readRemote(msg []byte, kind string) (int, error) {
	app.remLock.Lock()
	defer app.remLock.Unlock()
	r, ok := app.rem[kind]
	if !ok {
		err := app.classifyRemotes()
		if err != nil {
			return 0, err
		}
	}
	if r.port == nil || !r.up {
		err := app.classifyRemotes()
		if err != nil {
			return 0, err
		}
	}
	n, err := app.rem[kind].port.Read(msg)
	return n, err
}

// call this function with the exact slice you want processed
func crc16(message []byte) uint16 {
	//fmt.Println("Message: ", message)
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
