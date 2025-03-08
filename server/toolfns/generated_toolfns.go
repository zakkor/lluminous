// generated @ 2025-03-08T22:45:48+02:00 by gendoc
package toolfns

import "github.com/noonien/codoc"

func init() {
	codoc.Register(codoc.Package{
		ID:   "github.com/zakkor/server/toolfns",
		Name: "toolfns",
		Doc:  "generated @ 2025-03-08T22:36:30+02:00 by gendoc",
		Functions: map[string]codoc.Function{
			"NewGroup": {
				Name: "NewGroup",
				Args: []string{
					"name",
					"fns",
				},
			},
			"Shell": {
				Name: "Shell",
				Doc:  "Executes the given bash command and returns the output of the command.\ncommand: The bash command to execute.",
				Args: []string{
					"command",
				},
			},
			"init": {
				Name: "init",
			},
		},
		Structs: map[string]codoc.Struct{
			"ContentTypeResponse": {
				Name: "ContentTypeResponse",
			},
			"Group": {
				Name: "Group",
			},
		},
	})
}
