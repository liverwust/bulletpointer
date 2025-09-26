// Apply sequencing logic to apply "layers" to SVG files, which then produce
// PNG files for insertion to videos.

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/beevik/etree"
	"gopkg.in/yaml.v3"
)

// Represent an individual SVG file which will be used to generate the PNG
// files that represent layers on that image.
type Image struct {
	Filename string `yaml:"filename"`
	Layers []*ImageLayer `yaml:"layers"`
}

// In the context of an individual SVG file, loop through and apply the
// layering logic to produce individual "slides" for video insertion.
func (image *Image) processImage(inDir string, outDir string) {
	inFile := filepath.Join(inDir, image.Filename)
	if fileStat, err := os.Stat(inFile); err == nil {
		if !fileStat.Mode().IsRegular() {
			log.Fatalf("Input file %s is not regular file\n", inFile)
		}
	} else {
		log.Fatalf("Source file needs to exist: %s\n", inFile)
	}

	outPrefix := filepath.Base(inFile)
	outExt := filepath.Ext(outPrefix)
	outPrefix = outPrefix[0:(len(outPrefix) - len(outExt))]

	if strings.ToLower(outExt) != ".svg" {
		log.Fatalf("Expected .svg file but got %s\n", inFile)
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromFile(inFile); err != nil {
		log.Fatalf("Error reading SVG XML file: %s\n", err.Error())
	}

	for _, layer := range image.Layers {
		outBase := fmt.Sprintf("%s%s%s", outPrefix, layer.Suffix, outExt)
		outFile := filepath.Join(outDir, outBase)
		layer.processImageLayer(doc, outFile)
	}
}

// Represent the toggles that are applied to a "layer" of an image, which will
// then be exported as an individual instance of that image.
type ImageLayer struct {
	Suffix string `yaml:"suffix"`
	HideIDs []string `yaml:"hide_ids,omitempty"`
	ShowIDs []string `yaml:"show_ids,omitempty"`
}

// Within the context of a specific image layer, hide/show the relevant image
// elements for that particular layer.
func (layer *ImageLayer) processImageLayer(doc *etree.Document, outFile string) {
	for _, id := range layer.HideIDs {
		element := assertOneElementById(doc, id)
		setHidden(element, true)
	}
	for _, id := range layer.ShowIDs {
		element := assertOneElementById(doc, id)
		setHidden(element, false)
	}

	if err := doc.WriteToFile(outFile); err != nil {
		log.Fatalf("Problem writing to %s: %s\n", outFile, err.Error())
	}

	// The input filename, and therefore the output filename, was already
	// checked to end with .svg
	outPng := outFile[0:(len(outFile) - 4)] + ".png"

	cmd := exec.Cmd{
		Path: "/usr/bin/flatpak",
		Args: []string{
			"flatpak",
			"run",
			"org.inkscape.Inkscape",
			fmt.Sprintf("--export-filename=%s", outPng),
			"--export-width=1280",
			"--export-height=720",
			outFile,
		},
	}
	if err := cmd.Run(); err != nil{
		log.Fatalf("Could not convert SVG to PNG with Inkscape: %s\n", err.Error())
	}
}

// Find the singular element that has the given ID attribute. If there isn't
// exactly one of them, then fail the entire program.
func assertOneElementById(doc *etree.Document, id string) *etree.Element {
	xpath := fmt.Sprintf("//[@id='%s']", id)
	elements := doc.FindElements(xpath)
	if len(elements) != 1 {
		log.Fatalf("Expected one #%s element; found %d\n", id, len(elements))
	}
	return elements[0]
}

// Toggle the style: display: X sub-attribute on the element. If true, then set
// display:none; if false, then set display:inline.
func setHidden(element *etree.Element, hidden bool) {
	attrValue := element.SelectAttrValue("style", "")
	attrComponents := strings.Split(attrValue, ";")

	var expectedComponent string
	if hidden {
		expectedComponent = "display:none"
	} else {
		expectedComponent = "display:inline"
	}

	done := false
	for key, component := range attrComponents {
		if strings.HasPrefix(component, "display:") {
			attrComponents[key] = expectedComponent
			done = true
		}
	}

	if !done {
		attrComponents = append(attrComponents, expectedComponent)
	}

	element.CreateAttr("style", strings.Join(attrComponents, ";"))
}

// Main entry point for the program/script.
func main() {
	if len(os.Args) != 3 {
		log.Fatalln("Usage: bulletpointer /path/to/in.yaml /path/to/out/dir")
	}

	if dirStat, err := os.Stat(os.Args[2]); err == nil {
		if !dirStat.IsDir() {
			log.Fatalf("Destination should be a directory: %s\n", os.Args[2])
		}
	} else {
		log.Fatalf("Destination dir needs to exist: %s\n", os.Args[2])
	}

	var yamlImages []*Image
	if yamlBytes, err := os.ReadFile(os.Args[1]); err == nil {
		if err := yaml.Unmarshal(yamlBytes, &yamlImages); err != nil {
			log.Fatalf("Problem parsing YAML: %s\n", err.Error())
		}
	} else {
		log.Fatalf("Problem reading file: %s\n", err.Error())
	}

	for _, yamlImage := range yamlImages {
		yamlImage.processImage(filepath.Dir(os.Args[1]), os.Args[2])
	}
}
