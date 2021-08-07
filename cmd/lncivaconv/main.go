package main

import (
	"bufio"
	"fmt"
	"github.com/tamerh/xml-stream-parser"
	"math"
	"os"
	"strconv"
	"strings"
)

type Waypoint struct {
	lon float64  //waypoint longitude
	lat float64  // waypoint latitude
	ident string // waypoint ident
	isVOR bool //is waypoint VOR
	awcPosCnt uint16 //position of waypoint in awc file
	awcFileCnt uint16 //N of awc file
	adcPosCnt uint16 //position of waypoint
	adcFileCnt uint16 //N of adc file
}

type AWCWaypoint struct {
	posCnt uint16 //position of waypoint
	lon string //waypoint longitude
	lat string // waypoint ident
	ident string // waypoint ident, for comments
	isVOR bool //is waypoint VOR
}

type AWCFile struct {
	name string //name of generated file
	waypoints []AWCWaypoint
}

func DegreeToAWCString(value float64, isLon bool) string {
	var result string
	degree, minPart := math.Modf(value)
	minPart *= 60
	var degFmtString string
	if isLon {
		degFmtString = "%03d" //longitude is from 0 to 180
	} else {
		degFmtString = "%02d"
	}
	degString := fmt.Sprintf(degFmtString, int(math.Abs(degree)))

	minStr := fmt.Sprintf("%2.1f", math.Abs(minPart))
	if len(minStr) < 4 {
		minStr = strings.Repeat("0", 4 - len(minStr)) + minStr
	}

	result = fmt.Sprintf("%s*%s", degString, minStr)
	if isLon {
		if value < 0 {
			result = "W " + result
		} else {
			result = "E " + result
		}
	} else {
		if value < 0 {
			result = "S " + result
		} else {
			result = "N " + result
		}
	}
	return result
}


func CreateAWCFile(awcFile *AWCFile, isADC bool) {
	fileName := awcFile.name+".AWC"
	if isADC {
		fileName = awcFile.name+".ADC"
	}

	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	defer file.Close()

	if isADC {
		_, err = fmt.Fprint(file, ";\n; Carousel IV-A ADEU DME Card File\n;\n")
	} else {
		_, err = fmt.Fprint(file, ";\n; Carousel IV-A INS\n; ADEU Waypoints Data Card\n;\n")
	}

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	for idx := range awcFile.waypoints {
		wpt := awcFile.waypoints[idx]
		_, err = fmt.Fprintf(file, "%d %s %s ; %s\n", wpt.posCnt, wpt.lat, wpt.lon, wpt.ident)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(0)
		}
	}

	_, err = fmt.Fprint(file, ";\n; End Of File\n;\n")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
}

func CreateTxtFile(fileName string, waypoints []Waypoint) {
	file, err := os.Create(fileName+".TXT")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	defer file.Close()

	_, err = fmt.Fprintln(file, " AWC                                   ADC")
	_, err = fmt.Fprintln(file, " #/#  WAYPOINT  POSITION               #/#   FREQ")
	_, err = fmt.Fprintln(file, "-------------------------------------------------")

	lastAdcFileNum := uint16(1)
	for idx := range waypoints {
		wpt := waypoints[idx]
		ident:=wpt.ident
		if len(ident) < 5 {
			ident = ident + strings.Repeat(" ", 5 - len(ident))
		}

		if wpt.awcFileCnt == 0 {
			continue
		}

		if !(lastAdcFileNum == wpt.awcFileCnt) {
			lastAdcFileNum = wpt.awcFileCnt
			_, err = fmt.Fprintln(file,"")
		}

		adcPos :="-/-"
		if wpt.adcFileCnt > 0 {
			adcPos = fmt.Sprintf("%d/%d", wpt.adcPosCnt, wpt.adcFileCnt)
		}

		lat := DegreeToAWCString(wpt.lat, false)
		lon := DegreeToAWCString(wpt.lon, true)

		_, err = fmt.Fprintf(file, " %d/%d  %s     %s %s   %s   \n", wpt.awcPosCnt, wpt.awcFileCnt,  ident,  lat, lon, adcPos)
	}
}

