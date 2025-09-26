package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/beevik/etree"
	"gopkg.in/yaml.v3"
)

func assertOneElementById(doc *etree.Document, id string) *etree.Element {
	xpath := fmt.Sprintf("//[@id='%s']", id)
	elements := doc.FindElements(xpath)
	if len(elements) != 1 {
		log.Fatalf("Expected one #%s element; found %d\n", id, len(elements))
	}
	return elements[0]
}

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

func main() {
	const file string = "/home/louis/OneDrive/Videos/Interview Take 3/titles/Question1AllBullets.svg"
	var element *etree.Element

	doc := etree.NewDocument()
	if err := doc.ReadFromFile(file); err != nil {
		panic(err)
	}

	element = assertOneElementById(doc, "line1")
	setHidden(element, true)

	doc.WriteToFile("/tmp/out.svg")
}
