package cli

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/logutils"
)

type Config struct {
	ID            string
	ResourceName  string
	WafID         string
	Package       string
	Directory     string
	Version       int
	Interactive   bool
	ManageAll     bool
	ForceDestroy  bool
	SkipEditState bool
}

var Bold = color.New(color.Bold).SprintFunc()
var BoldGreen = color.New(color.Bold, color.FgGreen).FprintlnFunc()
var BoldGreenf = color.New(color.Bold, color.FgGreen).FprintfFunc()
var BoldYellow = color.New(color.Bold, color.FgYellow).FprintlnFunc()
var BoldYellowf = color.New(color.Bold, color.FgYellow).FprintfFunc()

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

func YesNo(message string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		BoldYellowf(os.Stderr, "%s [y/n]: ", message)

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
