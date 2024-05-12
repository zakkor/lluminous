# lluminous - a fully free, libre, fast, chatbot frontend

### Key features:

- Multiple providers, plug in your API keys (stored entirely locally) and you're good to go
    - OpenRouter (which lets you use all models: OpenAI, Anthropic, 50+ others)
    - Groq
    - Or just directly use your OpenAI or Anthropic API keys (coming soon)

- Tool use. Works with both OpenAI models as well as Groq models that support it. Parallel tool calls are supported.
- Multi-shot prompting
- Support for all available models across all providers
- Change model mid-conversation
- Branching conversation history (like the left-right ChatGPT arrows that you can click to go back to a previous response)

### Hosted instance (no need to install anything):

Available at: https://lluminous.chat

Note: If you want to use tool calls, you *will* need to have the lluminous server running on your machine.

### Privacy:
- Completely private and transparent. All your conversation history and keys are stored entirely locally, and kept only in your browser, on your device.

### Installation:

1. Clone the repository
2. Install and start the client: `npm i && npm run dev`. The client will be accessible at http://localhost:5173
3. Install and start the server: `cd server && go build && PASSWORD="chooseapassword" ./server -sandbox <sandbox_path>`
   - Note: the sandbox currently only works on macOS, since it uses macOS-specific sandboxing features.


