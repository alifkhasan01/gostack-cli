package cli

import (
	"fmt"
	"strings"

	"github.com/alifkhasan01/gostack-cli/internal/generator"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:     "generate <type> <name>",
	Aliases: []string{"g"},
	Short:   "Generate a new file (handler, service, repository, migration, module)",
	Example: `  gostack generate handler User
  gostack generate service User
  gostack generate repository User
  gostack generate migration create_users
  gostack generate module User
  gostack generate crud Product
  gostack g crud Order`,
	Args: cobra.ExactArgs(2),
	RunE: runGenerate,
}

func runGenerate(_ *cobra.Command, args []string) error {
	kindStr := strings.ToLower(args[0])
	name := args[1]

	kindMap := map[string]generator.Kind{
		"handler":    generator.KindHandler,
		"service":    generator.KindService,
		"repository": generator.KindRepository,
		"repo":       generator.KindRepository,
		"migration":  generator.KindMigration,
		"migrate":    generator.KindMigration,
		"module":     generator.KindModule,
		"mod":        generator.KindModule,
		"crud":       generator.KindCRUD,
	}

	kind, ok := kindMap[kindStr]
	if !ok {
		return fmt.Errorf(
			"unknown type '%s'. Available: handler, service, repository, migration, module, crud",
			kindStr,
		)
	}

	return generator.Generate(kind, name)
}
