package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

/*

I am moving from implementing the CW genrator, VFO driver and band reader all on the
Raspberry Pi/Linux written in Go and time sharing with other tasks on the Pi to one
where the CW generator (see the cwDuino file) and the frequency writer and band reader
each have thier own Arduino Uno R4 Minima units.  That will make the CW generator much
more predictable (no background processes to interfer with them).  It may also permit
me to build split transmit/recieve frequency function since the Arduino dedicated to
this task will be free of other activities.

For finding the ports, initializing and regularly testing them, see the duino.go file.

The remote vfo Arduino commands are composed of:

address: one byte, 0x81 in this case because there will be more remotes
command: one byte (see below)
data: one or more bytes (see below)
crc: two bytes

The remote returns 0xFF for a good command, 0x00 or nothing for a bad command

The inital commands are:
01: are you there, this command has no data bytes, so it is a total of 4 significant
bytes with the CRC being the last two bytes and the space beween the first two and
the last two bytes zero padded.
this command is common to all peripherals.  Only the address is different.

02: set split
data: one address byte, one command byte, and one data byte, 2 CRC bytes for a total
of 5 significant bytes.  The bytes beween the first two bytes and the last two CRC
bytes are zero padded.
A data byte of 0xFF puts the remote in split mode, a data byte of 0x00 puts the remote
in no split mode (recieve frequency the same as transmit frequency).

03: set transmit frequency
data: one address byte, one command byte, 4 bytes of data (32 bit integer set in four
bytes, high byte first,2 bytes of CRC for a total of 8 bytes

04: set recieve frequency
data: one address byte, one command byte, 4 bytes of data (32 bit integer set in four
bytes, high byte first,2 bytes of CRC for a total of 8 bytes
This command is only needed when in split/RIT mode.

The remote vfo reading of the band setting is composed of a command and a returned
value.  The returned CRC is calculated over the combination of the sent command
and the returned value

05: read band setting
data sent:  1 byte of address and one byte of command, and two bytes of CRC.
The space between the first two bytes and the last two CRC bytes are zero padded.
3 bytes data returned: one byte of value, 2 bytes of CRC calculated over the first
two bytes..  The bottom 3 bits (0, 1 and 2) of the returned byte are the
binary encoding of the discrete band select switch leads.  Bit 3 is the
status of the Transmit lead.

*/

const (
	vfoAddress byte = 0x81
	setSplit   byte = 0x02
	setTx      byte = 0x03
	setRx      byte = 0x04
	split      byte = 0xFF
	noSplit    byte = 0x00
	bandCmd    byte = 0x05
)

func (app *application) startVFORemote() error {
	portNames, err := findPorts(vidList)
	if err != nil {
		return err
	}
	log.Println("portNames: ", portNames)
	for _, portName := range portNames {
		log.Println(portName)
		port, err := initPort(portName)
		log.Printf("X Error ***%v***\n", err)
		if err != nil {
			if strings.Contains(err.Error(), "Serial port busy") {
				continue
			}
			return err
		}
		if remoteUp(port, vfoAddress) {
			app.rem[vfoKind] = &remote{
				port:     port,
				portName: portName,
				//		vid:      vid,
				up: true,
			}
			return nil
		}
		port.Close()
	}
	return fmt.Errorf("No vfo ports were found")
}

func (app *application) setSplit(s string) error {
	wBuff := make([]byte, sendMsgLen)
	rBuff := make([]byte, 1)

	wBuff[0] = vfoAddress
	wBuff[1] = setSplit
	switch s {
	case "split":
		wBuff[2] = split
	case "noSplit":
		wBuff[2] = noSplit
	default:
		wBuff[2] = noSplit
	}
	crc := crc16(wBuff[:3])
	wBuff[sendMsgLen-2] = byte(crc >> 8)
	wBuff[sendMsgLen-1] = byte(crc)

	n, err := app.writeRemote(wBuff, "vfo") //rem[vfoKind].port.Write(wBuff)
	if err != nil {
		return fmt.Errorf("failed write to usb port %v", err)
	}
	if n != 5 {
		return fmt.Errorf("did not write 5 bytes, it wrote %d", n)
	}
	n, err = app.readRemote(rBuff, "vfo") //rem[vfoKind].port.Read(rBuff)
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

	n, err := app.writeRemote(wBuff, "vfo") //rem[vfoKind].port.Write(wBuff)
	if err != nil {
		return fmt.Errorf("failed write to usb port %v", err)
	}
	if n != 8 {
		return fmt.Errorf("did not write 8 bytes, it wrote %d", n)
	}
	n, err = app.readRemote(rBuff, "vfo") //rem[vfoKind].port.Read(rBuff)
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
	n, err := app.writeRemote(wBuff, "vfo") //rem[vfoKind].port.Write(wBuff)
	if err != nil {
		return 0, errors.Join(fmt.Errorf("failed write to vfo usb port"), err)
	}
	if n != sendMsgLen {
		return 0, fmt.Errorf("did not write %d bytes, it wrote %d", sendMsgLen, n)
	}
	//	time.Sleep(time.Duration(120) * time.Millisecond)
	n, err = app.readRemote(rBuff, "vfo") //rem[vfoKind].port.Read(rBuff)
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