func IncWptCnt(wptCnt *uint16) {
	if *wptCnt == 9 {
		*wptCnt = 1
	} else {
		*wptCnt++
	}
}

func PrintUsage() {
	fmt.Println("Usage:")
	fmt.Println("\tlncivaconv [-1] flightplan.lnmpln")
}

func main() {
	waypoints := make([]Waypoint, 0)

	var departure string
	var destination string
	var ident string

	if len(os.Args) < 2 {
		fmt.Println("No filename given. Exiting.")
		PrintUsage()
		os.Exit(0)
	}

	dropFirstWP := true
	var lnmFileName = os.Args[1]
	if lnmFileName == "-1" {
		dropFirstWP = false
		if len(os.Args) < 3 {
			fmt.Println("No filename given. Exiting.")
			PrintUsage()
			os.Exit(0)
		}
		lnmFileName = os.Args[2]
	}

	if !strings.Contains(lnmFileName, string(os.PathSeparator)) {
		currDir, _ := os.Getwd()
		lnmFileName = currDir+string(os.PathSeparator)+lnmFileName
	}
	var lnmFile, err = os.Open(lnmFileName)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	br := bufio.NewReaderSize(lnmFile,65536)
	parser := xmlparser.NewXMLParser(br, "Waypoint")
	for xml := range parser.Stream() {
		//fmt.Println(xml.Name)

		hasIdent := false
		hasPos := false
		hasType := false
		var lon float64
		var lat float64
		isVOR := false

		for s := range xml.Childs {
			//fmt.Printf("%s: %s\n", s, xml.Childs[s][0].InnerText)
			if strings.ToLower(s) == "ident" {
				ident = xml.Childs[s][0].InnerText
				//fmt.Printf("ident: %s\n", ident)
				if len(waypoints) == 0 {
					departure = ident
				}
				hasIdent = true
			}

			if strings.ToLower(s) == "type" {
				hasType = true
				wptType := xml.Childs[s][0].InnerText
				if strings.ToLower(wptType) == "vor" {
					isVOR = true
				}
			}

			if strings.ToLower(s) == "pos" {
				hasLon := false
				hasLat := false

				attrs := xml.Childs[s][0].Attrs
				for attr:= range attrs {
					if strings.ToLower(attr) == "lon" {
						//fmt.Println(attrs[attr])
						hasLon = true
						lon, err = strconv.ParseFloat(attrs[attr],32)
						if err != nil {
							fmt.Println(err.Error())
							os.Exit(0)
						}
					}
					if strings.ToLower(attr) == "lat" {
						//fmt.Println(attrs[attr])
						hasLat = true
						lat, err = strconv.ParseFloat(attrs[attr],32)
						if err != nil {
							fmt.Println(err.Error())
							os.Exit(0)
						}
					}
				}
				if !hasLon || !hasLat {
					fmt.Println("Coords not found")
					os.Exit(0)
				}
				hasPos = true
			} //if pos

			if hasPos && hasIdent && hasType {
				waypoint := Waypoint {
					lat: lat,
					lon: lon,
					ident: ident,
					isVOR: isVOR,
				}
				waypoints = append(waypoints, waypoint)
				break
			}
		}// for s
	}
	destination = ident

	fmt.Printf("%s -> %s\n", departure, destination)
	fmt.Printf("waypoints cnt: %d\n", len(waypoints))

	// Creating AWC files
	awcFiles := make([]AWCFile, 0)
	posCnt := uint16(2) //by default first waypoint of first file always start from 2
	maxWPTs := 8 //max waypoints in file, first file contain 8 without wpt 1
	awcFileIdx := uint16(1) //index of current file
	if !dropFirstWP {
		posCnt = 1 //if no need to skip first waypoint
		maxWPTs = 9
	}

	awcFile := AWCFile {
		name: departure + "-" + destination,
	}

	for idx := range waypoints {

		if dropFirstWP {
			if idx == 0 {
				//skip departure waypoint
				continue
			}
		}

		awcWaypoint := AWCWaypoint{
			posCnt: posCnt,
			ident: waypoints[idx].ident,
			isVOR: waypoints[idx].isVOR,
		}

		waypoints[idx].awcFileCnt = awcFileIdx
		waypoints[idx].awcPosCnt = posCnt

		awcWaypoint.lat = DegreeToAWCString(waypoints[idx].lat, false)
		awcWaypoint.lon = DegreeToAWCString(waypoints[idx].lon, true)
		awcFile.waypoints = append(awcFile.waypoints, awcWaypoint)
		IncWptCnt(&posCnt)

		if len(awcFile.waypoints) == maxWPTs {
			awcFiles = append(awcFiles, awcFile)
			awcFile = AWCFile{
				name: departure + "-" + destination,
			}
			awcFileIdx++

			if awcFileIdx == 1 {
				maxWPTs = 8
			}
			continue
		}

		if idx == len(waypoints) - 1 {
			awcFiles = append(awcFiles, awcFile)
			break
		}
	}

	if len(awcFiles) > 1 {
		//if files more then one add number for each file
		for idx := range awcFiles {
			awcFiles[idx].name = fmt.Sprintf("%s_%d", awcFiles[idx].name, idx + 1)
		}
	}

	for idx := range awcFiles {
		CreateAWCFile(&awcFiles[idx], false)
		//fmt.Printf("File name: %s\n", awcFiles[idx].name)
		//for widx := range awcFiles[idx].waypoints {
		//	wpt := awcFiles[idx].waypoints[widx]
		//	fmt.Printf("%d %s %s ; %s\n", wpt.posCnt, wpt.lat, wpt.lon, wpt.ident)
		//}
	}
	fmt.Printf("Created %d wpt file(s)\n", len(awcFiles))

	//Creating ADC files
	adcFiles := make([]AWCFile, 0)
	adcFileIdx := uint16(1) //index of current file
	posCnt = 1 //ADC files start from 1
	vorWPTCnt := 0 //count of VOR waypoints
	adcFile := AWCFile {
		name: departure + "-" + destination,
	}
	for idx := range waypoints {
		if waypoints[idx].isVOR {

			vorWPTCnt++
			adcWaypoint := AWCWaypoint{
				posCnt: posCnt,
				ident:  waypoints[idx].ident,
				isVOR:  waypoints[idx].isVOR,
			}

			waypoints[idx].adcFileCnt = adcFileIdx
			waypoints[idx].adcPosCnt = posCnt

			adcWaypoint.lat = DegreeToAWCString(waypoints[idx].lat, false)
			adcWaypoint.lon = DegreeToAWCString(waypoints[idx].lon, true)
			adcFile.waypoints = append(adcFile.waypoints, adcWaypoint)
			IncWptCnt(&posCnt)

			if len(adcFile.waypoints) == 9 {
				adcFiles = append(adcFiles, awcFile)
				adcFile = AWCFile{
					name: departure + "-" + destination,
				}
				adcFileIdx++
				posCnt = 1
				continue
			}
		}
		if idx == len(waypoints) - 1 {
			adcFiles = append(adcFiles, adcFile)
			break
		}
	}

	fmt.Printf("VOR waypoints cnt: %d\n", vorWPTCnt)
	if vorWPTCnt > 0 {
		for idx := range adcFiles {
			CreateAWCFile(&adcFiles[idx], true)
		}
		fmt.Printf("Created %d ADC file(s)\n", len(adcFiles))
	}

	CreateTxtFile(departure+"-"+destination, waypoints)
	fmt.Println("Created TXT file")

	os.Exit(1)
}