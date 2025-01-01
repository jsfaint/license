package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/ncruces/zenity"
	"golang.org/x/mod/modfile"
)

func getLicense(url string) string {
	doc, err := htmlquery.LoadURL(url)
	if err != nil {
		panic(fmt.Sprintf("Failed to load URL %s: %v", url, err))
	}
	// Find all news item.
	node := htmlquery.FindOne(doc, `//*[@id="main-content"]/header/div/div[2]/span[3]/a`)
	if node == nil {
		return "Unknown"
	}

	txt := htmlquery.InnerText(node)

	if strings.Contains(txt, "not legal advice") {
		return "Unknown"
	}

	return txt
}

func getPackagesList(name string) (list []string) {
	data, err := os.ReadFile(name)
	if err != nil {
		return
	}

	file, err := modfile.Parse(filepath.Base(name), data, nil)
	if err != nil {
		return
	}

	for _, require := range file.Require {
		list = append(list, require.Mod.Path)
	}

	return list
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		zenity.Error("Failed to get current working directory: "+err.Error(), zenity.Title("Error"), zenity.ErrorIcon)
		return
	}

	inName, err := zenity.SelectFile(
		zenity.Filename(wd),
		zenity.FileFilters{
			{
				Name:     "Go Module",
				Patterns: []string{"go.mod"},
				CaseFold: false},
		},
	)
	if err != nil {
		zenity.Error("Failed to select go.mod file: "+err.Error(), zenity.Title("Error"), zenity.ErrorIcon)
		return
	}

	outName := filepath.Base(filepath.Dir(inName)) + ".csv"

	// create csv file
	out, err := os.Create(outName)
	if err != nil {
		zenity.Error("Failed to create output file: "+err.Error(), zenity.Title("Error"), zenity.ErrorIcon)
		return
	}
	defer out.Close()

	// Get package list from go.mod file
	list := getPackagesList(inName)

	dlg, err := zenity.Progress(
		zenity.Title("Running..."))
	if err != nil {
		zenity.Error("Create progress dialog failed: "+err.Error(), zenity.Title("Error"), zenity.ErrorIcon)
		os.Exit(1)
	}
	defer dlg.Close()

	// Write csv header
	fmt.Fprintln(out, "Name,License")
	for _, v := range list {
		// Get license from pkg.go.dev
		dlg.Text("Getting license for " + v + "...")
		license := getLicense("https://pkg.go.dev/" + v)
		// Write to csv file
		fmt.Fprintln(out, v+","+license)
	}

	dlg.Complete()
}
