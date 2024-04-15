package main

import (
	"os/exec"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mitchellh/go-homedir"
	"github.com/zakkor/server/funcreader"
)

//	{
//		"type": "function",
//		"function": {
//			"name": "ls",
//			"description": "Lists all files in the given directory. If no directory is given, the current directory is used.",
//			"parameters": {
//				"type": "object",
//				"properties": {
//					"path": {
//						"type": "string",
//						"description": "The path of the directory to list."
//					}
//				}
//			}
//		}
//	}
type FunctionSchema struct {
	Type     string `json:"type"`
	Function struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Parameters  struct {
			Type       string `json:"type"`
			Properties map[string]struct {
				Type        string `json:"type"`
				Description string `json:"description"`
			} `json:"properties"`
		} `json:"parameters"`
	} `json:"function"`
}

func NewFunctionSchema(f any) FunctionSchema {
	schema := FunctionSchema{}
	schema.Function.Name = strings.ToLower(funcreader.FuncName(f))

	// Get description and extract function description as well as arguments description
	desc := funcreader.FuncDescription(f)
	// Arguments begin at first "-", description is everything before that first "-"
	splitDesc := strings.Split(desc, "-")
	schema.Type = "function"
	schema.Function.Description = strings.TrimSpace(splitDesc[0])
	// Arguments are everything after the first "-"
	schema.Function.Parameters.Type = "object"
	schema.Function.Parameters.Properties = make(map[string]struct {
		Type        string `json:"type"`
		Description string `json:"description"`
	})
	// Get func arguments types and names:
	types := funcreader.FuncArgumentsType(f)
	for i := 1; i < len(splitDesc); i++ {
		name := strings.TrimSpace(strings.Split(splitDesc[i], ":")[0])
		desc := strings.TrimSpace(strings.Split(splitDesc[i], ":")[1])
		schema.Function.Parameters.Properties[name] = struct {
			Type        string `json:"type"`
			Description string `json:"description"`
		}{
			Type:        types[i-1],
			Description: desc,
		}
	}

	return schema
}

var tools = map[string]func(arguments any) any{
	"ls": func(arguments any) any {
		path, err := homedir.Expand(arguments.(map[string]any)["path"].(string))
		if err != nil {
			panic(err)
		}
		cmd := exec.Command("ls", "-l", path)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return string(out) + err.Error()
		}
		return string(out)
	},
	"cat": func(arguments any) any {
		path, err := homedir.Expand(arguments.(map[string]any)["path"].(string))
		if err != nil {
			panic(err)
		}
		cmd := exec.Command("cat", path)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return string(out) + err.Error()
		}
		return string(out)
	},
	"send_discord_message": func(arguments any) any {
		dg, err := discordgo.New("Bot " + "MTA5MDU3MjQzODE1ODQ0NjYyMw.Gb_6v_.X44hnIGbbHVBbMHesGAw05oiiKZ8ElB1DzFKh8")
		if err != nil {
			panic(err)
		}

		// In this example, we only care about receiving message events.
		dg.Identify.Intents = discordgo.IntentsGuildMessages

		// Open a websocket connection to Discord and begin listening.
		err = dg.Open()
		if err != nil {
			return "Error when sending message: " + err.Error()
		}
		defer dg.Close()

		_, err = dg.ChannelMessageSend("615562052206854340", arguments.(map[string]any)["message"].(string))
		if err != nil {
			return "Error when sending message: " + err.Error()
		}

		return "Message successfully sent."
	},
}

// ls lists all files in the given directory. If no directory is given, the current directory is used.
// - path: The path of the directory to list.
func Ls(path string) string {
	return ""
}
