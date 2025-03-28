package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Timeout struct {
	Create time.Duration
	Read   time.Duration
	Update time.Duration
	Delete time.Duration
}

func parseTimeoutsFromGoFile(filePath string) (Timeout, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return Timeout{}, err
	}
	defer file.Close()

	var timeouts Timeout
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Create:") {
			timeouts.Create = parseDuration(line)
		} else if strings.Contains(line, "Read:") {
			timeouts.Read = parseDuration(line)
		} else if strings.Contains(line, "Update:") {
			timeouts.Update = parseDuration(line)
		} else if strings.Contains(line, "Delete:") {
			timeouts.Delete = parseDuration(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return Timeout{}, err
	}

	return timeouts, nil
}

func parseTimeoutsFromMarkdownFile(filePath string) (Timeout, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return Timeout{}, err
	}
	defer file.Close()

	var timeouts Timeout
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "`create`") {
			timeouts.Create = parseDuration(line)
		} else if strings.Contains(line, "`read`") {
			timeouts.Read = parseDuration(line)
		} else if strings.Contains(line, "`update`") {
			timeouts.Update = parseDuration(line)
		} else if strings.Contains(line, "`delete`") {
			timeouts.Delete = parseDuration(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return Timeout{}, err
	}

	return timeouts, nil
}

func parseDuration(line string) time.Duration {
	re := regexp.MustCompile(`(\d+).*(Hour|Minute|hour|minute)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) != 3 {
		return 0
	}

	value := matches[1]
	unit := strings.ToLower(matches[2])[0]

	if d, err := time.ParseDuration(fmt.Sprintf("%s%c", value, unit)); err != nil {
		fmt.Errorf("unable to parse duration expression: %s", err)
		return 0
	} else {
		return d
	}
}

func findFile(root, fileName string) (string, error) {
	var foundPath string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == fileName {
			foundPath = path
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if foundPath == "" {
		return "", fmt.Errorf("file %s not found", fileName)
	}
	return foundPath, nil
}

func viewResults(data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"", "Def", "Docs"})

	for _, v := range data {
		if v[1] != v[2] {
			v[0] = color.New(color.FgRed, color.Bold).Sprint(v[0])
			v[1] = color.New(color.FgRed, color.Bold).Sprint(v[1])
			v[2] = color.New(color.FgRed, color.Bold).Sprint(v[2])
		}
		table.Append(v)
	}

	table.Render()
}

func extractResourceName(path string) string {
	base := filepath.Base(path)
	return strings.Replace(strings.Replace(base, "_resource.go", "", 1), "_data_source.go", "", 1)
}

func main() {
	if len(os.Args) < 3 {
		if exec, err := os.Executable(); err != nil {
			fmt.Errorf("unable to get executable path: %s", err)
		} else {
			fmt.Printf("Usage: %s <path_to_source_file> <repo_root>\n", exec)
		}
		return
	}

	fmt.Println("Checking Timeouts for Resource/ Data Source Documentation")

	goFileName := os.Args[1]

	if !strings.HasSuffix(goFileName, "_resource.go") && !strings.HasSuffix(goFileName, "_data_source.go") {
		fmt.Printf("'%s' is probably not a resource or data source definition.\n", goFileName)
		os.Exit(0)
	}

	rootDir := os.Args[2]
	mdFileName := fmt.Sprintf("%s.html.markdown", extractResourceName(goFileName))

	file, err := os.Open(goFileName)
	if err != nil {
		fmt.Errorf("unable to open file: %s", err)
	}
	defer file.Close()

	goTimeouts, err := parseTimeoutsFromGoFile(file.Name())
	if err != nil {
		fmt.Printf("Error reading Go file: %v\n", err)
		return
	}

	if strings.Contains(goFileName, "_resource.go") {
		rootDir += "/website/docs/r/"
	} else {
		rootDir += "/website/docs/d/"
	}

	mdFilePath, err := findFile(rootDir, mdFileName)
	if err != nil {
		fmt.Printf("Error finding Markdown file: %v\n", err)
		return
	}
	fmt.Printf("Found matching documentation file %s\n", mdFilePath)
	mdTimeouts, err := parseTimeoutsFromMarkdownFile(mdFilePath)
	if err != nil {
		fmt.Printf("Error reading Markdown file: %v\n", err)
		return
	}

	results := [][]string{
		[]string{"Create", goTimeouts.Create.String(), mdTimeouts.Create.String()},
		[]string{"Read", goTimeouts.Read.String(), mdTimeouts.Read.String()},
		[]string{"Update", goTimeouts.Update.String(), mdTimeouts.Update.String()},
		[]string{"Delete", goTimeouts.Delete.String(), mdTimeouts.Delete.String()},
	}

	viewResults(results)

	if goTimeouts == mdTimeouts {
		fmt.Println(color.GreenString("* Documentation Matches! *"))
	} else {
		fmt.Println(color.RedString("* Documentation Does Not Match! *"))
		os.Exit(1)
	}
}
