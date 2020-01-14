package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
)

func randomTempo(min, max int) int {
	if min > max {
		fmt.Printf("Minimum BPM specified (%d) is greater than Maximum BPM (%d). Swapping them!\n", min, max)
		min, max = max, min
	}
	rand.Seed(time.Now().UnixNano())
	tempo := rand.Intn(max-min) + min
	fmt.Printf("Setting new tempo to %d\n", tempo)
	return tempo
}

func updateTempo(tempo int, data []byte) []byte {
	tempoRe := regexp.MustCompile(`<Tempo>\s+<LomId Value="(\d+)" />\s+<Manual Value="(\d+)" />`)
	tVals := tempoRe.FindAllSubmatchIndex(data, -1)
	tIndeces := tVals[0]
	lomVal := data[tIndeces[2]:tIndeces[3]]
	repl := fmt.Sprintf("<Tempo>\n\t\t\t\t\t\t<LomId Value=\"%s\" />\n\t\t\t\t\t\t<Manual Value=\"%d\" />", string(lomVal), tempo)
	data = tempoRe.ReplaceAllLiteral(data, []byte(repl))
	aEventRe := regexp.MustCompile(`<AutomationEnvelope Id="1">\s+<EnvelopeTarget>\s+<PointeeId Value="(\d+)" />\s+</EnvelopeTarget>\s+<Automation>\s+<Events>\s+<FloatEvent Id="(\d+)" Time="(-*\d+)" Value="(\d+)" />`)
	aVals := aEventRe.FindAllSubmatchIndex(data, -1)
	aIndeces := aVals[0]
	aPointeeIDVal, aFEventID, aTime := data[aIndeces[2]:aIndeces[3]], data[aIndeces[4]:aIndeces[5]], data[aIndeces[6]:aIndeces[7]]
	repl = fmt.Sprintf("<AutomationEnvelope Id=\"1\">\n\t\t\t\t\t\t<EnvelopeTarget>\n\t\t\t\t\t\t\t<PointeeId Value=\"%s\" />\n\t\t\t\t\t\t</EnvelopeTarget>\n\t\t\t\t\t\t<Automation>\n\t\t\t\t\t\t\t<Events>\n\t\t\t\t\t\t\t\t<FloatEvent Id=\"%s\" Time=\"%s\" Value=\"%d\" />", string(aPointeeIDVal), string(aFEventID), string(aTime), tempo)
	data = aEventRe.ReplaceAllLiteral(data, []byte(repl))
	return data
}

func listVersions() ([]string, error) {
	versionRe := regexp.MustCompile(`(Live \d+)(\.*\d*\.*\d*)`)
	fs := make([]string, 0)
	files, err := ioutil.ReadDir(abletonBase)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if versionRe.MatchString(file.Name()) {
			fs = append(fs, strings.TrimLeft(file.Name(), "Live "))
		}
	}
	return fs, nil
}

func main() {
	vers, err := listVersions()
	if err != nil {
		log.Fatal(err)
	}
	latest := vers[len(vers)-1]
	ver := flag.String("version", latest, "Version of Ableton Live to target. Defaults to latest version.\nUSAGE alrt -version \""+latest+"\"")
	listVer := flag.Bool("list", false, "List installed versions of Ableton Live")
	min := flag.Int("min", 120, "The minimum BPM")
	max := flag.Int("max", 120, "The maximum BPM")
	flag.Parse()
	if *listVer {
		fmt.Println("Installed versions of Ableton Live:")
		for _, v := range vers {
			fmt.Printf("\t%s\n", v)
		}
		os.Exit(0)
	}
	if *min < 20 || *min > 999 || *max < 20 || *max > 999 {
		fmt.Println(fmt.Errorf("invalid BPM Specified \"-min\" and \"-max\" must be between 20 and 999 inclusive"))
		os.Exit(1)
	}
	templLoc := getDefaultTemplate(*ver)
	templZip, err := os.OpenFile(templLoc, os.O_RDWR, 0755)
	if err != nil {
		fmt.Println(fmt.Errorf("Could not open file %s\n\n\t%v\n\nAbleton Live %s may not be installed, or \"Template.als\" has not been created", templLoc, err, *ver))
		os.Exit(1)
	}
	defer templZip.Close()
	templBak, err := os.OpenFile(templLoc+".bak", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println(fmt.Errorf("Could not open backup file %s: %v", templLoc+".bak", err))
		os.Exit(1)
	}
	defer templBak.Close()
	n, err := io.Copy(templBak, templZip)
	if err != nil {
		fmt.Println(fmt.Errorf("Could not backup template file %s to %s: %v", templLoc, templLoc+".bak", err))
		os.Exit(1)
	}
	fmt.Printf("Backed up %d bytes from %s to %s\n", n, templLoc, templLoc+".bak")
	_, err = templZip.Seek(0, 0)
	if err != nil {
		fmt.Println(fmt.Errorf("Could not seek to beginning of file %s: %v", templLoc, err))
		os.Exit(1)
	}
	templ, err := gzip.NewReader(templZip)
	if err != nil {
		fmt.Println(fmt.Errorf("Could not decompress file %s: %v", templLoc, err))
		os.Exit(1)
	}
	data, err := ioutil.ReadAll(templ)
	if err != nil {
		fmt.Println(fmt.Errorf("Could not read from gzip reader: %v", err))
		os.Exit(1)
	}
	if err := templ.Close(); err != nil {
		fmt.Println(fmt.Errorf("Could not close gzip reader: %v", err))
		os.Exit(1)
	}
	out := updateTempo(randomTempo(*min, *max), data)
	err = templZip.Truncate(0)
	if err != nil {
		fmt.Println(fmt.Errorf("Could not truncate file %s: %v", templLoc, err))
		os.Exit(1)
	}
	_, err = templZip.Seek(0, 0)
	if err != nil {
		fmt.Println(fmt.Errorf("Could not seek to beginning of file %s: %v", templLoc, err))
		os.Exit(1)
	}
	zipWriter := gzip.NewWriter(templZip)
	nz, err := zipWriter.Write(out)
	if err != nil {
		fmt.Println(fmt.Errorf("Could not write template file: %v", err))
		os.Exit(1)
	}
	if err := zipWriter.Close(); err != nil {
		fmt.Println(fmt.Errorf("Could not close gzip writer: %v", err))
		os.Exit(1)
	}
	fmt.Printf("Successfully wrote %d to %s\n", nz, templLoc)
}
