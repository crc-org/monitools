package tools

import (
	"context"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// AvgCPUOverNSeconds returns average %CPU usage
// taken over n seconds
// n : number of polling occasions
// c : channel to report back to the main process
func AvgCPUOverNSeconds(n int, c chan float64) {

	sumCPUPercent := 0.0

	for i := 0; i < n; i++ {
		cmdTop := exec.Command("top", "-bn1", "-u", "qemu")
		out, err := cmdTop.Output()
		if err != nil {
			log.Fatalf("could not capture output of the `top` command")
		}

		outTail := strings.Split(string(out), "qemu")[1]
		cpuPercent, err := strconv.ParseFloat(strings.Fields(outTail)[6], 32)
		if err != nil {
			log.Fatalf("could not parse CPU percentage from the output of the `top` command")
		}
		sumCPUPercent += cpuPercent
		time.Sleep(1 * time.Second)
	}

	c <- sumCPUPercent / float64(n)

}

// RecordHostCPUUsage returns a list of n cpu usage stats
// (in %) taken with nap breaks inbetween each poll
// filename : relative location of file to write into
// reps     : number of times to record CPU usage
// nap      : sleep between the reps
// c        : channel used to report back to the main process
func RecordHostCPUUsage(filename string, reps int, nap int, c chan bool) {

	napLength := time.Duration(nap)
	success := true

	// collect data
	cpuData := make([]string, reps) // can't initialize array of length calculated at runtime :(
	for i := 0; i < reps; i++ {
		// get qemu's line, static output, only once
		cmdTop := exec.Command("top", "-bn1", "-u", "qemu")
		out, err := cmdTop.Output()
		if err != nil {
			log.Fatalf("could not capture output of the `top` command")
			success = false
		}

		outTail := strings.Split(string(out), "qemu")[1]
		cpuData[i] = strings.Fields(outTail)[6]

		time.Sleep(napLength * time.Second)
	}

	// create CSV file and write data to it
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("could not create %s err: %s", filename, err)
		success = false
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	err = writer.Write(cpuData)
	if err != nil {
		log.Fatalf("Could not write data to %s", filename)
		success = false
	}

	c <- success
}

// GetCRIStatsFromVM returns the output of `sudo crictl stats -o yaml`
// from inside the CRC VM
// destinationDir : location where dump YAML file will be saved
// c              : channel to report routines completion/error
func GetCRIStatsFromVM(destinationDir string, c chan error) {

	cmdCrictl := exec.Command("ssh", "-i", "~/.crc/machines/crc/id_ecdsa",
		"core@192.168.130.11",
		"sudo", "crictl", "stats", "-o", "yaml")
	out, err := cmdCrictl.Output() // out is []byte
	if err != nil {
		log.Fatalf("could not capture output of the command: %s", cmdCrictl)
	}

	t := time.Now()
	timestamp := t.Format("20060102150405")
	filename := filepath.Join(destinationDir, "crictl-stats-"+timestamp+".yaml")

	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("could not create %s err: %s", filename, err)
	}
	defer f.Close()

	_, err = f.Write(out)
	if err != nil {
		log.Fatalf("could not write to %s err: %s", filename, err)
	}
	f.Sync()

	c <- err
}

// PushFileToGithub posts a file to Github (GH) repo via GH API
// localFile           : local file to push
// githubRepo          : GH repository name
// githubFile          : what the pushed file should be called on GH
// githubTokenLocation : location of GH token file
// githubOrg           : organisation or username where repo sits
// githubUser          : full name of user to display as author
// githubEmail         : GH email associated with the GH token
func PushFileToGithub(localFile string,
	githubRepo string,
	githubFile string,
	githubTokenLocation string,
	githubOrg string,
	githubUser string,
	githubEmail string) error {

	ctx := context.Background()

	// authenticate & create client
	tokenBytes, err := ioutil.ReadFile(githubTokenLocation)
	if err != nil {
		fmt.Printf("Cannot read the token: %s", err)
		return err
	}
	tokenString := string(tokenBytes)
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tokenString},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// prepare fileContent to be posted to Github
	fileContent, err := ioutil.ReadFile(localFile)
	if err != nil {
		fmt.Println("Could not read local file:", err)
		return err
	}

	// the file must not already exist on Github
	// specify options
	opts := &github.RepositoryContentFileOptions{
		Message:   github.String("New data"),
		Content:   fileContent,
		Branch:    github.String("master"),
		Committer: &github.CommitAuthor{Name: github.String(githubUser), Email: github.String(githubEmail)},
	}

	_, _, err = client.Repositories.CreateFile(ctx, githubOrg, githubRepo, githubFile, opts)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// PushTodaysData pushes today's folder in local directory to Github (GH) via GH API
// localDir            : local dir whose contents to push to GH
// githubRepo          : GH repository name
// githubTokenLocation : location of GH token file
// githubOrg           : organisation or username where repo sits
// githubUser          : full name of user to display as author
// githubEmail         : GH email associated with the GH token
func PushTodaysData(localDir string, githubRepo string, githubTokenLocation string, githubOrg string, githubUser string, githubEmail string) error {

	fmt.Printf("Pushing today's files from: %s to `github.com/%s/%s... ", localDir, githubOrg, githubRepo)

	// round up files to upload to Github
	var files []string
	err := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}

	for i, file := range files {
		if i == 0 { // skip the folder itself
			continue
		}
		err := PushFileToGithub(file, githubRepo, file, githubTokenLocation, githubOrg, githubUser, githubEmail)
		if err != nil {
			log.Printf("Pushing files to GitHub failed: %s", err)
			return err
		}
	}

	fmt.Println("done.")
	return nil
}

// RunCRCCommand runs a CRC command with args
func RunCRCCommand(cmdArgs []string) error {

	completeCommand := exec.Command("crc", cmdArgs...)
	_, err := completeCommand.Output()
	if err != nil {
		log.Fatalf("could not successfully run the command: %s\n err: %s", completeCommand, err)
	}

	return err
}
