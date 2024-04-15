package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"github.com/mitchellh/go-homedir"
)

type Role int

const (
	RoleSystem Role = iota
	RoleUser
	RoleAssistant
	RoleTool
)

type Message struct {
	Role    Role
	Content string
}

type PromptFormat int

const (
	ChatML PromptFormat = iota
	Cerebrum
)

type Conversation struct {
	Format      PromptFormat
	Messages    []Message
	precomplete string
}

func Convo() Conversation {
	return Conversation{Format: ChatML}
}

func (p Conversation) Template(format PromptFormat) Conversation {
	p.Format = format
	return p
}

func (p Conversation) System(content string) Conversation {
	p.Messages = append(p.Messages, Message{RoleSystem, content})
	return p
}

func (p Conversation) User(content string) Conversation {
	p.Messages = append(p.Messages, Message{RoleUser, content})
	return p
}

func (p Conversation) Assistant(content string) Conversation {
	p.Messages = append(p.Messages, Message{RoleAssistant, content})
	return p
}

func (p Conversation) Tool(content string) Conversation {
	p.Messages = append(p.Messages, Message{RoleTool, content})
	return p
}

func (p Conversation) Precomplete(content string) Conversation {
	p.precomplete = content
	return p
}

func (p Conversation) String() string {
	switch p.Format {
	case ChatML:
		return p.ChatML()
	case Cerebrum:
		return p.Cerebrum()
	}
	panic("unknown template")
}

func (p Conversation) ChatML() string {
	var b bytes.Buffer
	for _, msg := range p.Messages {
		b.WriteString("<|im_start|>")
		switch msg.Role {
		case RoleSystem:
			b.WriteString("system")
		case RoleUser:
			b.WriteString("user")
		case RoleAssistant:
			b.WriteString("assistant")
		case RoleTool:
			b.WriteString("tool")
		}
		b.WriteString("\n")
		b.WriteString(msg.Content)
		b.WriteString("<|im_end|>\n")
	}
	b.WriteString("<|im_start|>assistant\n" + p.precomplete)
	return b.String()
}

// Cerebrum format:
// <s>A chat between a user and a thinking artificial intelligence assistant. The assistant describes its thought process and gives helpful and detailed answers to the user's questions.
// User: Are you conscious?
// AI:
func (p Conversation) Cerebrum() string {
	var b bytes.Buffer
	for _, msg := range p.Messages {
		switch msg.Role {
		case RoleSystem:
			b.WriteString("<s>")
		case RoleUser:
			b.WriteString("User: ")
		case RoleAssistant:
			b.WriteString("AI: ")
		}
		b.WriteString(msg.Content)
		b.WriteString("\n")
	}
	b.WriteString("AI:" + p.precomplete)
	return b.String()
}

func (p Conversation) Stop() []string {
	switch p.Format {
	case Cerebrum:
		return []string{"</s>"}
	case ChatML:
		return []string{"<|im_end|>"}
	}
	panic("unknown template")
}

type CompleteRequest struct {
	Prompt      string   `json:"prompt"`
	NPredict    int      `json:"n_predict"`
	Grammar     string   `json:"grammar"`
	Stop        []string `json:"stop"`
	Temperature float64  `json:"temperature"`
}

type CompleteResponse struct {
	Content         string `json:"content"`
	TokensEvaluated int    `json:"tokens_evaluated"`
}

func (p Conversation) Complete(server *LlamaServer) Conversation {
	return p.CompleteOpts(server, CompleteOptions{})
}

type CompleteOptions struct {
	Temperature float64
	NPredict    int
	Grammar     string
	Stop        []string
}

func (p Conversation) CompleteOpts(srv *LlamaServer, opts CompleteOptions) Conversation {
	cmpltReq := CompleteRequest{
		Prompt:   p.String(),
		Stop:     append(p.Stop(), opts.Stop...),
		NPredict: -1,
	}
	if opts.Temperature != 0 {
		cmpltReq.Temperature = opts.Temperature
	}
	if opts.NPredict != 0 {
		cmpltReq.NPredict = opts.NPredict
	}
	if opts.Grammar != "" {
		cmpltReq.Grammar = opts.Grammar
	}

	// JSON data to send with POST request
	jsonData, err := json.Marshal(cmpltReq)
	if err != nil {
		panic(err)
	}

	// Create a new request using http
	req, err := http.NewRequest("POST", srv.Addr+"/completion", bytes.NewReader(jsonData))
	if err != nil {
		panic(err)
	}

	// Add headers to the request
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var cmpltResp map[string]any
	err = json.Unmarshal(body, &cmpltResp)
	if err != nil {
		panic(err)
	}

	if cmpltResp["stop"].(bool) && cmpltResp["stopped_word"].(bool) && (cmpltResp["stopping_word"].(string) == "</tool_call>") {
		cmpltResp["content"] = cmpltResp["content"].(string) + "</tool_call>"
		// Parse the response and extract what is inside <tool_call>...</tool_call>:
		toolCallContent := extractToolCallContent(cmpltResp["content"].(string))
		if toolCallContent != "" {
			var toolcallMap map[string]any
			err = json.Unmarshal([]byte(toolCallContent), &toolcallMap)
			if err != nil {
				panic(err)
			}

			// Call the tool
			toolName := toolcallMap["name"].(string)
			toolArgs := toolcallMap["arguments"]
			tool, ok := tools[toolName]
			if !ok {
				panic("unknown tool: " + toolName)
			}
			content := tool(toolArgs)
			contentJSON, err := json.Marshal(content)
			if err != nil {
				panic(err)
			}
			toolconvo := p.Assistant(cmpltResp["content"].(string)).Tool(fmt.Sprintf(`<tool_response>
{"name": "%s", "content": %s}
</tool_response>`, toolName, contentJSON))
			return toolconvo.CompleteOpts(srv, opts)
		}
	}

	return p.Assistant(cmpltResp["content"].(string))
}

var tools = map[string]func(arguments any) any{
	"ls": func(arguments any) any {
		path, err := homedir.Expand(arguments.(map[string]any)["path"].(string))
		if err != nil {
			panic(err)
		}
		cmd := exec.Command("ls", path)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return string(out) + err.Error()
		}
		return string(out)
	},
}

func extractToolCallContent(content string) string {
	startTag := "<tool_call>"
	endTag := "</tool_call>"

	startIndex := strings.Index(content, startTag)
	endIndex := strings.Index(content, endTag)

	if startIndex != -1 && endIndex != -1 && endIndex > startIndex+len(startTag) {
		return content[startIndex+len(startTag) : endIndex]
	}

	return ""
}

func (p Conversation) LastMessage() string {
	return p.Messages[len(p.Messages)-1].Content
}
