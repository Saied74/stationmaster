package main

///*
//Format of DX Clusters return with call
//Call^Frequency^Date/Time^Spotter^Comment^LoTW user^eQSL user^Continent^Band^Country name
//*/

import (
//"errors"
//"fmt"
//"io/ioutil"
//"net"

////	"log"
//"net/http"
////	"os"
//"strconv"
//"strings"
//"time"

////	"text/tabwriter"
//"unicode/utf8"
)

//type dxItemType int

//type dxItem struct {
//typ dxItemType
//val string
//}

//type dxLexer struct {
//name  string      // used only for error reports.
//input string      // the string being scanned.
//start int         // start position of this item.
//pos   int         // current position in the input.
//width int         // width of last rune read from input.
//field dxItemType  //field to emit into the chanel
//items chan dxItem // channel of scanned items.
//}

//const (
//dxItemCall dxItemType = iota
//dxItemFreq
//dxItemSpotter
//dxItemComment
//dxItemDate
//dxItemLOTW
//dxItemEQSL
//dxItemBand
//dxItemContinent
//dxItemCountry
//dxItemADIFCountry
//dxItemEOR
//dxItemEOF
//dxItemError
//)

//type dxStateFn func(*dxLexer) dxStateFn

////type DXClusters struct {
////DE        string
////DXStation string
////Country   string
////Frequency string
////Need      string
////}

//var errNoDXSpots = errors.New("no dx spots")

//var scanList = []dxItemType{dxItemSpotter, dxItemFreq, dxItemCall, dxItemComment, dxItemDate,
//dxItemLOTW, dxItemEQSL, dxItemBand, dxItemContinent, dxItemCountry, dxItemADIFCountry}
//var scanIndex int
//var scanMod = len(scanList)
//var printList = []string{"Call", "Freq", "Spotter", "Comment", "Date",
//"LOTW", "EQSL", "Band", "Continent", "Country", "ADIFCountry"}

//const dxeof = -1

//var pattern = `UA3T^14010.0^S79/DH5FS^^0255 2021-09-24^^^EU^20M^Germany^230 DM3F^1234567.0^S79/DH5FS^up1^0252 2021-09-24^^^EU^20M^Germany^230 VK7XX^14080.0^3D2CR^FT4 CQ CQ CQ very few takers^0243 2021-09-24^L^^OC^20M^Fiji^176 WG5G^14080.0^3D2CR^wked qrp^0241 2021-09-24^L^^OC^20M^Fiji^176 K6VOX^14228.0^ZL1KEN^55-58 in San Diego^0232 2021-09-24^L^^OC^20M^New Zealand^170 K6VOX^14205.0^ZL1ACE^57-59+ in San Diego^0229 2021-09-24^^^OC^20M^New Zealand^170 KG7V^14076.6^RK9S^^0229 2021-09-24^^^AS^20M^Russia (Asiatic)^15 KG7V^14076.6^JA8ISK^^0227 2021-09-24^^^AS^20M^Japan^339 KG7V^14076.6^UA0AV^^0226 2021-09-24^L^E^AS^20M^Russia (Asiatic)^15 W5UC^14275.0^W3FF^USB EM12xh -> CN80um^0216 2021-09-24^^^NA^20M^United States^291`

//func dxLex(name, input string) (*dxLexer, chan dxItem) {
//l := &dxLexer{
//name:  name,
//input: input,
//items: make(chan dxItem),
//}
//go l.run() // Concurrently run state machine.
//return l, l.items
//}

//// run lexes the input by executing state functions until
//// the state is nil.
//func (l *dxLexer) run() {
//for state := dxLexText; state != nil; {
//state = state(l)
//}
//close(l.items) // No more tokens will be delivered.
//}

//func (l *dxLexer) emit(t dxItemType) {
//l.items <- dxItem{t, l.input[l.start:l.pos]}
//l.start = l.pos
//// l.items <- item{l.field, l.input[l.start:l.pos]}
//// l.start = l.pos
//}

//// next returns the next rune in the input.
//func (l *dxLexer) next() (r rune) {
//if l.pos >= len(l.input) {
//l.width = 0
//return dxeof
//}
//r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
//l.pos += l.width
//return r
//}

//func dxLexText(l *dxLexer) dxStateFn {
//for {
//if strings.HasPrefix(l.input[l.pos:], "^") {
//if l.pos > l.start {
//if scanIndex != 0 {
//l.start++
//}
//l.emit(scanList[scanIndex])
//if scanIndex == scanMod-2 {
//l.pos++
//return munchNumbers
//}
//return dxLexText // Next state.
//}
//scanIndex++
//if scanIndex >= scanMod {
//scanIndex = 0
//l.emit(dxItemEOR)
//}
//}
//if l.next() == eof {
//break
//}
//}
//// Correctly reached EOF.
//l.field = dxItemEOF
//l.emit(dxItemEOF) // Useful to make EOF a token.
//return nil        // Stop the run loop.
//}

//func munchNumbers(l *dxLexer) dxStateFn {
//for {
//w, _ := utf8.DecodeRuneInString(l.input[l.pos:])
//_, err := strconv.Atoi(string(w))
//if err == nil {
//if l.next() == eof {
//break
//}
//// fmt.Println(l.input[l.start:l.pos])
//return munchNumbers
//}
//if l.pos > l.start {
//l.start++
//l.emit(dxItemADIFCountry)
//l.emit(dxItemEOR)
//scanIndex = 0
//return dxLexText
//}
//if l.next() == eof {
//break
//}
//}
//l.emit(dxItemEOF)
//return nil
//}

//func getCluster(url string) ([]byte, error) {
//client := &http.Client{
//Timeout: 2 * time.Second,
//}
//resp, err := client.Get(url)
//if err != nil {
//if e, ok := err.(net.Error); ok && e.Timeout() {
////fmt.Println("This was a timeout")
//return []byte{}, errNoDXSpots
//}
//return []byte{}, err
//}
//defer resp.Body.Close()

//if resp.StatusCode != http.StatusOK {
//return []byte{}, fmt.Errorf("Status error: %v", resp.StatusCode)
//}

//data, err := ioutil.ReadAll(resp.Body)
//if err != nil {
//return []byte{}, fmt.Errorf("Read body: %v", err)
//}
//return data, nil
//}

//func clusters(band string, lines int) ([]DXClusters, error) {

//url := fmt.Sprintf("https://www.hamqth.com/dxc_csv.php?limit=%d&band=%s", lines, band)

//data, err := getCluster(url)
//if err != nil {
//return []DXClusters{}, err
//}
//outputItem := DXClusters{}
//outputTable := []DXClusters{}
//var b bool
//_, c := dxLex("dxclusters", string(data))
//for {
//d := <-c
//switch d.typ {
//case dxItemEOR:
//outputTable = append(outputTable, outputItem)
//outputItem = DXClusters{}
//case dxItemEOF:
//b = true
//break
//case dxItemError:
//return []DXClusters{}, fmt.Errorf("error from scanner %v", dxItemError)
//case dxItemSpotter:
//outputItem.DE = strings.TrimLeft(d.val, " ")
//outputItem.DE = strings.TrimLeft(d.val, "\n")
//case dxItemCall:
//outputItem.DXStation = strings.TrimLeft(d.val, " ")
//case dxItemCountry:
//outputItem.Country = strings.TrimLeft(d.val, " ")
//case dxItemFreq:
//outputItem.Frequency = strings.TrimLeft(d.val, " ")
//default:
//continue
//}
//if b {
//break
//}
//}
//return outputTable, nil
//}
