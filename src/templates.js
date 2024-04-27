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
};
