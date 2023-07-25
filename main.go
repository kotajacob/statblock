package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/muesli/reflow/wordwrap"
	"golang.org/x/net/html"
)

const R20BaseURL = "https://roll20.net/compendium/dnd5e/"

var (
	dot         = regexp.MustCompile(`•`)
	space       = regexp.MustCompile(`\s+`)
	trailing    = regexp.MustCompile(` +\n`)
	doubleblank = regexp.MustCompile(`\n\n\n`)
)

type Monster struct {
	Name        string
	Description string

	Size      string
	Type      string
	Alignment string

	AC    string
	HP    string
	Speed string

	STR string
	DEX string
	CON string
	INT string
	WIS string
	CHA string

	Skills      string
	Saves       string
	Senses      string
	Languages   string
	Challenge   string
	Proficiency string
}

func usage() {
	fmt.Fprintln(
		os.Stderr,
		"usage: statblock [Roll20 URL or Monster Name]",
	)
	fmt.Fprintln(
		os.Stderr,
		"alternatively the URL/Name can be passed from STDIN",
	)
	os.Exit(1)
}

func main() {
	flag.Parse()

	url := strings.Join(flag.Args(), " ")
	if url == "" {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: failed reading STDIN:", err)
			os.Exit(1)
		}
		url = strings.TrimSpace(string(b))
	}

	// Accept a monster's name instead of a full URL.
	if !strings.Contains(url, R20BaseURL) {
		url = R20BaseURL + url
	}

	c := colly.NewCollector(colly.AllowedDomains("roll20.net"))

	var m Monster
	c.OnHTML(".page-title", m.getName)
	c.OnHTML("#pagecontent", m.getDescription)
	c.OnHTML(".attrListItem", m.getAttrs)
	c.Visit(url)

	if m.Name == "" {
		fmt.Fprintln(os.Stderr, "unknown monster")
		os.Exit(1)
	}

	fmt.Println(m)
}

// getName parses and sets the monster's name.
func (m *Monster) getName(e *colly.HTMLElement) {
	m.Name = strings.TrimSpace(e.Text)
}

// getDescription parses and sets the monster's name.
func (m *Monster) getDescription(e *colly.HTMLElement) {
	m.Description = trailing.ReplaceAllString(renderHTML(e.DOM), "\n")
}

func renderHTML(s *goquery.Selection) string {
	var buf bytes.Buffer

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(
				space.ReplaceAllString(
					n.Data,
					" ",
				),
			)
		}
		if n.Type == html.ElementNode && n.Data == "br" {
			buf.WriteString("\n")
		}
		if n.Type == html.ElementNode && n.Data == "h2" {
			buf.WriteString("\n# ")
		}
		if n.PrevSibling != nil {
			p := n.PrevSibling
			if p.Type == html.ElementNode {
				if p.Data == "h2" {
					buf.WriteString("\n")
				}
			}
		}
		if n.FirstChild != nil {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
	}
	for _, n := range s.Nodes {
		f(n)
	}

	return dot.ReplaceAllString(
		doubleblank.ReplaceAllString(
			wordwrap.String(
				strings.TrimSpace(buf.String()),
				80,
			),
			"\n\n",
		),
		"-",
	)
}

func (m *Monster) getAttrs(e *colly.HTMLElement) {
	attr := strings.TrimSpace(e.ChildText(".attrName"))
	value := strings.TrimSpace(e.ChildText(".attrValue"))
	switch attr {
	case "Size":
		m.Size = value
	case "Type":
		m.Type = value
	case "Alignment":
		m.Alignment = value
	case "AC":
		m.AC = value
	case "HP":
		m.HP = value
	case "Speed":
		m.Speed = value
	case "STR":
		m.STR = value
	case "DEX":
		m.DEX = value
	case "CON":
		m.CON = value
	case "INT":
		m.INT = value
	case "WIS":
		m.WIS = value
	case "CHA":
		m.CHA = value
	case "Skills":
		m.Skills = value
	case "Saving Throws":
		m.Saves = value
	case "Passive Perception":
		// Are there other senses? Should we use a slice instead of a string?
		m.Senses = "passive Perception " + value
	case "Languages":
		m.Languages = value
	case "Challenge Rating":
		m.Challenge = value
	case "Proficiency":
		m.Proficiency = value
	}
}

func (m Monster) String() string {
	var b strings.Builder
	b.WriteString(m.Description)
	b.WriteString("\n\n")

	b.WriteString("# Stats\n")
	b.WriteString(m.Size)
	b.WriteString(" ")
	b.WriteString(m.Type)
	b.WriteString(" ")
	b.WriteString(m.Alignment)

	b.WriteString("\n```\n")
	b.WriteString("Armor Class ")
	b.WriteString(m.AC)
	b.WriteString("\n")
	b.WriteString("Hit Points ")
	b.WriteString(m.HP)
	b.WriteString("\n")
	b.WriteString("Speed ")
	b.WriteString(m.Speed)
	b.WriteString("\n")
	b.WriteString("STR ")
	b.WriteString(m.STR)
	b.WriteString("\n")
	b.WriteString("DEX ")
	b.WriteString(m.DEX)
	b.WriteString("\n")
	b.WriteString("CON ")
	b.WriteString(m.CON)
	b.WriteString("\n")
	b.WriteString("INT ")
	b.WriteString(m.INT)
	b.WriteString("\n")
	b.WriteString("WIS ")
	b.WriteString(m.WIS)
	b.WriteString("\n")
	b.WriteString("CHA ")
	b.WriteString(m.CHA)
	b.WriteString("\n```\n")

	if m.Skills != "" {
		b.WriteString("*Skills* ")
		b.WriteString(m.Skills)
		b.WriteString("\n")
	}
	if m.Senses != "" {
		b.WriteString("*Senses* ")
		b.WriteString(m.Senses)
		b.WriteString("\n")
	}
	if m.Languages != "" {
		b.WriteString("*Languages* ")
		b.WriteString(m.Languages)
		b.WriteString("\n")
	}
	b.WriteString("*Challenge* ")
	b.WriteString(m.Challenge)
	return b.String()
}
