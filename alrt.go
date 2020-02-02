package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"fyne.io/fyne/app"

	"github.com/apenwarr/fixconsole" // Imported for windows console / GUI mixed use. Has no effect otherwise
)

func randomTempo(min, max int) (int, bool) {
	tempo := min
	if min != max {
		rand.Seed(time.Now().UnixNano())
		tempo = rand.Intn(max-min) + min
	}
	fmt.Printf("Setting new tempo to %d\n", tempo)
	return tempo, true
}

func updateTempo(tempo int, data []byte) []byte {
	tempoRe := regexp.MustCompile(`<Tempo>\s+<LomId Value="(\d+)" />\s+<Manual Value="(\d+)" />`)
	tVals := tempoRe.FindAllSubmatchIndex(data, -1)
	tIndices := tVals[0]
	lomVal := data[tIndices[2]:tIndices[3]]
	repl := fmt.Sprintf("<Tempo>\n\t\t\t\t\t\t<LomId Value=\"%s\" />\n\t\t\t\t\t\t<Manual Value=\"%d\" />", string(lomVal), tempo)
	data = tempoRe.ReplaceAllLiteral(data, []byte(repl))
	aEnvRe := regexp.MustCompile(`<AutomationEnvelope Id="1">\s+<EnvelopeTarget>\s+<PointeeId Value="(\d+)" />\s+</EnvelopeTarget>\s+<Automation>\s+<Events>\s+<FloatEvent Id="(\d+)" Time="(-*\d+)" Value="(\d+)" />`)
	aVals := aEnvRe.FindAllSubmatchIndex(data, -1)
	aIndices := aVals[0]
	aPointeeIDVal, aFEventID, aTime := data[aIndices[2]:aIndices[3]], data[aIndices[4]:aIndices[5]], data[aIndices[6]:aIndices[7]]
	repl = fmt.Sprintf("<AutomationEnvelope Id=\"1\">\n\t\t\t\t\t\t<EnvelopeTarget>\n\t\t\t\t\t\t\t<PointeeId Value=\"%s\" />\n\t\t\t\t\t\t</EnvelopeTarget>\n\t\t\t\t\t\t<Automation>\n\t\t\t\t\t\t\t<Events>\n\t\t\t\t\t\t\t\t<FloatEvent Id=\"%s\" Time=\"%s\" Value=\"%d\" />", string(aPointeeIDVal), string(aFEventID), string(aTime), tempo)
	data = aEnvRe.ReplaceAllLiteral(data, []byte(repl))
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

var (
	min     *int
	max     *int
	ver     *string
	vers    []string
	success bool
	err     error
)

func main() {
	err = fixconsole.FixConsoleIfNeeded()
	if err != nil {
		fmt.Println(fmt.Errorf("Could not attach to console: %v", err))
		os.Exit(1)
	}
	vers, err = listVersions()
	if err != nil {
		fmt.Println(fmt.Errorf("Could not enumerate installed Ableton Live versions: %v", err))
		os.Exit(1)
	}
	latest := vers[len(vers)-1]
	cli := flag.Bool("cli", false, "Run CLI Version")
	listVer := flag.Bool("list", false, "List installed versions of Ableton Live")
	help := flag.Bool("help", false, "Display this help message")
	ver = flag.String("version", latest, "Version of Ableton Live to target. Defaults to latest version.\nUSAGE alrt -version \""+latest+"\"")
	min = flag.Int("min", 110, "The minimum BPM")
	max = flag.Int("max", 130, "The maximum BPM")
	flag.Parse()
	if *help {
		fmt.Printf("\nUsage:\n")
		flag.PrintDefaults()
		pressEnterKey()
		os.Exit(0)
	}
	if *listVer {
		fmt.Println("Installed versions of Ableton Live:")
		for _, v := range vers {
			fmt.Printf("\t%s\n", v)
		}
		pressEnterKey()
		os.Exit(0)
	}
	if *min < 20 || *min > 999 || *max < 20 || *max > 999 {
		fmt.Println(fmt.Errorf("invalid BPM Specified \"-min\" and \"-max\" must be between 20 and 999 inclusive"))
		os.Exit(1)
	}
	if *min > *max {
		fmt.Printf("Minimum BPM specified (%d) is greater than Maximum BPM (%d). Swapping them!\n", *min, *max)
		min, max = max, min
	}

	var rt int
	switch {
	case *cli:
		rt, success = randomTempo(*min, *max)
	default:
		a := app.New()
		rt, success = runGUI(a)
	}

	if success {
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

		out := updateTempo(rt, data)
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
		fmt.Printf("Successfully wrote %d bytes (uncompressed) to %s\n", nz, templLoc)
		pressEnterKey()
	}
}
