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
}

type AWCWaypoint struct {
	posCnt int //position of waypoint
	lon string //waypoint longitude
	lat string // waypoint ident
	ident string // waypoint ident, for comments
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


func CreateAWCFile(awcFile *AWCFile) {
	file, err := os.Create(awcFile.name+".AWC")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	defer file.Close()


	_, err = fmt.Fprint(file, ";\n; Carousel IV-A INS\n; ADEU Waypoints Data Card\n;\n")
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

func PrintUsage() {
	fmt.Println("Usage:")
	fmt.Println("\tlncivaconv [-1] flightplan.lnmpln")
}

func main() {
	waypoints := make([]Waypoint, 0)
	awcFiles := make([]AWCFile, 0)

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
		var lon float64
		var lat float64

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

			if hasPos && hasIdent {
				waypoint := Waypoint {
					lat: lat,
					lon: lon,
					ident: ident,
				}
				waypoints = append(waypoints, waypoint)
				break
			}
		}// for s
	}
	destination = ident

	fmt.Printf("%s -> %s\n", departure, destination)
	fmt.Printf("waypoints cnt: %d\n", len(waypoints))

	awcFile := AWCFile {
		name: departure + "-" + destination,
	}

	posCnt := 2 //by default first waypoint of first file always start from 2
	if !dropFirstWP {
		posCnt = 1 //if no need to skip first waypoint
	}
	awcFileIdx := 1 //index current file
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
		}

		awcWaypoint.lat = DegreeToAWCString(waypoints[idx].lat, false)
		awcWaypoint.lon = DegreeToAWCString(waypoints[idx].lon, true)
		awcFile.waypoints = append(awcFile.waypoints, awcWaypoint)

		if awcFileIdx == 1 {
			//first file start from 2 to 9
			if posCnt == 9 {
				awcFiles = append(awcFiles, awcFile)
				//reset waypoint position to 1
				posCnt = 1
				//create new file record
				awcFile = AWCFile{
					name: departure + "-" + destination,
				}
				awcFileIdx ++
				continue
			}
		}

		if awcFileIdx > 1 {
			//second file start from 1 to 8 and others from 9 to 8
			if posCnt == 8 {
				awcFiles = append(awcFiles, awcFile)
				//reset waypoint position to 9
				posCnt = 9
				//create new file record
				awcFile = AWCFile{
					name: departure + "-" + destination,
				}
				awcFileIdx ++
				continue
			}
		}

		if idx == len(waypoints) - 1 {
			awcFiles = append(awcFiles, awcFile)
			break
		}

		posCnt++
		if posCnt == 10 {
			posCnt = 1
		}
	}

	if len(awcFiles) > 1 {
		//if files more then one add number for each file
		for idx := range awcFiles {
			awcFiles[idx].name = fmt.Sprintf("%s_%d", awcFiles[idx].name, idx + 1)
		}
	}

	for idx := range awcFiles {
		CreateAWCFile(&awcFiles[idx])
		//fmt.Printf("File name: %s\n", awcFiles[idx].name)
		//for widx := range awcFiles[idx].waypoints {
		//	wpt := awcFiles[idx].waypoints[widx]
		//	fmt.Printf("%d %s %s ; %s\n", wpt.posCnt, wpt.lat, wpt.lon, wpt.ident)
		//}
	}
	fmt.Printf("Created %d file(s)\n", len(awcFiles))
}