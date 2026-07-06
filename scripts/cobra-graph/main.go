package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	appcmd "github.com/anynines/a9s-cli-v2/cmd"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

const (
	formatMarkdown = "markdown"
	formatASCII    = "ascii"
)

func main() {
	outputPath := flag.String("o", "", "Output markdown file path (default: stdout)")
	includeHidden := flag.Bool("include-hidden", false, "Include hidden commands")
	format := flag.String("format", formatMarkdown, "Output format: markdown|ascii")
	descMode := flag.String("print-descriptions", "", "Include command descriptions in ASCII output: short|long (only applies to ASCII format)")
	lineSpacing := flag.Bool("line-spacing", true, "Include empty lines between nodes in ASCII output (only applies to ASCII format), default: true)")
	flag.Parse()

	root := appcmd.RootCommand()

	if *format != formatMarkdown && *format != formatASCII {
		fmt.Fprintf(os.Stderr, "invalid --format value %q, use: markdown|ascii\n", *format)
		os.Exit(1)
	}

	if *descMode != "" && *descMode != "short" && *descMode != "long" {
		fmt.Fprintf(os.Stderr, "invalid --print-descriptions value %q, use: short|long\n", *descMode)
		os.Exit(1)
	}

	var output string
	switch *format {
	case formatMarkdown:
		output = buildMarkdownGraph(root, *includeHidden)
	case formatASCII:
		output = buildASCIIGraph(root, *includeHidden, *descMode, *lineSpacing)
	}

	if *outputPath == "" {
		fmt.Print(output)
		return
	}

	if err := os.WriteFile(*outputPath, []byte(output), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write output file %q: %v\n", *outputPath, err)
		os.Exit(1)
	}
}

func buildMarkdownGraph(root *cobra.Command, includeHidden bool) string {
	var nodes, edges strings.Builder
	counter := 0
	writeMermaidNode(&nodes, &edges, root, "", &counter, includeHidden)

	var b strings.Builder
	b.WriteString("# Cobra Command Graph\n\n")
	b.WriteString("```mermaid\n")
	b.WriteString("flowchart LR\n")
	b.WriteString(nodes.String())
	b.WriteString(edges.String())
	b.WriteString("```\n")
	return b.String()
}

func writeMermaidNode(nodes, edges *strings.Builder, cmd *cobra.Command, parentID string, counter *int, includeHidden bool) {
	id := fmt.Sprintf("n%d", *counter)
	*counter++

	fmt.Fprintf(nodes, "  %s[\"%s\"]\n", id, escapeMermaidLabel(cmd.CommandPath()))
	if parentID != "" {
		fmt.Fprintf(edges, "  %s --> %s\n", parentID, id)
	}

	for _, child := range sortedChildren(cmd, includeHidden) {
		writeMermaidNode(nodes, edges, child, id, counter, includeHidden)
	}
}

func buildASCIIGraph(root *cobra.Command, includeHidden bool, descMode string, lineSpacing bool) string {
	var b strings.Builder
	b.WriteString(root.Name())
	b.WriteByte('\n')

	children := sortedChildren(root, includeHidden)
	for i, child := range children {
		writeASCIINode(&b, child, "", i == len(children)-1, includeHidden, descMode, lineSpacing)
	}

	return b.String()
}

func writeASCIINode(b *strings.Builder, cmd *cobra.Command, prefix string, isLast bool, includeHidden bool, descriptionMode string, lineSpacing bool) {
	connector := "├── "
	nextPrefix := prefix + "│   "
	if isLast {
		connector = "└── "
		nextPrefix = prefix + "    "
	}

	b.WriteString(prefix)
	b.WriteString(connector)
	b.WriteString(cmd.Name())
	children := sortedChildren(cmd, includeHidden)
	b.WriteByte('\n')
	if len(children) == 0 && descriptionMode != "" {
		b.WriteString(commandDescription(descriptionMode, prefix, utf8.RuneCountInString(connector), cmd, isLast, lineSpacing))
	}
	for i, child := range children {
		writeASCIINode(b, child, nextPrefix, i == len(children)-1, includeHidden, descriptionMode, lineSpacing)
	}
}

func commandDescription(descriptionMode, prefix string, connectorLen int, cmd *cobra.Command, isLast, lineSpacing bool) string {
	if !isLast {
		prefix += "│"
		connectorLen -= 1
	}
	description := ""
	if descriptionMode == "short" {
		description = cmd.Short
	}
	if descriptionMode == "long" {
		description = cmd.Long
	}
	if description != "" {
		description = strings.ReplaceAll(description, "\n", " ")
		lines := splitEvery(description, 80-utf8.RuneCountInString(prefix)-connectorLen)
		description = ""
		for _, line := range lines {
			description += prefix + strings.Repeat(" ", connectorLen) + line + "\n"
		}

		if lineSpacing {
			description += prefix + "\n"
		}
		// description = "\n" + strings.Repeat(" ", utf8.RuneCountInString(prefix)+utf8.RuneCountInString(connector)) + description
		// description = " (" + strings.ToLower(description[:1]) + description[1:] + ")"
	}
	return description
}

func sortedChildren(parent *cobra.Command, includeHidden bool) []*cobra.Command {
	children := make([]*cobra.Command, 0, len(parent.Commands()))
	for _, child := range parent.Commands() {
		if child.Hidden && !includeHidden {
			continue
		}
		children = append(children, child)
	}

	sort.Slice(children, func(i, j int) bool {
		return children[i].Name() < children[j].Name()
	})

	return children
}

func escapeMermaidLabel(label string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "\"", "\\\"")
	return replacer.Replace(label)
}

func splitEvery(s string, n int) []string {
	var result []string
	result = append(result, s)
	// words := strings.Fields(s)
	// currentLine := ""
	// for _, word := range words {
	// 	if utf8.RuneCountInString(currentLine)+utf8.RuneCountInString(word)+1 > n {
	// 		result = append(result, currentLine)
	// 		currentLine = word
	// 		continue
	// 	}

	// 	if currentLine != "" {
	// 		currentLine += " "
	// 	}
	// 	currentLine += word
	// }

	// if currentLine != "" {
	// 	result = append(result, currentLine)
	// }

	return result
}
