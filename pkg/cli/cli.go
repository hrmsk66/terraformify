package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/logutils"
)

type Config struct {
	ID          string
	Version     int
	Directory   string
	Interactive bool
	ManageAll   bool
}

var Bold = color.New(color.Bold).SprintFunc()
var BoldGreen = color.New(color.Bold, color.FgGreen).SprintFunc()
var BoldYellow = color.New(color.Bold, color.FgYellow).SprintFunc()

func CreateLogFilter() io.Writer {
	minLevel := os.Getenv("TMFY_LOG")
	if minLevel == "" {
		minLevel = "INFO"
	}
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(minLevel),
		Writer:   os.Stderr,
	}
	return filter
}

func CheckDirEmpty(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	if !info.IsDir() {
		log.Fatal(fmt.Errorf("%s is not a directory", path))
	}

	d, err := os.Open(path)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = d.Readdir(1)
	if err == io.EOF {
		return nil
	}

	return errors.New("Working directory is not empty")
}

func YesNo(message string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", message)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
