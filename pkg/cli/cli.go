package cli

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
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
	TestMode      bool
	ReplaceDictionary bool
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

// DataStoreType prompts the user to select a number for a Data Store type and returns the corresponding TF resource name.
// 1 for Config Store, 2 for Secret Store, and 3 for KV Store. It keeps prompting if the input is invalid.
func AskDataStoreType(resource string) string {
	reader := bufio.NewReader(os.Stdin)

	for {
		BoldYellowf(os.Stderr, `"%s" - Select Data Store Type:
	1: Config Store
	2: Secret Store
	3: KV Store
	Enter number: `, resource)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.TrimSpace(response)
		choice, _ := strconv.Atoi(response)

		switch choice {
		case 1:
			return "fastly_configstore"
		case 2:
			return "fastly_secretstore"
		case 3:
			return "fastly_kvstore"
		default:
			fmt.Fprintf(os.Stderr, "\n")
		}
	}
}
