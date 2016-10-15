package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
)

var projectName, envName string

// var projectName, runTagPrefix, envName string

func main() {
	flag.StringVar(&projectName, "PrjName", "TOS", "Project Name")
	// flag.StringVar(&runTagPrefix, "RunTag", "", "Run Tag prefix")
	flag.StringVar(&envName, "envName", "ist7", "test env name")

	flag.Usage = func() {
		fmt.Printf("%s by Jusong Chen\n", os.Args[0])
		fmt.Println("Usage:")
		fmt.Printf("   %s [flags] path pattern \n", os.Args[0])
		fmt.Println("Flags:")
		flag.PrintDefaults()
		os.Exit(-1)
	}

	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
	}
	path, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Fatalf("Cannot get absolute path:%s", flag.Arg(0))
	}
	pattern := flag.Arg(1)

	m, err := getFiles(path, pattern)
	if err != nil {
		log.Fatal(err)
	}
	for _, name := range m {
		// fmt.Println(name)
		if err := loadAWR(name, projectName); err != nil {
			log.Print(err)
		}
	}
}

/*
generate cmd like these:
./loadawr.sh   / gia_0916_basemist2a204na1-11036164_276243_276246mist2a204na1-1rate_200.dmp GIA.9b target_run1@rate_200
./loadawr.sh   / tos_round2_2_baseinitist7a202na1-11041757_278053_278055ist7a202na1-1rate_160.dmp TOS target_round2_run2@rate_160
*/
func loadAWR(Fullname string, projectName string) error {
	dir, fileName := filepath.Split(Fullname)
	ratePattern := "(.*?)" + envName + ".*(rate[_].*)[.]dmp"
	re := regexp.MustCompile(ratePattern)

	//extract rate_120 from tos_round2_2_ba40ist7a202na1-1rate_120.dmp
	matches := re.FindStringSubmatch(fileName)
	if len(matches) != 3 {
		return errors.Errorf("File skipped as pattern %s not matched:%s", ratePattern, fileName)
	}

	runTag := matches[1] + "@" + matches[2]
	cmd := exec.Command("./loadawr.sh", "/", fileName, projectName, runTag, dir)
	fmt.Println(cmd.Args)
	return cmd.Run()
}

// getFiles search directory to get files with the pattern
func getFiles(root string, pattern string) ([]string, error) {

	m := []string{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		matched, err := regexp.MatchString(pattern, info.Name())
		if err != nil {
			return err
		}
		// matched := true
		if matched {

			// fmt.Println("Find file:", path)
			m = append(m, path)
		}
		return nil
	})
	// fmt.Printf("Files:%v", m)
	return m, err
}
