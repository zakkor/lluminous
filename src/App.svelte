<script>
	import { v4 as uuidv4 } from 'uuid';
	import { onMount, tick } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { complete, generateImage } from './convo.js';
	import KnobsSidebar from './KnobsSidebar.svelte';
	import Button from './Button.svelte';
	import { marked } from 'marked';
	import markedKatex from './marked-katex-extension';

	import { persisted } from './localstorage.js';
	import { getRelativeDate } from './date.js';
	import { compressAndEncode, decodeAndDecompress } from './share.js';
	import {
		formatModelName,
		providers,
		hasCompanyLogo,
		openAIAdditionalModelsMultimodal,
		openAIImageGenerationModels,
		anthropicModels,
		openAIIgnoreIds,
		priorityOrder,
	} from './providers.js';
	import ModelSelector from './ModelSelector.svelte';
	import CompanyLogo from './CompanyLogo.svelte';
	import {
		controller,
		remoteServer,
		openaiAPIKey,
		groqAPIKey,
		openrouterAPIKey,
		config,
		params,
		toolSchema,
		anthropicAPIKey,
	} from './stores.js';
	import SettingsModal from './SettingsModal.svelte';
	import ToolcallButton from './ToolcallButton.svelte';
	import MessageContent from './MessageContent.svelte';
	import Toolcall from './Toolcall.svelte';
	import Modal from './Modal.svelte';
	import Icon from './Icon.svelte';
	import {
		feArrowUp,
		feCheck,
		feCheckCircle,
		feChevronLeft,
		feChevronRight,
		feCpu,
		feEdit2,
		feMenu,
		feMessageCircle,
		feMoreHorizontal,
		fePaperclip,
		fePlus,
		feRefreshCw,
		feSettings,
		feShare,
		feSliders,
		feSquare,
		feTrash,
		feUser,
		feX,
	} from './feather.js';
	import { defaultToolSchema } from './tools.js';
	import { debounce, readFileAsDataURL } from './util.js';
	import FilePreview from './FilePreview.svelte';

	marked.use(
		markedKatex({
			throwOnError: false,
		})
	);

	const convoId = persisted('convoId');
	let convos = {};

	// Initialize history and convo with a blank slate while we wait for IndexedDB to start
	const defaultConvo = {
		id: uuidv4(),
		time: Date.now(),
		model: { id: null, name: 'Select a model', provider: null },
		messages: [],
		versions: {},
		tools: [],
	};
	let convo = defaultConvo;

	let db;
	const request = indexedDB.open('lluminous', 2);
	request.onupgradeneeded = (event) => {
		const db = event.target.result;
		if (!db.objectStoreNames.contains('messages')) {
			db.createObjectStore('messages', { keyPath: 'id' });
		}
		if (!db.objectStoreNames.contains('conversations')) {
			db.createObjectStore('conversations', { keyPath: 'id' });
		}
	};
	request.onsuccess = async (event) => {
		db = event.target.result;
		await fetchAllConversations();

		const restored = await restoreConversation();

		if (!restored) {
			if (!$convoId || !convos[$convoId]) {
				newConversation();
			} else {
				convo = convos[$convoId];
			}
			if (!convo.tools) {
				convo.tools = [];
				saveConversation(convo);
			}
		}
	};
	request.onerror = (event) => {
		console.error(event.target.error);
	};

	const saveMessage = debounce((msg) => {
		const transaction = db.transaction(['messages'], 'readwrite');
		const store = transaction.objectStore('messages');

		store.put(msg);

		transaction.onerror = () => {
			console.error('Message save failed', transaction.error);
		};
	}, 100);

	async function fetchAllConversations() {
		const transaction = db.transaction(['conversations', 'messages'], 'readonly');
		const conversationsStore = transaction.objectStore('conversations');
		const messagesStore = transaction.objectStore('messages');

		const fetchConversations = new Promise((resolve, reject) => {
			const conversationsRequest = conversationsStore.getAll();
			conversationsRequest.onsuccess = (event) => {
				const conversations = event.target.result;
				const convosData = {};
				conversations.forEach((conversation) => {
					convosData[conversation.id] = conversation;
				});
				resolve(convosData);
			};
			conversationsRequest.onerror = (event) => {
				reject(event.target.error);
			};
		});

		const fetchMessages = (convosData) => {
			return new Promise((resolve, reject) => {
				const messagesRequest = messagesStore.getAll();
				messagesRequest.onsuccess = (event) => {
					const messages = event.target.result;

					messages.forEach((message) => {
						// Migrate old tool_call_id to toolcallId
						if ('tool_call_id' in message) {
							message.toolcallId = message.tool_call_id;
							delete message.tool_call_id;
							saveMessage(message);
						}

						for (let cid in convosData) {
							const index = convosData[cid].messages.indexOf(message.id);
							if (index !== -1) {
								// Replace the message ID with the actual message object
								convosData[cid].messages[index] = message;
							}
							// Handle versions
							for (let versionKey in convosData[cid].versions) {
								for (let messages of convosData[cid].versions[versionKey]) {
									if (!messages) {
										continue;
									}
									const i = messages.indexOf(message.id);
									if (i !== -1) {
										// Replace the message ID with the actual message object
										messages[i] = message;
									}
								}
							}
						}
					});
					resolve(convosData);
				};
				messagesRequest.onerror = (event) => {
					reject(event.target.error);
				};
			});
		};

		try {
			const convosData = await fetchConversations;
			const updatedConvos = await fetchMessages(convosData);
			convos = updatedConvos;
		} catch (error) {
			console.error('Error fetching history:', error);
		}
	}

	const saveConversation = debounce((convo) => {
		const transaction = db.transaction(['conversations'], 'readwrite');
		const store = transaction.objectStore('conversations');
		store.put({
			...convo,
			messages: convo.messages.map((msg) => msg.id),
			versions: Object.fromEntries(
				Object.entries(convo.versions).map(([key, value]) => [
					key,
					value.map((messages) => {
						if (!messages) {
							return null;
						}
						return messages.map((msg) => {
							return msg.id;
						});
					}),
				])
			),
		});

		transaction.onerror = () => {
			console.error('Conversation save failed', transaction.error);
		};
	}, 100);

	function deleteConversation(convo) {
		const transaction = db.transaction(['conversations'], 'readwrite');
		const store = transaction.objectStore('conversations');

		store.delete(convo.id);

		transaction.onerror = () => {
			console.error('Conversation delete failed', transaction.error);
		};
	}

	$: isMultimodal =
		convo.model.modality === 'text+image->text' ||
		openAIAdditionalModelsMultimodal.includes(convo.model.id);

	let historyBuckets = [];
	$: {
		historyBuckets = [];
		for (const entry of Object.values(convos).sort((a, b) => b.time - a.time)) {
			if (entry.shared || isNaN(new Date(entry.time).getTime())) {
				continue;
			}

			const bucketKey = getRelativeDate(entry.time);

			const existingBucket = historyBuckets.find((bucket) => bucket.relativeDate === bucketKey);
			if (!existingBucket) {
				historyBuckets.push({ relativeDate: bucketKey, convos: [entry] });
			} else {
				existingBucket.convos.push(entry);
			}
		}
		historyBuckets.sort((a, b) => b.convos[0].time - a.convos[0].time);
	}

	let content = '';
	let imageUrls = [];
	let imageUrlsBlacklist = [];
	const imageUrlRegex = /https?:\/\/[^\s]+?\.(png|jpe?g)(?=\s|$)/gi;
	let pendingFiles = [];

	let generating = false;

	let historyOpen = false;
	let knobsOpen = false;

	let activeToolcall = null;

	let scrollableEl = null;
	let textareaEls = [];
	let inputTextareaEl;
	let fileInputEl;

	let innerWidth = window.innerWidth;

	$: splitView = innerWidth > 1215 && activeToolcall && !$config.explicitToolView;

	let settingsModalOpen = false;
	let toolcallModalOpen = false;

	function submitEdit(i) {
		// Update the ID of the edited message:
		if (convo.messages[i].submitted || convo.messages[i].generated) {
			let vid = null;
			const msgBeforeEdit = { ...convo.messages[i] };
			msgBeforeEdit.editing = false;
			msgBeforeEdit.pendingContent = '';
			if (!msgBeforeEdit.vid) {
				vid = uuidv4();
				msgBeforeEdit.vid = vid;
			}
			saveMessage(msgBeforeEdit);

			saveVersion(msgBeforeEdit, i);

			convo.messages[i].id = uuidv4();
			if (!convo.messages[i].vid) {
				convo.messages[i].vid = vid;
			}
			convo.messages[i].editing = false;
			convo.messages[i].content = convo.messages[i].pendingContent;
			convo.messages[i].pendingContent = '';
			saveMessage(convo.messages[i]);
		}

		convo.messages = convo.messages.slice(0, i + 1);
		saveConversation(convo);

		submitCompletion();
	}

	async function submitCompletion(insertUnclosed = true) {
		if (!convo.model.provider) {
			const msg = {
				id: uuidv4(),
				role: 'assistant',
				error: 'No model selected. Please add at least one API key and select a model to begin.',
				content: '',
			};
			saveMessage(msg);
			convo.messages.push(msg);
			convo.messages = convo.messages;
			saveConversation(convo);
			return;
		}

		if (generating) {
			$controller.abort();
		}

		generating = true;

		if (insertUnclosed) {
			const msg = {
				id: uuidv4(),
				role: 'assistant',
				content: '',
				unclosed: true,
				generated: true,
				model: convo.model,
			};
			convo.messages.push(msg);
			convo.messages = convo.messages;

			saveMessage(msg);
			saveConversation(convo);
		}

		for (let i = 0; i < convo.messages.length; i++) {
			if (convo.messages[i].editing) {
				convo.messages[i].content = convo.messages[i].pendingContent;
			}
			convo.messages[i].pendingContent = '';
			convo.messages[i].editing = false;
			convo.messages[i].submitted = true;

			saveMessage(convo.messages[i]);
		}
		await tick();
		scrollableEl.scrollTop = scrollableEl.scrollHeight;

		const i = convo.messages.length - 1;

		if (convo.model.modality === 'text->image') {
			await generateImage(convo, {
				oncomplete: (resp) => {
					convo.messages[i].generatedImageUrl = resp.data[0].url;
					generating = false;
					saveMessage(convo.messages[i]);
				},
			});
			return;
		}

		const onupdate = async (chunk) => {
			if (chunk.error) {
				convo.messages[i].error = chunk.error.message || chunk.error;
				generating = false;
				saveMessage(convo.messages[i]);
				return;
			}

			if (convo.model.provider === 'Local') {
				convo.messages[i].content += chunk.content;
				saveMessage(convo.messages[i]);
			} else {
				if (chunk.choices.length === 0) {
					convo.messages[i].error = 'Refused to respond.';
					generating = false;
					saveMessage(convo.messages[i]);
					return;
				}

				const choice = chunk.choices[0];

				if (choice.delta.content) {
					convo.messages[i].content += choice.delta.content;
					saveMessage(convo.messages[i]);
				}

				if (choice.delta.tool_calls) {
					if (!convo.messages[i].toolcalls) {
						convo.messages[i].toolcalls = [];
						saveMessage(convo.messages[i]);
					}

					for (const tool_call of choice.delta.tool_calls) {
						let index = tool_call.index;
						// Watch out! Anthropic tool call indices are 1-based, not 0-based, when message.content is involved.
						if (
							(convo.model.provider === 'Anthropic' || convo.model.id.startsWith('anthropic/')) &&
							convo.messages[i].content
						) {
							index--;
						}

						if (!convo.messages[i].toolcalls[index]) {
							convo.messages[i].toolcalls[index] = {
								id: tool_call.id,
								name: tool_call.function.name,
								arguments: '',
								expanded: true,
							};
							activeToolcall = convo.messages[i].toolcalls[index];
							toolcallModalOpen = true;
						}
						if (tool_call.function.arguments) {
							convo.messages[i].toolcalls[index].arguments += tool_call.function.arguments;
							activeToolcall = convo.messages[i].toolcalls[index];
						}
						saveMessage(convo.messages[i]);
					}
				}
			}

			// Scroll to bottom if we're at or near the bottom of the conversation:
			if (scrollableEl.scrollHeight - scrollableEl.scrollTop - scrollableEl.clientHeight < 100) {
				scrollableEl.scrollTop = scrollableEl.scrollHeight;
			}

			// Check for stoppage:
			if (convo.model.provider === 'Local' && chunk.stop) {
				generating = false;
				return;
			}

			// External tool calls:
			if (
				convo.model.provider !== 'Local' &&
				chunk.choices &&
				(chunk.choices[0].finish_reason === 'stop' ||
					chunk.choices[0].finish_reason === 'end_turn' ||
					chunk.choices[0].finish_reason === 'tool_calls')
			) {
				generating = false;

				// Toolcall arguments are now finalized, we can parse them:
				if (convo.messages[i].toolcalls) {
					let toolPromises = [];

					for (let ti = 0; ti < convo.messages[i].toolcalls.length; ti++) {
						const toolcall = convo.messages[i].toolcalls[ti];

						try {
							toolcall.arguments = JSON.parse(toolcall.arguments);
						} catch (err) {
							convo.messages[i].error =
								'Failed to parse tool call arguments: ' + err + toolcall.arguments;
							saveMessage(convo.messages[i]);
							return;
						}

						// Do we have a client-side tool for this?
						const clientGroup = $toolSchema.find((g) => g.name === 'Client-side');
						const clientToolIndex = clientGroup?.schema.findIndex(
							(t) => t.clientDefinition && t.clientDefinition.name === toolcall.name
						);
						if (clientGroup && clientToolIndex !== -1 && convo.tools.includes(toolcall.name)) {
							const clientTool = clientGroup.schema[clientToolIndex];
							const clientFn = new Function('args', clientTool.clientDefinition.body);
							const result = clientFn(toolcall.arguments);
							toolPromises.push(Promise.resolve(result));
						} else {
							// Otherwise, call server-side tool
							const promise = fetch(`${$remoteServer.address}/tool`, {
								method: 'POST',
								// credentials: 'include',
								headers: {
									Authorization: `Basic ${$remoteServer.password}`,
								},
								body: JSON.stringify({
									id: toolcall.id,
									chat_id: convo.id,
									name: toolcall.name,
									arguments: toolcall.arguments,
								}),
							}).then((resp) => {
								// Mark tool call as finished to we can display it nicely in the UI
								// (still need to await all tool calls to deliver the final response).
								convo.messages[i].toolcalls[ti].finished = true;
								saveMessage(convo.messages[i]);

								return resp.text().then((text) => {
									return JSON.parse(text);
								});
							});

							toolPromises.push(promise);
						}
					}

					const toolResponses = await Promise.all(toolPromises);

					for (let ti = 0; ti < toolResponses.length; ti++) {
						const msg = {
							id: uuidv4(),
							role: 'tool',
							toolcallId: convo.messages[i].toolcalls[ti].id,
							name: convo.messages[i].toolcalls[ti].name,
							content: toolResponses[ti],
						};
						convo.messages.push(msg);
						convo.messages = convo.messages;
						saveMessage(msg);
						saveConversation(convo);
					}

					submitCompletion();
				}

				return;
			}
		};

		const onabort = () => {
			generating = false;
		};

		complete(convo, onupdate, onabort);
	}

	async function insertSystemPrompt() {
		const msg = { id: uuidv4(), role: 'system', content: '', editing: true };
		convo.messages.unshift(msg);
		convo.messages = convo.messages;
		await tick();
		textareaEls[0].focus();

		saveMessage(msg);
		saveConversation(convo);
	}

	function autoresizeTextarea() {
		inputTextareaEl.style.height = 'auto';
		inputTextareaEl.style.height = inputTextareaEl.scrollHeight + 2 + 'px';
	}

	async function sendMessage() {
		if (content.length > 0) {
			if (
				$params.customInstructions &&
				convo.messages.length === 0 &&
				!convo.messages.find((m) => m.role === 'system')
			) {
				const systemMsg = {
					id: uuidv4(),
					role: 'system',
					customInstructions: true,
					content: $params.customInstructions,
				};
				convo.messages.push(systemMsg);
				convo.messages = convo.messages;
				saveMessage(systemMsg);
				saveConversation(convo);
			}

			const msg = {
				id: uuidv4(),
				role: 'user',
				content: content,
				submitted: true,
			};

			const imageUrlMapper = (url) => ({
				type: 'image_url',
				image_url: {
					url,
					detail: 'high',
				},
			});

			if (imageUrls.length > 0) {
				msg.contentParts = [...imageUrls.map(imageUrlMapper)];
			}

			let fileContent = '';
			if (pendingFiles.length > 0) {
				for (const file of pendingFiles) {
					fileContent += `\`\`\` filename="${file.name}"
${file.text}
\`\`\`

`;
				}

				msg.content = fileContent + msg.content;
			}

			convo.messages.push(msg);
			convo.messages = convo.messages;
			await tick();
			scrollableEl.scrollTop = scrollableEl.scrollHeight;

			saveMessage(msg);
			saveConversation(convo);

			content = '';
			imageUrls = [];
			imageUrlsBlacklist = [];
			pendingFiles = [];

			await tick();
			if (innerWidth < 880) {
				inputTextareaEl.blur();
			}
			autoresizeTextarea();

			submitCompletion();
		}
	}

	function cleanShareLink() {
		const params = new URLSearchParams(window.location.search);
		if (params.has('s') || params.has('sl')) {
			window.history.pushState('', document.title, window.location.pathname);
		}
	}

	function newConversation() {
		cleanShareLink();
		activeToolcall = null;

		// if (convo.messages.length === 0) {
		// 	historyOpen = false;
		// 	inputTextareaEl.focus();
		// 	return;
		// }

		const existingNewConvo = Object.values(convos).find((convo) => convo.messages.length === 0);
		if (existingNewConvo) {
			const oldModel = convo.model;
			$convoId = existingNewConvo.id;
			convo = convos[$convoId];
			convo.model = oldModel;

			historyOpen = false;
			inputTextareaEl.focus();
			return;
		}

		const convoData = {
			id: uuidv4(),
			time: Date.now(),
			model: convo.model || models.find((m) => m.id === 'meta-llama/llama-3-8b-instruct'),
			messages: [],
			versions: {},
			tools: [],
		};
		$convoId = convoData.id;
		convos[convoData.id] = convoData;
		convo = convoData;

		saveConversation(convo);

		historyOpen = false;
		if (innerWidth > 880) {
			inputTextareaEl.focus();
		}
	}

	// Split history at this point:
	function saveVersion(message, i) {
		if (!convo.versions) {
			convo.versions = {};
		}
		if (!convo.versions[message.vid]) {
			convo.versions[message.vid] = [null];
		}
		const nullIdx = convo.versions[message.vid].findIndex((v) => v === null);
		convo.versions[message.vid][nullIdx] = [
			structuredClone(message),
			...structuredClone(convo.messages.slice(i + 1)).map((m) => {
				return {
					...m,
					editing: false,
					pendingContent: '',
				};
			}),
		];
		convo.versions[message.vid].push(null);
		saveConversation(convo);
	}

	function shiftVersion(dir, message, i) {
		const activeVersionIndex = convo.versions[message.vid].findIndex((v) => v === null);
		const newVersionIndex = activeVersionIndex + dir;

		convo.versions[message.vid][activeVersionIndex] = convo.messages.slice(i);

		const newMessages = convo.versions[message.vid][newVersionIndex];

		convo.messages = convo.messages.slice(0, i).concat(newMessages);

		convo.versions[message.vid][newVersionIndex] = null;

		saveConversation(convo);
	}

	function closeSidebars(event) {
		if (
			historyOpen &&
			!event.target.closest('[data-sidebar="history"]') &&
			!event.target.closest('[data-trigger="history"]')
		) {
			historyOpen = false;
		}
		if (
			knobsOpen &&
			!event.target.closest('[data-trigger="knobs"]') &&
			!event.target.closest('[data-sidebar="knobs"]')
		) {
			knobsOpen = false;
		}
	}

	async function shareConversation() {
		const sharePromise = new Promise(async (resolve) => {
			const encoded = await compressAndEncode({
				model: convo.model,
				messages: convo.messages,
			});
			const share = `${window.location.protocol}//${window.location.host}/?s=${encoded}`;
			if (share.length > 200) {
				const data = new FormData();
				data.append('pwd', 'muie_webshiti');
				data.append('f:1', new Blob([encoded], { type: 'text/plain' }), 'content.txt');
				const response = await fetch(`https://llum.kilonova.ro`, {
					method: 'POST',
					body: data,
				});

				const shortenedLink = await response.text();
				resolve(
					`${window.location.protocol}//${window.location.host}/?sl=${shortenedLink.split('/').reverse()[1]}`
				);
			} else {
				resolve(`${window.location.protocol}//${window.location.host}/?s=${encoded}`);
			}
		});

		try {
			const clipboardItem = new ClipboardItem({
				'text/plain': sharePromise,
			});
			await navigator.clipboard.write([clipboardItem]);
		} catch (err) {
			await navigator.clipboard.writeText(await sharePromise);
		}
	}

	async function restoreConversation() {
		const params = new URLSearchParams(window.location.search);
		let share;
		if (params.has('s')) {
			share = params.get('s');
		} else if (params.has('sl')) {
			const response = await fetch(`https://llum.kilonova.ro/p/${params.get('sl')}/`);
			share = await response.text();
		}
		if (!share) {
			return false;
		}

		try {
			let decoded = await decodeAndDecompress(share);
			if (Array.isArray(decoded)) {
				decoded = { name: 'Shared conversation', messages: decoded };
			}
			let id = uuidv4();
			const existingShared = Object.values(convos).find((convo) => convo.shared);
			if (existingShared) {
				id = existingShared.id;
			}
			const convoData = {
				id,
				time: Date.now(),
				shared: true,
				model: decoded.model,
				messages: decoded.messages,
				versions: {},
				tools: [],
			};
			$convoId = convoData.id;
			convos[convoData.id] = convoData;
			convo = convoData;
			return true;
		} catch (err) {
			console.error('Error decoding shared conversation:', err);
		}
	}

	let loadedModel = null;
	let loadingModel = false;
	let modelFinishedLoading = [];

	// For local models, we need to tell the server to load them:
	async function loadModel(newModel) {
		loadingModel = true;
		loadedModel = null;

		await fetch(`${$remoteServer.address}/model`, {
			method: 'POST',
			headers: {
				Authorization: `Basic ${$remoteServer.password}`,
			},
			body: JSON.stringify({
				model: newModel.id,
			}),
		});
		loadedModel = newModel;
		if (modelFinishedLoading) {
			modelFinishedLoading[0]();
			modelFinishedLoading[1]();
		}
		loadingModel = false;
	}

	let models = [];

	async function fetchLoadedModel() {
		try {
			const response = await fetch(`${$remoteServer.address}/model`, {
				method: 'GET',
				headers: {
					Authorization: `Basic ${$remoteServer.password}`,
				},
			});
			if (!response.ok) {
				return;
			}
			const json = await response.json();

			loadedModel = {
				provider: 'Local',
				id: json.model,
				// Strip .gguf suffix:
				name: json.model.replace(/\.gguf$/, ''),
			};
		} catch (error) {}
	}

	let loading = false;

	async function fetchModels() {
		loading = true;
		try {
			const promises = providers.map((provider) => {
				if (!provider.apiKeyFn() && provider.name !== 'Local') {
					return [];
				}
				// Anthropic doesn't support the /v1/models endpoint, so we hardcode it:
				if (provider.name === 'Anthropic') {
					return anthropicModels;
				}

				return fetch(`${provider.url}/v1/models`, {
					method: 'GET',
					headers: {
						Authorization:
							provider.name === 'Local'
								? `Basic ${provider.apiKeyFn()}`
								: `Bearer ${provider.apiKeyFn()}`,
					},
				})
					.then((response) => response.json())
					.then((json) => {
						const externalModels = json.data.map((m) => ({
							id: m.id,
							name: m.name || m.id,
							provider: provider.name,
							modality:
								m.architecture?.modality ||
								(openAIImageGenerationModels.includes(m.id) ? 'text->image' : undefined),
						}));
						return externalModels;
					})
					.catch(() => {
						console.log('Error fetching models from provider', provider.name);
						return [];
					});
			});

			const ignoreIds = [...openAIIgnoreIds];

			const results = await Promise.all(promises);
			const externalModels = results.flat().filter((m) => !ignoreIds.includes(m.id));

			function getPriorityIndex(model) {
				for (let i = 0; i < priorityOrder.length; i++) {
					const rule = priorityOrder[i];
					if (rule.exactly) {
						const exactIndex = rule.exactly.indexOf(model.id);
						if (exactIndex !== -1) {
							return [i, exactIndex];
						}
					}
					if (rule.startsWith) {
						for (let j = 0; j < rule.startsWith.length; j++) {
							if (model.id.startsWith(rule.startsWith[j])) {
								if (rule.exactlyNot && rule.exactlyNot.includes(model.id)) {
									continue;
								}
								return [i, j];
							}
						}
					}
					if (rule.fromProvider && model.provider === rule.fromProvider) {
						if (rule.exactlyNot && rule.exactlyNot.includes(model.id)) {
							continue;
						}
						return [i, -1];
					}
				}
				return [priorityOrder.length, -1];
			}

			externalModels.sort((a, b) => {
				const [aIndex, aExactIndex] = getPriorityIndex(a);
				const [bIndex, bExactIndex] = getPriorityIndex(b);

				if (aIndex === bIndex) {
					if (aExactIndex === bExactIndex) {
						return a.id.localeCompare(b.id);
					}
					return aExactIndex - bExactIndex;
				}
				return aIndex - bIndex;
			});

			models = externalModels;
		} catch (error) {
			console.error('Error:', error);
		} finally {
			loading = false;
		}
	}

	function initializePWAStyles() {
		if (window.matchMedia('(display-mode: standalone)').matches || window.navigator.standalone) {
			document.body.classList.add('standalone');
		}
	}

	onMount(async () => {
		// Clear old deprecated local storage data:
		localStorage.removeItem('tools');
		localStorage.removeItem('toolSchema');
		localStorage.removeItem('history');
		localStorage.removeItem('remoteServerAddress');

		// Populate params with default values, in case of old data:
		if ($params.messagesContextLimit == null) {
			$params.messagesContextLimit = 0;
		}

		// Init client tools with default values
		if ($toolSchema.length === 0 && !window.localStorage.getItem('initializedClientTools')) {
			$toolSchema = defaultToolSchema;
			window.localStorage.setItem('initializedClientTools', 'true');
		}

		initializePWAStyles();

		// Async
		fetchLoadedModel();
		fetchModels();
	});

	// For displaying compact tools, we need to collapse sequences of Assistant and Tool messages into a single message
	// inside which we'll display all the tool calls.
	// Returns a list of ranges of messages containing the start and end indices of messages that should be collapsed.
	let collapsedRanges = [];
	$: if (!$config.explicitToolView) {
		collapsedRanges = [];
		let range = { starti: null, endi: null };
		for (let i = convo.messages.length - 1; i >= 0; i--) {
			if (
				convo.messages[i].role === 'tool' ||
				(convo.messages[i].role === 'assistant' &&
					convo.messages[i].toolcalls &&
					i !== convo.messages.length - 1)
			) {
				if (range.endi === null) {
					range.endi = i + 1;
				}
			} else {
				if (range.endi !== null) {
					range.starti = i + 1;
					collapsedRanges.push(range);
					collapsedRanges = collapsedRanges;
					range = { starti: null, endi: null };
				}
			}
		}
	}

	// Sequences of tool calls which are not interrupted by messages also containing text content will be displayed on the same line.
	function collapsedToolcalls(collapsedRange, collapsedMessages, ci, message) {
		if (ci < collapsedMessages.length - 1) {
			// Don't show duplicated toolcalls if these toolcalls will
			// be collapsed into a single line in a later message.
			const nextMessage = collapsedMessages[ci + 1];
			if (nextMessage.role === 'assistant' && nextMessage.toolcalls && !nextMessage.content) {
				return [];
			}
		}

		const i = convo.messages.findIndex((m) => m.id === message.id);

		// Starting from the message `i`, and going backwards, collect all `.toolcalls` until
		// we are interrupted by a message that contains `.content`, or we reach `collapsedRange.starti`
		const toolcalls = [];
		for (let j = i; j >= collapsedRange.starti; j--) {
			const msgIter = convo.messages[j];
			if (msgIter.role === 'assistant' && msgIter.toolcalls) {
				toolcalls.push(msgIter.toolcalls);
			}
			if (msgIter.role === 'assistant' && msgIter.content) {
				break;
			}
		}
		return toolcalls.reverse().flat();
	}
