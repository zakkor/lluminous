export const conversationTemplates = {
	Markdown: {
		model: 'openchat/openchat-7b:free',
		messages: [
			{
				role: 'system',
				content: `You are a helpful assistant. Answer the user's question to the best of your ability.`,
			},
			{ role: 'user', content: 'Give me markdown examples?' },
			{
				role: 'assistant',
				content: `Markdown is a lightweight markup language used for formatting text. Here are some examples of common Markdown elements:
# Markdown
The following provides a quick reference to the most commonly used Markdown syntax.

## Headers
### H3
#### H4
##### H5
###### H6

*Italic* and **Bold** and ~~Scratched Text~~

1. Item 1
2. Item 2
3. Item 3
	- Item 3a
	- Item 3b

Inline equation: $equation$

Display equation: $$equation$$

Display equation:

$$equation$$

- $x + y$
- $x - y$
- $x \times y$ 
- $x \div y$
- $\dfrac{x}{y}$
- $\sqrt{x}$

- $\pi \approx 3.14159$
- $\pm \, 0.2$
- $\dfrac{0}{1} \neq \infty$
- $0 < x < 1$
- $0 \leq x \leq 1$
- $x \geq 10$
- $\forall \, x \in (1,2)$
- $\exists \, x \notin [0,1]$
- $A \subset B$
- $A \subseteq B$
- $A \cup B$
- $A \cap B$
- $X \implies Y$
- $X \impliedby Y$
- $a \to b$
- $a \longrightarrow b$
- $a \Rightarrow b$
- $a \Longrightarrow b$
- $a \propto b$

\`\`\`r
norm <- function(x) {
  sqrt(x%*%x)
}
norm(1:4)
\`\`\`

> Imagination is more important than knowledge.
>
> Albert Einstein
`,
			},
		],
	},
// 	'OpenHermes-2.5': {
// 		model: 'openhermes-2.5-mistral-7b.Q5_K_M.gguf',
// 		messages: [
// 			{
// 				role: 'system',
// 				content: `You are a helpful assistant. Answer the user's question to the best of your ability.`,
// 			},
// 		],
// 	},
// 	'Hermes 2 Pro Function Calling': {
// 		model: 'Hermes-2-Pro-Mistral-7B.Q6_K.gguf',
// 		messages: [
// 			{
// 				role: 'system',
// 				contentFn: (functionSchema) => {
// 					return `You are a function calling AI model. You are provided with function signatures within <tools></tools> XML tags. You may call one or more functions to assist with the user query. Don't make assumptions about what values to plug into functions. Here are the available tools: <tools> ${functionSchema} </tools> Use the following pydantic model json schema for each tool call you will make: {"title": "FunctionCall", "type": "object", "properties": {"arguments": {"title": "Arguments", "type": "object"}, "name": {"title": "Name", "type": "string"}}, "required": ["arguments", "name"]} For each function call return a json object with function name and arguments within <tool_call></tool_call> XML tags as follows:
// <tool_call>
// {"arguments": <args-dict>, "name": <function-name>}
// </tool_call>
// Be very precise about which tool to select. If no function matches, DO NOT use a tool call.
// If needed, and multiple functions match, you may use multiple tool calls at once, one after the other.
// If no function matches, the conversation continues normally.`;
// 				},
// 			},
// 		],
// 	},
// 	'Hermes 2 Pro Function Calling - Example 1': {
// 		model: 'Hermes-2-Pro-Mistral-7B.Q6_K.gguf',
// 		messages: [
// 			{
// 				role: 'system',
// 				content:
// 					'You are a function calling AI model. You are provided with function signatures within <tools></tools> XML tags. You may call one or more functions to assist with the user query. Don\'t make assumptions about what values to plug into functions. Here are the available tools: <tools> [{"type": "function", "function": {"name": "ls", "description": "Lists all files in the given directory. If no directory is given, the current directory is used.", "parameters": {"type": "object", "properties": {"path": {"type": "string", "description": "The path of the directory to list."}}}}}, {"type": "function", "function": {"name": "cat", "description": "Reads the contents of a file.", "parameters": {"type": "object", "properties": {"path": {"type": "string", "description": "The path of the file to read."}}}}}, {"type": "function", "function": {"name": "send_discord_message", "description": "Sends a message on Discord.", "parameters": {"type": "object", "properties": {"message": {"type": "string", "description": "The contents of the message to be sent."}}}}}] </tools> Use the following pydantic model json schema for each tool call you will make: {"title": "FunctionCall", "type": "object", "properties": {"arguments": {"title": "Arguments", "type": "object"}, "name": {"title": "Name", "type": "string"}}, "required": ["arguments", "name"]} For each function call return a json object with function name and arguments within <tool_call></tool_call> XML tags as follows:\n<tool_call>\n{"arguments": <args-dict>, "name": <function-name>}\n</tool_call>\nBe very precise about which tool to select. If no function matches, DO NOT use a tool call.\nIf needed, and multiple functions match, you may use multiple tool calls at once, one after the other.\nIf no function matches, the conversation continues normally.',
// 				pendingContent: '',
// 				submitted: false,
// 				editing: false,
// 				unclosed: false,
// 			},
// 			{
// 				role: 'user',
// 				content: 'List the directory at /Users/ed',
// 				pendingContent: '',
// 				submitted: true,
// 				editing: false,
// 				unclosed: false,
// 			},
// 			{
// 				role: 'assistant',
// 				content: '<tool_call>\n{"arguments": {"path": "/Users/ed"}, "name": "ls"}\n</tool_call>',
// 				pendingContent: '',
// 				submitted: false,
// 				editing: false,
// 				unclosed: true,
// 				generated: true,
// 			},
// 			{
// 				role: 'tool',
// 				content:
// 					'<tool_response>total 48\ndrwxr-xr-x@   9 ed  staff    288 Feb 25 13:19 Applications\ndrwxr-xr-x@   3 ed  staff     96 Apr  7  2023 Applications (Parallels)\ndrwxr-xr-x    7 ed  staff    224 Dec  2 23:15 Calibre Library\ndrwx------+  21 ed  staff    672 Mar 16 00:59 Desktop\ndrwx------+  14 ed  staff    448 Mar 12 13:31 Documents\ndrwx------@ 644 ed  staff  20608 Mar 16 13:48 Downloads\ndrwxr-xr-x    7 ed  staff    224 Sep  3  2023 Games\ndrwx------@ 107 ed  staff   3424 Jan 26 22:25 Library\ndrwx------@  17 ed  staff    544 Mar 16 00:59 Movies\ndrwx------@   6 ed  staff    192 May  3  2023 Music\ndrwxr-xr-x    3 ed  staff     96 Mar 30  2023 Parallels\ndrwx------@  11 ed  staff    352 Dec 26 18:46 Pictures\ndrwxr-xr-x+   4 ed  staff    128 Mar 28  2023 Public\n-rw-------@   1 ed  staff    550 Feb  6 19:27 acme.dump\ndrwxr-xr-x@  43 ed  staff   1376 Jun  1  2023 bin\n-rw-r--r--    1 ed  staff    117 Jan 19 22:37 foo.jimage\ndrwxr-xr-x    4 ed  staff    128 Mar 29  2023 go\n-rw-r--r--    1 ed  staff   1199 Mar 15 19:06 go.mod\n-rw-r--r--    1 ed  staff  10293 Mar 15 19:06 go.sum\ndrwxr-xr-x    3 ed  staff     96 Jan  5 19:22 il\ndrwxr-xr-x@   3 ed  staff     96 May 21  2023 lib\ndrwxr-xr-x    3 ed  staff     96 Dec  3 18:36 nltk_data\ndrwxr-xr-x    6 ed  staff    192 Mar 16 00:55 puzzle\ndrwxr-xr-x   46 ed  staff   1472 Mar 15 18:46 src\n</tool_response>',
// 				pendingContent: '',
// 				submitted: false,
// 				editing: false,
// 				unclosed: false,
// 			},
// 			{
// 				role: 'assistant',
// 				content:
// 					'Here is the directory listing for /Users/ed:\n\n- Applications\n- Applications (Parallels)\n- Calibre Library\n- Desktop\n- Documents\n- Downloads\n- Games\n- Library\n- Movies\n- Music\n- Parallels\n- Pictures\n- Public\n- acme.dump\n- bin\n- foo.jimage\n- go\n- go.mod\n- go.sum\n- il\n- lib\n- nltk_data\n- puzzle\n- src',
// 				pendingContent: '',
// 				submitted: false,
// 				editing: false,
// 				unclosed: true,
// 				generated: true,
// 			},
// 		],
// 	},
// 	'DeepSeek Coder': {
// 		model: 'deepseek-coder-33b-instruct.Q4_K_M.gguf',
// 		messages: [
// 			{
// 				role: 'system',
// 				content: `You are an AI programming assistant, utilizing the Deepseek Coder model, developed by Deepseek Company, and you only answer questions related to computer science. For politically sensitive questions, security and privacy issues, and other non-computer science questions, you will refuse to answer.`,
// 			},
// 		],
// 	},
};
