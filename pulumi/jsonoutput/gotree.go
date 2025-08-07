package jsonoutput

// copied from https://github.com/d6o/GoTree and tweaked

import (
	"bufio"
	"strings"
	"unicode/utf8"

	"golang.org/x/exp/utf8string"
)

const (
	newLine      = "\n"
	emptySpace   = "   "
	middleItem   = "├─ "
	continueItem = "│  "
	lastItem     = "└─ "
)

type (
	tree struct {
		text   string
		parent *tree
		col1   string
		col2   string
		col3   string
		items  []Tree
	}

	// Tree is tree interface
	Tree interface {
		Add(text string) Tree
		AddTree(tree Tree)
		Items() []Tree
		Text(bool, int) string
		Print(bool) string
		SetCol1(string)
		SetCol2(string)
		SetCol3(string)
	}

	printer struct {
	}

	// Printer is printer interface
	Printer interface {
		Print(Tree, bool, int) string
	}
)

//NewTree returns a new GoTree.Tree
func NewTree(text string) Tree {
	return &tree{
		text:  text,
		items: []Tree{},
	}
}

//Add adds a node to the tree
func (t *tree) Add(text string) Tree {
	n := &tree{
		text:   text,
		parent: t,
		items:  []Tree{},
	}
	t.items = append(t.items, n)
	return n
}

//AddTree adds a tree as an item
func (t *tree) AddTree(tree Tree) {
	t.items = append(t.items, tree)
}

func (t *tree) Depth() int {
	depth := 0
	ref := t.parent
	for ref != nil {
		depth += 1
		ref = ref.parent
	}
	return depth
}

func truncateOrPad(val string, target int) string {
	valCount := utf8.RuneCountInString(val)
	var res string
	if valCount < target {
		res = val + strings.Repeat(" ", target-valCount)
	} else {
		u := utf8string.NewString(val)
		res = u.Slice(valCount-target, valCount)
	}
	return res
}

//Text returns the node's value
func (t *tree) Text(includeColumns bool, col1Start int) string {
	res := t.text
	padding := ""
	if includeColumns {
		resCount := utf8.RuneCountInString(res)
		paddingSize := max(0, col1Start - resCount - (t.Depth() * 3))
		padding = strings.Repeat(" ", paddingSize)
	}
	if includeColumns && t.col1 != "" {
		res += padding + truncateOrPad(t.col1, 20)
	}
	if includeColumns && t.col2 != "" {
		res += "  " + truncateOrPad(t.col2, 10)
	}
	if includeColumns && t.col3 != "" {
		res += "  " + t.col3
	}
	return res
}

func (t *tree) SetCol1(val string) {
	t.col1 = val
}

func (t *tree) SetCol2(val string) {
	t.col2 = val
}

func (t *tree) SetCol3(val string) {
	t.col3 = val
}

//Items returns all items in the tree
func (t *tree) Items() []Tree {
	return t.items
}

//Print returns an visual representation of the tree
func (t *tree) Print(includeColumns bool) string {
	printer := newPrinter()
	col1Start := 0

	if includeColumns {
		maxLen := 0
		// print the tree once without columns to figure out the max width
		res := printer.Print(t, false, 0)
		scanner := bufio.NewScanner(strings.NewReader(res))
		for scanner.Scan() {
			l := utf8.RuneCountInString(scanner.Text())
			if l > maxLen {
				maxLen = l
			}
		}
		col1Start = maxLen + 2
	}
	return newPrinter().Print(t, includeColumns, col1Start)
}

func newPrinter() Printer {
	return &printer{}
}

//Print prints a tree to a string
func (p *printer) Print(t Tree, includeColumns bool, col1Start int) string {
	return t.Text(includeColumns, col1Start) + newLine + p.printItems(t.Items(), []bool{}, includeColumns, col1Start)
}

func (p *printer) printText(text string, spaces []bool, last bool) string {
	var result string
	for _, space := range spaces {
		if space {
			result += emptySpace
		} else {
			result += continueItem
		}
	}

	indicator := middleItem
	if last {
		indicator = lastItem
	}

	var out string
	lines := strings.Split(text, "\n")
	for i := range lines {
		text := lines[i]
		if i == 0 {
			out += result + indicator + text + newLine
			continue
		}
		if last {
			indicator = emptySpace
		} else {
			indicator = continueItem
		}
		out += result + indicator + text + newLine
	}

	return out
}

func (p *printer) printItems(t []Tree, spaces []bool, includeColumns bool, col1start int) string {
	var result string
	for i, f := range t {
		last := i == len(t)-1
		result += p.printText(f.Text(includeColumns, col1start), spaces, last)
		if len(f.Items()) > 0 {
			spacesChild := append(spaces, last)
			result += p.printItems(f.Items(), spacesChild, includeColumns, col1start)
		}
	}
	return result
}