</script>

<svelte:window
	bind:innerWidth
	on:touchstart={closeSidebars}
	on:click={closeSidebars}
	on:keydown={(event) => {
		if (
			event.key === 'Escape' &&
			generating &&
			convo.messages.filter((msg) => msg.generated).length > 0
		) {
			$controller.abort();
		}
	}}
/>

<main class="flex h-dvh w-screen flex-col">
	<div class="flex h-12 items-center gap-1 border-b border-slate-200 px-2 py-1 md:hidden">
		<button
			on:click={newConversation}
			class="flex rounded-full p-2 transition-colors hover:bg-gray-100"
		>
			<Icon icon={fePlus} strokeWidth={3} class="ml-auto h-4 w-4 text-slate-700" />
		</button>
		<button
			data-trigger="history"
			class="flex rounded-full p-2 transition-colors hover:bg-gray-100"
			on:click={() => (historyOpen = !historyOpen)}
		>
			<Icon icon={feMenu} strokeWidth={3} class="m-auto h-4 w-4 text-slate-700" />
		</button>

		{#if !convo.shared}
			<ModelSelector
				{convo}
				{models}
				{loadingModel}
				{loadedModel}
				bind:modelFinishedLoading={modelFinishedLoading[0]}
				on:change={({ detail }) => {
					convo.model = detail;
					saveConversation(convo);
					if (convo.model.provider === 'Local' && convo.model.id !== loadedModel.id) {
						loadModel(convo.model);
					}
				}}
				on:setTools={({ detail }) => {
					convo.tools = convo.tools.concat(detail);
					saveConversation(convo);
				}}
				on:unsetTools={({ detail }) => {
					convo.tools = convo.tools.filter((t) => !detail.includes(t));
					saveConversation(convo);
				}}
				on:clearTools={() => {
					convo.tools = [];
					saveConversation(convo);
				}}
				class="!absolute left-1/2 z-[99] -translate-x-1/2"
			/>
		{:else if convo.model}
			<p
				class="!absolute left-1/2 line-clamp-1 flex -translate-x-1/2 items-center gap-x-2 whitespace-nowrap text-sm font-semibold"
			>
				<CompanyLogo model={convo.model} size="w-4 h-4" />
				{formatModelName(convo.model)}
			</p>
		{/if}

		<button
			class="ml-auto flex rounded-full p-2 transition-colors hover:bg-gray-100"
			on:click={shareConversation}
		>
			<Icon icon={feShare} strokeWidth={3} class="m-auto h-4 w-4 text-slate-700" />
		</button>
		<button
			data-trigger="knobs"
			class="flex rounded-full p-2 transition-colors hover:bg-gray-100"
			on:click={() => (knobsOpen = !knobsOpen)}
		>
			<Icon icon={feSliders} strokeWidth={3} class="m-auto h-4 w-4 text-slate-700" />
		</button>
	</div>
	<div class="relative flex h-full flex-1 overflow-hidden">
		<aside
			data-sidebar="history"
			class="{historyOpen
				? ''
				: '-translate-x-full'} fixed top-0 z-[100] flex h-full w-[230px] flex-col border-r bg-white pl-3 pt-4 transition-transform duration-300 ease-in-out md:static md:translate-x-0"
		>
			<div class="mb-1 pr-3">
				<button
					on:click={newConversation}
					class="flex w-full items-center rounded-lg border py-2.5 pl-3 pr-4 text-left text-sm font-medium hover:bg-gray-100"
				>
					New chat
					<Icon icon={fePlus} strokeWidth={3} class="ml-auto h-3.5 w-3.5 text-slate-700" />
				</button>
			</div>
			<ol
				class="flex list-none flex-col overflow-y-auto pb-3 pr-3 pt-5 !scrollbar-white scrollbar-slim hover:!scrollbar-slim"
			>
				{#each historyBuckets as { relativeDate, convos: historyConvos } (relativeDate)}
					<li class="mb-2 ml-3 text-xs font-medium text-slate-600 [&:not(:first-child)]:mt-6">
						{relativeDate}
					</li>
					{#each historyConvos as historyConvo (historyConvo.id)}
						<li class="group relative">
							<button
								on:click={() => {
									if (generating) {
										$controller.abort();
									}

									historyOpen = false;
									activeToolcall = null;

									$convoId = historyConvo.id;
									convo = convos[$convoId];

									cleanShareLink();
								}}
								class="{$convoId === historyConvo.id
									? 'bg-gray-100'
									: ''} leading-0 w-full rounded-lg px-3 py-2 text-left text-sm group-hover:bg-gray-100"
							>
								<span class="line-clamp-1">
									{historyConvo.messages.length === 0
										? 'New conversation'
										: historyConvo.messages
												.find((m) => m.role === 'user')
												?.content.split(' ')
												.slice(0, 5)
												.join(' ') || 'Untitled'}
								</span>
							</button>
							<button
								on:click={() => {
									if ($convoId === historyConvo.id) {
										const newId = Object.values(convos).find(
											(e) => e.id !== historyConvo.id && !e.shared
										)?.id;
										if (!newId) {
											newConversation();
										} else {
											$convoId = newId;
										}
									}
									delete convos[historyConvo.id];
									convos = convos;
									deleteConversation(historyConvo);
								}}
								class="z-1 absolute right-0 top-0 flex h-full w-12 rounded-br-lg rounded-tr-lg bg-gradient-to-l {$convoId ===
								historyConvo.id
									? 'from-gray-100'
									: 'from-white group-hover:from-gray-100'} from-65% to-transparent pr-3 transition-opacity sm:from-gray-100 sm:opacity-0 sm:group-hover:opacity-100"
							>
								<Icon icon={feTrash} class="m-auto mr-0 h-3 w-3 shrink-0 text-slate-700" />
							</button>
						</li>
					{/each}
				{/each}
			</ol>

			<div class="settings-trigger-container -ml-3 mt-auto flex pb-3">
				<button
					data-trigger="settings"
					class="mx-3 flex flex-1 items-center gap-x-4 rounded-lg border border-slate-200 px-4 py-3 text-left text-sm font-medium hover:bg-gray-100"
					on:click={() => {
						historyOpen = false;
					}}
				>
					Settings
					<Icon icon={feSettings} class="ml-auto h-4 w-4 text-slate-700" />
				</button>
			</div>
		</aside>
		<div class="flex flex-1 flex-col">
			<div class="relative hidden items-center border-b border-slate-200 px-2 py-1 md:flex">
				{#if !convo.shared}
					<ModelSelector
						{convo}
						{models}
						{loadingModel}
						{loadedModel}
						bind:modelFinishedLoading={modelFinishedLoading[1]}
						on:change={({ detail }) => {
							convo.model = detail;
							saveConversation(convo);
							if (convo.model.provider === 'Local' && convo.model.id !== loadedModel.id) {
								loadModel(convo.model);
							}
						}}
						on:setTools={({ detail }) => {
							convo.tools = convo.tools.concat(detail);
							saveConversation(convo);
						}}
						on:unsetTools={({ detail }) => {
							convo.tools = convo.tools.filter((t) => !detail.includes(t));
							saveConversation(convo);
						}}
						on:clearTools={() => {
							convo.tools = [];
							saveConversation(convo);
						}}
						class="!absolute left-1/2 z-[99] -translate-x-1/2"
					/>
				{:else if convo.model}
					<p
						class="!absolute left-1/2 flex -translate-x-1/2 items-center gap-x-2 text-sm font-semibold"
					>
						<CompanyLogo model={convo.model} size="w-4 h-4" />
						{formatModelName(convo.model)}
					</p>
				{/if}

				<button
					class="ml-auto flex rounded-full p-3 transition-colors hover:bg-gray-100"
					on:click={shareConversation}
				>
					<Icon icon={feShare} strokeWidth={3} class="m-auto h-4 w-4 text-slate-700" />
				</button>
				<button
					data-trigger="knobs"
					class="flex rounded-full p-3 transition-colors hover:bg-gray-100"
					on:click={() => (knobsOpen = !knobsOpen)}
				>
					<Icon icon={feSliders} strokeWidth={3} class="m-auto h-4 w-4 text-slate-700" />
				</button>
			</div>

			<div class="flex h-full w-full">
				<!-- svelte-ignore a11y-no-static-element-interactions -->
				<div
					class="{splitView
						? 'w-[50%]'
						: 'w-full'} relative max-h-[calc(100vh-49px)] transition-[width] duration-500 ease-in-out"
					on:dragover={(event) => {
						event.preventDefault();
					}}
					on:drop={async (event) => {
						event.preventDefault();

						let filenames = [];
						let promises = [];
						if (event.dataTransfer.items) {
							// Use DataTransferItemList interface to access the file(s)
							[...event.dataTransfer.items].forEach((item, _) => {
								// If dropped items aren't files, reject them
								if (item.kind !== 'file') {
									return;
								}

								const file = item.getAsFile();
								filenames.push(file.name);
								promises.push(file.text());
							});
						} else {
							// Use DataTransfer interface to access the file(s)
							[...event.dataTransfer.files].forEach((file, _) => {
								filenames.push(file.name);
								promises.push(file.text());
							});
						}

						const texts = await Promise.all(promises);
						for (let i = 0; i < texts.length; i++) {
							const text = texts[i];
							const filename = filenames[i];
							pendingFiles.push({ name: filename, text: text });
							pendingFiles = pendingFiles;
						}

						tick().then(() => {
							autoresizeTextarea();
						});
					}}
				>
					<div
						bind:this={scrollableEl}
						class="{splitView
							? 'scrollbar-none'
							: 'scrollbar-ultraslim'} scrollable flex h-full w-full flex-col overflow-y-auto pb-[80px]"
					>
						{#if convo.messages.length > 0}
							<ul
								class="{splitView
									? 'rounded-br-lg border-r'
									: ''} mb-3 flex w-full !list-none flex-col divide-y divide-slate-200/50 border-b border-slate-200/50"
							>
								{#each convo.messages as message, i (message.id)}
									{@const showMessage =
										(['user', 'assistant'].includes(message.role) ||
											(message.role === 'system' &&
												(!message.customInstructions ||
													(message.customInstructions && message.showCustomInstructions)))) &&
										($config.explicitToolView ||
											!collapsedRanges.some((r) => i >= r.starti && i < r.endi))}
									{#if showMessage}
										{@const hasLogo = hasCompanyLogo(message.model)}
										<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-noninteractive-element-interactions -->
										<li
											data-role={message.role}
											class="{!message.generated &&
											!message.submitted &&
											message.role !== 'system' &&
											message.role !== 'tool'
												? 'bg-yellow-50/40'
												: message.role === 'assistant'
													? 'bg-slate-50/30'
													: ''} group relative px-5 pb-10 pt-6 ld:px-8"
											style="z-index: {convo.messages.length - i};"
											on:touchstart={(event) => {
												// Make click trigger hover on mobile:
												event.target.dispatchEvent(new MouseEvent('mouseenter'));
											}}
										>
											{#if i === 0 && message.role !== 'system'}
												<Button
													variant="outline"
													class="absolute left-1/2 top-0 z-[98] -translate-x-1/2 rounded-t-none !border-t-0 border-dashed text-xs opacity-0 transition-opacity group-hover:opacity-100"
													on:click={insertSystemPrompt}
												>
													<Icon icon={feMessageCircle} class="mr-2 h-3 w-3 text-slate-600" />
													Add system prompt
												</Button>
											{:else if i === 1 && convo.messages[i - 1].role === 'system' && convo.messages[i - 1].customInstructions && !convo.messages[i - 1].showCustomInstructions}
												<Button
													variant="outline"
													class="absolute left-1/2 top-0 z-[98] -translate-x-1/2 rounded-t-none !border-t-0 text-xs opacity-0 transition-opacity group-hover:opacity-100"
													on:click={() => {
														convo.messages[i - 1].showCustomInstructions = true;
														saveMessage(convo.messages[i - 1]);
													}}
												>
													<Icon icon={feEdit2} class="mr-2 h-3 w-3 text-slate-600" />
													Custom instructions
												</Button>
											{/if}
											<div
												class="relative mx-auto flex w-full max-w-[680px] gap-x-3.5 self-start md:gap-x-5 ld:max-w-[768px]"
											>
												<button
													disabled={message.role === 'system'}
													on:click={() => {
														// Toggle between user and assistant:
														if (message.role === 'user') {
															message.role = 'assistant';
														} else {
															message.role = 'user';
														}
													}}
													class="flex h-8 w-8 shrink-0 rounded-md md:h-9 md:w-9 md:rounded-[7px] {message.role ===
													'system'
														? 'border border-teal-200 bg-teal-100'
														: message.role === 'user'
															? 'border border-slate-200 bg-white'
															: message.role === 'assistant' && !hasLogo
																? 'border border-teal-200 bg-teal-100 pb-px'
																: ''}"
												>
													{#if message.role === 'assistant' && hasLogo}
														<CompanyLogo
															model={message.model}
															size="w-full h-full"
															rounded="rounded-[inherit]"
														/>
													{:else}
														<span class="m-auto">
															{#if message.role === 'system'}
																<Icon icon={feMessageCircle} class="h-4 w-4 text-slate-800" />
															{:else if message.role === 'assistant'}
																<Icon icon={feCpu} class="h-4 w-4 text-slate-800" />
															{:else}
																<Icon icon={feUser} class="h-4 w-4 text-slate-800" />
															{/if}
														</span>
													{/if}
												</button>

												<!-- svelte-ignore a11y-no-static-element-interactions -->
												{#if message.editing}
													<textarea
														bind:this={textareaEls[i]}
														class="w-full resize-none border-none bg-transparent p-0 leading-[28px] text-slate-800 outline-none focus:ring-0"
														rows={1}
														bind:value={message.pendingContent}
														on:keydown={(event) => {
															if (
																event.key === 'Enter' &&
																!event.shiftKey &&
																message.role === 'user' &&
																message.submitted &&
																message.pendingContent &&
																message.content !== message.pendingContent
															) {
																event.preventDefault();
																submitEdit(i);
																event.target.blur();
															}
														}}
														on:input={(event) => {
															// Resize textarea as content grows:
															event.target.style.height = 'auto';
															event.target.style.height = event.target.scrollHeight + 'px';
														}}
													/>
												{:else}
													<div class="flex w-full flex-col gap-6">
														{#if !$config.explicitToolView}
															{@const collapsedRange = collapsedRanges.find((r) => i === r.endi)}
															{#if collapsedRange}
																{@const collapsedMessages = convo.messages
																	.slice(collapsedRange.starti, collapsedRange.endi)
																	.filter((m) => m.role === 'assistant')}
																{#if collapsedMessages.length > 0}
																	{#each collapsedMessages as message, ci}
																		{@const toolcallsOnLine = collapsedToolcalls(
																			collapsedRange,
																			collapsedMessages,
																			ci,
																			message
																		)}

																		<MessageContent {message} />

																		{#if toolcallsOnLine?.length > 0}
																			<div class="-mb-1 flex flex-wrap gap-3 [&:first-child]:mt-1">
																				{#each toolcallsOnLine as toolcall, ti}
																					{@const toolresponse = convo.messages.find(
																						(msg) => msg.toolcallId === toolcall.id
																					)}
																					<ToolcallButton
																						{toolcall}
																						{toolresponse}
																						active={toolcall.id === activeToolcall?.id}
																						on:click={() => {
																							activeToolcall = toolcall;
																						}}
																					/>
																				{/each}
																			</div>
																		{/if}
																	{/each}
																{/if}
															{/if}
														{/if}

														{#if generating && message.role === 'assistant' && i === convo.messages.length - 1 && message.content === '' && !message.toolcalls}
															<div
																class="mt-2 h-3 w-3 shrink-0 animate-bounce rounded-full bg-slate-600"
															/>
														{/if}

														<MessageContent {message} />

														{#if !$config.explicitToolView && message.toolcalls?.length > 0}
															<div class="-mb-1 flex flex-wrap gap-3 [&:first-child]:mt-1">
																{#each message.toolcalls as toolcall, ti}
																	{@const toolresponse = convo.messages.find(
																		(msg) => msg.toolcallId === toolcall.id
																	)}
																	<ToolcallButton
																		{toolcall}
																		{toolresponse}
																		active={toolcall.id === activeToolcall?.id}
																		on:click={() => {
																			activeToolcall = toolcall;
																		}}
																	/>
																{/each}
															</div>
														{/if}

														<!-- OAI toolcalls will always be at the end -->
														{#if message.toolcalls && $config.explicitToolView}
															{#each message.toolcalls as toolcall, ti}
																{@const toolresponse = convo.messages.find(
																	(msg) => msg.toolcallId === toolcall.id
																)}
																<Toolcall
																	{toolcall}
																	{toolresponse}
																	class="mb-1"
																	on:click={() => {
																		convo.messages[i].toolcalls[ti].expanded =
																			!convo.messages[i].toolcalls[ti].expanded;
																		saveMessage(convo.messages[i]);
																	}}
																/>
															{/each}
														{/if}
													</div>
												{/if}

												{#if message.editing}
													<div class="absolute -bottom-8 right-1 flex gap-x-1 md:right-0">
														{#if convo.messages.filter((msg) => msg.role !== 'system' && !msg.submitted).length >= 2 && i === convo.messages.length - 1 && message.role !== 'assistant'}
															<button
																class="flex items-center gap-x-1 rounded-full bg-green-100 px-3 py-2"
																on:click={() => {
																	submitCompletion();
																}}
															>
																<Icon icon={feCheckCircle} class="h-3.5 w-3.5 text-slate-600" />
																<span class="text-xs text-slate-600"> Submit all </span>
															</button>
														{/if}
														{#if message.role !== 'assistant' && message.pendingContent && message.pendingContent !== message.content}
															<button
																class="flex items-center gap-x-1 rounded-full bg-green-50 px-3 py-2 hover:bg-green-100"
																on:click={(event) => {
																	if (message.role === 'system') {
																		// If system message, accept the edit instead of submitting at point:
																		convo.messages[i].content = message.pendingContent;
																		convo.messages[i].pendingContent = '';
																		convo.messages[i].editing = false;
																		saveMessage(convo.messages[i]);
																		return;
																	}
																	submitEdit(i);
																	event.target.blur();
																}}
															>
																<Icon icon={feCheck} class="h-3.5 w-3.5 text-slate-600" />
																<span class="text-xs text-slate-600">
																	{#if message.role === 'system'}
																		Set system prompt
																	{:else}
																		Submit
																	{/if}
																</span>
															</button>
														{/if}
														{#if message.role === 'assistant' && message.pendingContent && message.pendingContent !== message.content && message.content !== '...' && i === convo.messages.length - 1}
															<button
																class="flex items-center gap-x-1.5 rounded-full bg-green-50 px-3 py-2 hover:bg-green-100"
																on:click={async () => {
																	convo.messages[i].unclosed = true;
																	saveMessage(convo.messages[i]);
																	submitCompletion(false);
																}}
															>
																<Icon icon={feMoreHorizontal} class="h-3.5 w-3.5 text-slate-600" />
																<span class="text-xs text-slate-600">Pre-filled response</span>
															</button>
														{/if}
														<button
															class="flex items-center gap-x-1 rounded-full bg-gray-50 px-3 py-2 hover:bg-gray-100"
															on:click={() => {
																convo.messages[i].editing = false;
																convo.messages[i].pendingContent = '';
																saveMessage(convo.messages[i]);
															}}
														>
															<Icon icon={feX} class="h-3.5 w-3.5 text-slate-600" />
															<span class="text-xs text-slate-600">Cancel</span>
														</button>
													</div>
												{/if}
												{#if !message.editing}
													<div
														class="absolute bottom-[-32px] left-11 flex items-center gap-x-4 md:bottom-[-28px] md:left-14"
													>
														{#if message.role === 'user' && convo.versions?.[message.vid]}
															{@const versions = convo.versions[message.vid]}
															{@const versionIndex = versions.findIndex((v) => v === null)}
															<div class="flex items-center md:gap-x-1">
																<button
																	class="group flex h-6 w-6 shrink-0 rounded-full md:h-3 md:w-3"
																	disabled={versionIndex === 0}
																	on:click={() => {
																		shiftVersion(-1, message, i);
																	}}
																>
																	<Icon
																		icon={feChevronLeft}
																		class="m-auto h-3.5 w-3.5 text-slate-800 group-disabled:text-slate-500 md:h-3 md:w-3"
																	/>
																</button>
																<span class="text-xs tabular-nums">
																	{versionIndex + 1} / {versions.length}
																</span>
																<button
																	class="group flex h-6 w-6 shrink-0 rounded-full md:h-3 md:w-3"
																	disabled={versionIndex === versions.length - 1}
																	on:click={() => {
																		shiftVersion(1, message, i);
																	}}
																>
																	<Icon
																		icon={feChevronRight}
																		class="m-auto h-3.5 w-3.5 text-slate-800 group-disabled:text-slate-500 md:h-3 md:w-3"
																	/>
																</button>
															</div>
														{/if}

														{#if (message.role === 'assistant' && i > 2 && convo.messages[i - 2].role === 'assistant' && message.model && convo.messages[i - 2].model && convo.messages[i - 2].model.id !== message.model.id) || (message.role === 'assistant' && (i === 1 || i === 2) && message.model && convo.model.id !== message.model.id)}
															<p class="text-[10px]">{formatModelName(message.model)}</p>
														{/if}
													</div>

													<div
														class="absolute bottom-[-32px] right-1 flex gap-x-2 opacity-0 transition-opacity group-hover:opacity-100 md:gap-x-0.5"
													>
														<button
															class="flex h-7 w-7 shrink-0 rounded-full hover:bg-gray-100"
															on:click={async () => {
																convo.messages[i].editing = true;
																convo.messages[i].pendingContent = convo.messages[i].content;
																await tick();
																textareaEls[i].style.height = 'auto';
																textareaEls[i].style.height = textareaEls[i].scrollHeight + 'px';
																textareaEls[i].focus();
																saveMessage(convo.messages[i]);
															}}
														>
															<Icon
																icon={feEdit2}
																strokeWidth={3}
																class="m-auto h-[11px] w-[11px] text-slate-600"
															/>
														</button>
														{#if message.role !== 'system'}
															<button
																class="flex h-7 w-7 shrink-0 rounded-full hover:bg-gray-100"
																on:click={() => {
																	activeToolcall = null;

																	if (message.role === 'user') {
																		if (!message.vid) {
																			message.vid = uuidv4();
																			saveMessage(message);
																		}
																		saveVersion(message, i);

																		// If user message, remove all messages after this one, then regenerate:
																		convo.messages = convo.messages.slice(0, i + 1);
																		submitCompletion();
																	} else {
																		// History is split on the user message, so get the message before this (which will be the user's):
																		const previousUserMessage = convo.messages[i - 1];
																		if (!previousUserMessage.vid) {
																			previousUserMessage.vid = uuidv4();
																			saveMessage(previousUserMessage);
																		}
																		saveVersion(previousUserMessage, i - 1);

																		// If assistant message, remove all messages after this one, including this one, then regenerate:
																		convo.messages = convo.messages.slice(0, i);
																		submitCompletion();
																	}
																	saveConversation(convo);
																}}
															>
																<Icon
																	icon={feRefreshCw}
																	strokeWidth={3}
																	class="m-auto h-[12px] w-[12px] text-slate-600"
																/>
															</button>
														{/if}
														<button
															class="flex h-7 w-7 shrink-0 rounded-full hover:bg-gray-100"
															on:click={() => {
																// Remove this message from the conversation:
																convo.messages = convo.messages
																	.slice(0, i)
																	.concat(convo.messages.slice(i + 1));
																// FIXME: Delete message from db
																saveConversation(convo);
															}}
														>
															<Icon
																icon={feX}
																strokeWidth={3}
																class="m-auto h-[14px] w-[14px] text-slate-600"
															/>
														</button>
													</div>
												{/if}
											</div>
											<button
												on:click={async () => {
													// Insert a blank message inbetween the next message and the next next message:
													let role;
													if (message.role === 'assistant' || message.role === 'system') {
														role = 'user';
													} else {
														role = 'assistant';
													}
													const msg = {
														id: uuidv4(),
														role,
														content: '',
														editing: true,
													};
													convo.messages.splice(i + 1, 0, msg);
													convo.messages = convo.messages;
													await tick();
													textareaEls[i + 1].focus();

													saveMessage(msg);
													saveConversation(convo);
												}}
												class="z-1 absolute bottom-0 left-1/2 flex h-6 w-6 -translate-x-1/2 translate-y-1/2 items-center justify-center rounded-md border border-slate-200 bg-white opacity-0 transition-opacity hover:bg-gray-200 group-hover:opacity-100"
											>
												<Icon icon={fePlus} class="m-auto h-3 w-3 text-slate-600" />
											</button>
										</li>
									{/if}
								{/each}
							</ul>
						{:else}
							<Button
								variant="outline"
								class="z-[98] mx-auto rounded-t-none !border-t-0 border-dashed text-xs"
								on:click={insertSystemPrompt}
							>
								<Icon icon={feMessageCircle} class="mr-2 h-3 w-3 text-slate-600" />
								Add system prompt
							</Button>
						{/if}
					</div>

					<div
						class="input-floating absolute bottom-4 left-1/2 z-[99] w-full -translate-x-1/2 px-5 ld:px-8"
					>
						<div class="mx-auto flex w-full max-w-[680px] flex-col ld:max-w-[768px]">
							<div class="relative flex">
								{#if imageUrls.length > 0 || pendingFiles.length > 0}
									<div class="absolute left-[50px] top-2.5 flex gap-x-3">
										{#each pendingFiles as file, i}
											<div class="relative">
												<FilePreview
													filename={file.name}
													class="my-auto !gap-1 whitespace-pre-wrap px-4 text-center [overflow-wrap:anywhere]"
													outerClass="!gap-1 h-20 w-20"
													filenameClass="!text-[10px] !leading-relaxed line-clamp-2"
													badgeClass="absolute bottom-0 left-1/2 -translate-x-1/2 translate-y-1/2"
												/>
												<button
													on:click={() => {
														pendingFiles.splice(i, 1);
														pendingFiles = pendingFiles;
														tick().then(() => {
															autoresizeTextarea();
														});
													}}
													class="absolute -bottom-1 -right-1 flex h-4 w-4 rounded-full bg-black transition-transform hover:scale-110"
												>
													<Icon icon={feX} class="m-auto h-3 w-3 text-white" />
												</button>
											</div>
										{/each}
										{#each imageUrls as url, i}
											<div class="relative">
												<img
													src={url}
													alt=""
													class="h-20 w-20 rounded-lg border border-slate-200 object-cover"
												/>
												<button
													on:click={() => {
														imageUrls.splice(i, 1);
														imageUrls = imageUrls;
														imageUrlsBlacklist.push(url);
														tick().then(() => {
															autoresizeTextarea();
														});
													}}
													class="absolute -bottom-1 -right-1 flex h-4 w-4 rounded-full bg-black transition-transform hover:scale-110"
												>
													<Icon icon={feX} class="m-auto h-3 w-3 text-white" />
												</button>
											</div>
										{/each}
									</div>
								{/if}
								{#if isMultimodal}
									<button
										class="absolute bottom-[13px] left-4 h-8 w-8 rounded-full bg-slate-800 transition-transform hover:scale-110"
										on:click={() => fileInputEl.click()}
									>
										<input
											type="file"
											class="hidden"
											bind:this={fileInputEl}
											on:change={async (event) => {
												const files = event.target.files;
												for (let i = 0; i < files.length; i++) {
													const file = files[i];
													if (file.type.startsWith('image/')) {
														const dataUrl = await readFileAsDataURL(file);
														imageUrls.push(dataUrl);
														imageUrls = imageUrls;
														tick().then(() => {
															autoresizeTextarea();
														});
													} else {
														const text = await file.text();
														pendingFiles.push({ name: file.name, text });
														pendingFiles = pendingFiles;
														tick().then(() => {
															autoresizeTextarea();
														});
													}
												}
											}}
										/>
										<Icon
											icon={fePaperclip}
											class="m-auto h-3.5 w-3.5 text-white transition-colors group-disabled:text-slate-400"
										/>
									</button>
								{/if}
								<textarea
									bind:this={inputTextareaEl}
									class="{isMultimodal ? '!pl-[58px]' : ''} {imageUrls.length > 0 ||
									pendingFiles.length > 0
										? '!pt-[112px]'
										: ''} max-h-[90dvh] w-full resize-none rounded-2xl border border-slate-200 py-4 pl-5 pr-14 font-normal text-slate-800 shadow-sm transition-colors scrollbar-slim focus:border-slate-400 focus:outline-none"
									rows={1}
									bind:value={content}
									on:paste={async (event) => {
										const items = (event.clipboardData || event.originalEvent.clipboardData).items;
										for (let i = 0; i < items.length; i++) {
											if (items[i].kind === 'file' && items[i].type.startsWith('image/')) {
												const file = items[i].getAsFile();
												const dataUrl = await readFileAsDataURL(file);
												imageUrls.push(dataUrl);
												imageUrls = imageUrls;
												tick().then(() => {
													autoresizeTextarea();
												});
											}
										}
									}}
									on:keydown={(event) => {
										if (event.key === 'Enter' && !event.shiftKey && innerWidth > 880) {
											event.preventDefault();
											sendMessage();
										}
									}}
									on:input={async () => {
										autoresizeTextarea();

										const imageLinkedUrls = content.match(imageUrlRegex) || [];
										for (const url of imageLinkedUrls) {
											if (!imageUrls.includes(url) && !imageUrlsBlacklist.includes(url)) {
												imageUrls.push(url);
												imageUrls = imageUrls;
												tick().then(() => {
													autoresizeTextarea();
												});
											}
										}
									}}
								/>
								{#if content.length > 0}
									<button
										transition:fly={{ x: 2, duration: 300 }}
										disabled={content.length === 0}
										class="group absolute bottom-[13px] right-4 flex h-8 w-8 rounded-full bg-slate-800 transition-transform hover:scale-110"
										on:click={sendMessage}
									>
										<Icon
											icon={feArrowUp}
											class="m-auto h-4 w-4 text-white transition-colors group-disabled:text-slate-100"
										/>
									</button>
								{:else if generating && convo.messages.filter((msg) => msg.generated).length > 0}
									<button
										transition:fly={{ x: 2, duration: 300 }}
										class="group absolute bottom-[13px] right-4 flex h-8 w-8 rounded-full bg-slate-800 transition-transform hover:scale-110"
										on:click={() => {
											$controller.abort();
											generating = false;
										}}
									>
										<Icon
											icon={feSquare}
											strokeWidth={4}
											class="m-auto h-3.5 w-3.5 text-white transition-colors group-disabled:text-slate-100"
										/>
									</button>
								{/if}
							</div>
						</div>
					</div>
				</div>

				{#if splitView}
					{@const toolresponse = convo.messages.find((msg) => msg.toolcallId === activeToolcall.id)}
					<div in:fade={{ duration: 500 }} class="w-[50%] p-3.5">
						<Toolcall
							toolcall={activeToolcall}
							{toolresponse}
							collapsable={false}
							closeButton
							class="!rounded-xl"
							on:close={() => {
								activeToolcall = null;
							}}
						/>
					</div>
				{/if}
			</div>
		</div>

		<KnobsSidebar {knobsOpen} />
	</div>
</main>

<SettingsModal
	open={settingsModalOpen}
	trigger="settings"
	on:fetchModels={fetchModels}
	on:disableTool={({ detail: name }) => {
		convo.tools = convo.tools.filter((n) => n !== name);
		saveConversation(convo);
	}}
/>

{#if innerWidth <= 1215 && !$config.explicitToolView}
	<Modal
		bind:open={toolcallModalOpen}
		trigger="toolcall"
		class="!p-0"
		buttonClass="hidden"
		on:close={() => {
			activeToolcall = null;
		}}
	>
		<Toolcall
			toolcall={activeToolcall}
			toolresponse={convo.messages.find((msg) => msg.toolcallId === activeToolcall.id)}
			collapsable={false}
			closeButton
			class="!rounded-xl"
			on:close={() => {
				toolcallModalOpen = false;
				activeToolcall = null;
			}}
		/>
	</Modal>
{/if}

<style lang="postcss">
	:global(.standalone .input-floating) {
		bottom: 32px;
	}
	:global(.standalone .scrollable) {
		@apply pb-[100px];
	}
	:global(.standalone .settings-trigger-container) {
		@apply pb-10;
	}

	:global(.markdown.prose :where(p):not(:where([class~='not-prose'], [class~='not-prose'] *))) {
		@apply my-3;
	}

	:global(
			.markdown.prose
				:where(.prose > :first-child):not(:where([class~='not-prose'], [class~='not-prose'] *))
		) {
		@apply mt-0;
	}

	:global(
			.markdown.prose
				:where(.prose > :last-child):not(:where([class~='not-prose'], [class~='not-prose'] *))
		) {
		@apply mb-0;
	}
</style>
