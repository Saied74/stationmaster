package main

/*
I am moving from hardware based CW tone genertor to a softwware based CW generator
built on an Arduino Uno R4 minima.  The interface will change from GPIO to USB/Serial.
I do not want to supply power from the USB to the Arduino since there is likely to be
many loads like it (frequency synthesizer, 500 watt tuner, linear amp...).  So, the
interface will be USB to serial using one of the Adafruit modules.
For that, I need to be able to determine if the USB is plugged in and if there is
an Arduino implementing the CW function beyond it.  I will also need a minimal set of
commands over this link.
For that, I will copy the work that I did on the dishmaster program (also on my GitHub
page.  First, the tests:

function findPort takes in a USB vendor ID (vid) and returns the name of the port found
or error if no port was found.  It will be used both during initialization and when
accessing the port.  Port name will be saved in the structure app at initialization
so it can be checked while the program is running.

---------------------------------- Important Note -------------------------------------
Since I am planning to use multiple USB to Serial converters with the same vendor ID,
vid/pid by themselves are not sufficient.  Also, since there will be multiple devices
with the same vid/pid detected, the whole device discovery strategy needs to be rethought.
---------------------------------------------------------------------------------------

function initPort takes in the port name string and returns the port (serial.Port)
or error

Function remoteUp takes in port type serial.Port and returns true or fale.  It depends
on a function testRemote which implements one of the remote cw Arduino commands.

The remote cw Arduino commands are composed of:

address: one byte, 0x80 in this case because there will be more remotes
command: one byte (see below)
data: one or more bytes (see below)
crc: two bytes

The remote returns 0xFF for a good command, 0x00 or nothing for a bad command

The inital commands are:
01: are you there, this command has no data bytes, so it is a total of 4 bytes

02: tutor mode
data: one address byte, one command byte, 2 bytes of tone (uint16), high byte first,
one byte of volume, one byte of dit time in milliseconds (dah time is 3 times dit time),
2 bytes of CRC for a total of 8 bytes

03: keyer mode
data: one address byte, one command byte, 2 bytes of tone (uint16), high byte first,
one byte of speed, one byte of dit time in milliseconds (dah time is 3 times dit time),
2 bytes of CRC for a total of 8 bytes



*/

import (
	"fmt"
)

const (
	cwAddress byte = 0x80 //cw module address
	tutor     byte = 0x02
	keyer     byte = 0x03
)

type cwCmd struct {
	cmd  byte   //see the commands in the comments above
	tone uint16 //requency in HZ
	vol  uint8  //volume
	dit  uint8  // 1200/wpm in milliseconds
}

func (app *application) startCWRemote() error {
	portNames, err := findPorts(vidList)
	if err != nil {
		return err
	}
	for _, portName := range portNames {
		port, err := initPort(portName)
		if err != nil {
			if err.Error() == "Serial port busy" {
				continue
			}
			return err
		}
		if remoteUp(port, cwAddress) {
			app.rem[cwKind] = &remote{
				port:     port,
				portName: portName,
				//		vid:      vid,
				up: true,
			}
			return nil
		}
		port.Close()
	}
	return fmt.Errorf("No cw ports were found")
}

//func findPort(v string) (port string, err error) {
//	v = strings.ToUpper(v)
//	ports, err := enumerator.GetDetailedPortsList()
//	if err != nil {
//		return "", err
//	}
//	if len(ports) == 0 {
//		return "", fmt.Errorf("no ports were found")
//	}
//	for _, port := range ports {
//		if port.IsUSB {
//			if port.VID == v {
//				return port.Name, nil
//			}
//		}
//	}
//	return "", fmt.Errorf("right usb port not found")
//}

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

// call this function with the exact slice you want processed
//func crc16(message []byte) uint16 {
//	fmt.Println("Message: ", message)
//	crc := uint16(0x0000) // Initial value
//	for _, b := range message {
//		crc ^= uint16(b) << 8
//		for i := 0; i < 8; i++ {
//			if crc&0x8000 != 0 {
//				crc = (crc << 1) ^ 0x1021
//			} else {
//				crc <<= 1
//			}
//		}
//	}
//	return crc
//}
