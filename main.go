package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/malashin/ffinfo"
	"github.com/wlredeye/jsonlines"
)

var databasePath string = "database.json"
var outputPath string = "output.txt"

var dbSlice []File
var dbMap = make(map[string]File)

type UuidURL struct {
	UUID string
	URL  string
}

type File struct {
	UuidURL
	Probe ffinfo.File
}

type Data struct {
	File
	FormatStartTime float64
	StreamStartTime []float64
	NonZero         bool
}

func main() {
	// Read database file and fill up the database map.
	if _, err := os.Stat(databasePath); err != nil {
		panic(err)
	}

	file, err := os.Open(databasePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = jsonlines.Decode(file, &dbSlice)
	if err != nil {
		panic(err)
	}

	for _, entry := range dbSlice {
		dbMap[entry.UUID] = entry
	}

	f, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0775)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	i := 0
	total := len(dbMap)
	for _, f := range dbMap {
		fmt.Printf("%v/%v: ", i+1, total)
		i++

		d, err := getStartTimes(f)
		if err != nil {
			panic(err)
		}

		if d.NonZero {
			fmt.Printf("%v %v %v %v\n", d.UUID, d.URL, d.FormatStartTime, d.StreamStartTime)
			writeStringToFile(outputPath, fmt.Sprintf("%v\t%v\t%v\t%v\n", d.UUID, d.URL, d.FormatStartTime, (strings.Trim(strings.Join(strings.Fields(fmt.Sprint(d.StreamStartTime)), "\t"), "[]"))))
			continue
		}
		fmt.Print("\r")
	}
}

func getStartTimes(f File) (d Data, err error) {
	d.File = f

	d.FormatStartTime, err = strconv.ParseFloat(d.File.Probe.Format.StartTime, 64)
	if err != nil {
		return d, err
	}

	if d.FormatStartTime > 0 {
		d.NonZero = true
	}

	for _, s := range d.File.Probe.Streams {
		streamStartTime, err := strconv.ParseFloat(s.StartTime, 64)
		if err != nil {
			return d, err
		}

		d.StreamStartTime = append(d.StreamStartTime, streamStartTime)

		if streamStartTime > 0 {
			d.NonZero = true
		}
	}

	return d, nil
}

func writeStringToFile(filename string, str string) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(str); err != nil {
		return err
	}

	return nil
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
