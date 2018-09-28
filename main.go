package main

import (
	"github.com/c-bata/go-prompt"
	"github.com/snowdrop/spring-boot-cloud-devex/cmd"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/common/logger"
	"github.com/spf13/pflag"
	"os"
	"strings"
)

func main() {
	// Enable Debug if env var is defined
	logger.EnableLogLevelDebug()

	// Call commands
	p := prompt.New(
		func(in string) {
			promptArgs := strings.Fields(in)
			os.Args = append([]string{os.Args[0]}, promptArgs...)
			cmd.Execute()
		},
		func(d prompt.Document) []prompt.Suggest {
			return findSuggestions(d)
		},
	)
	p.Run()
}

func findSuggestions(d prompt.Document) []prompt.Suggest {
	command := cmd.RootCmd()
	args := strings.Fields(d.CurrentLine())

	if found, _, err := command.Find(args); err == nil {
		command = found
	}

	var suggestions []prompt.Suggest
	addFlags := func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		if strings.HasPrefix(d.GetWordBeforeCursor(), "--") {
			suggestions = append(suggestions, prompt.Suggest{Text: "--" + flag.Name, Description: flag.Usage})
		} else if strings.HasPrefix(d.GetWordBeforeCursor(), "-") && flag.Shorthand != "" {
			suggestions = append(suggestions, prompt.Suggest{Text: "-" + flag.Shorthand, Description: flag.Usage})
		}

		if suggester, ok := cmd.Suggesters[cmd.GetFlagSuggesterName(command, flag.Name)]; ok {
			suggestions = append(suggestions, suggester.Suggest(d)...)
		}
	}

	command.LocalFlags().VisitAll(addFlags)
	command.InheritedFlags().VisitAll(addFlags)

	if command.HasAvailableSubCommands() {
		for _, c := range command.Commands() {
			if !c.Hidden {
				suggestions = append(suggestions, prompt.Suggest{Text: c.Name(), Description: c.Short})
			}
		}
	}

	if suggester, ok := cmd.Suggesters[cmd.GetCommandSuggesterName(command)]; ok {
		suggestions = append(suggestions, suggester.Suggest(d)...)
	}

	return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
}
