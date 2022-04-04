package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rivo/tview"
)

var (
	keyValRX = regexp.MustCompile(`\A(\s*)([\w|\-|\.|\/|\s]+):\s(.+)\z`)
	keyRX    = regexp.MustCompile(`\A(\s*)([\w|\-|\.|\/|\s]+):\s*\z`)
)

const (
	yamlFullFmt  = "%s[key::b]%s[colon::-]: [val::]%s"
	yamlKeyFmt   = "%s[key::b]%s[colon::-]:"
	yamlValueFmt = "[val::]%s"
)

func colorizeYAML(raw string) string {
	lines := strings.Split(tview.Escape(raw), "\n")

	fullFmt := strings.Replace(yamlFullFmt, "[key", "["+"magenta", 1)
	fullFmt = strings.Replace(fullFmt, "[colon", "["+"blue", 1)
	fullFmt = strings.Replace(fullFmt, "[val", "["+"foreground", 1)

	keyFmt := strings.Replace(yamlKeyFmt, "[key", "["+"magenta", 1)
	keyFmt = strings.Replace(keyFmt, "[colon", "["+"blue", 1)

	valFmt := strings.Replace(yamlValueFmt, "[val", "["+"foreground", 1)

	buff := make([]string, 0, len(lines))
	for _, l := range lines {
		res := keyValRX.FindStringSubmatch(l)
		if len(res) == 4 {
			buff = append(buff, fmt.Sprintf(fullFmt, res[1], res[2], res[3]))
			continue
		}

		res = keyRX.FindStringSubmatch(l)
		if len(res) == 3 {
			buff = append(buff, fmt.Sprintf(keyFmt, res[1], res[2]))
			continue
		}

		buff = append(buff, fmt.Sprintf(valFmt, l))
	}

	return strings.Join(buff, "\n")
}
