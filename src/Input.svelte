<script>
	import { fly } from 'svelte/transition';
	import FilePreview from './FilePreview.svelte';
	import Icon from './Icon.svelte';
	import { feArrowUp, fePaperclip, feSquare, feX } from './feather.js';
	import { tick } from 'svelte';
	import { v4 as uuidv4 } from 'uuid';
	import { openAIAdditionalModelsMultimodal } from './providers.js';
	import { readFileAsDataURL } from './util.js';
	import { controller, params } from './stores.js';

	const imageUrlRegex = /https?:\/\/[^\s]+?\.(png|jpe?g)(?=\s|$)/gi;

	export let generating;
	export let convo;
	export let saveMessage;
	export let saveConversation;
	export let submitCompletion;
	export let scrollToBottom;

	let content = '';
	let pendingImages = [];
	let imageUrlsBlacklist = [];
	let pendingFiles = [];

	$: isMultimodal =
		convo.models[0].modality === 'text+image->text' ||
		openAIAdditionalModelsMultimodal.includes(convo.models[0].id);

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

			const imageUrlMapper = (image) => ({
				type: 'image_url',
				image_url: {
					url: image.url,
					detail: image.fidelity,
				},
			});

			if (pendingImages.length > 0) {
				msg.contentParts = [...pendingImages.map(imageUrlMapper)];
			}

			let fileContent = '';
			if (pendingFiles.length > 0) {
				for (const file of pendingFiles) {
					fileContent += `\`\`\`filename="${file.name}"
${file.text}
\`\`\`

`;
				}

				msg.content = fileContent + msg.content;
			}

			convo.messages.push(msg);
			convo.messages = convo.messages;
			await tick();
			scrollToBottom();

			saveMessage(msg);
			saveConversation(convo);

			content = '';
			pendingImages = [];
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

	let fileInputEl;
	export let inputTextareaEl;

	export function autoresizeTextarea() {
		inputTextareaEl.style.height = 'auto';
		inputTextareaEl.style.height = inputTextareaEl.scrollHeight + 2 + 'px';
	}

	async function handlePDF(file) {
		try {
			const pdfjs = await import('pdfjs-dist');

			pdfjs.GlobalWorkerOptions.workerSrc = new URL(
				'pdfjs-dist/build/pdf.worker.min.mjs',
				import.meta.url
			).toString();

			const arrayBuffer = await file.arrayBuffer();
			const pdf = await pdfjs.getDocument({ data: arrayBuffer }).promise;
			let text = '';

			for (let i = 1; i <= pdf.numPages; i++) {
				const page = await pdf.getPage(i);
				const content = await page.getTextContent();
				const pageText = content.items.map((item) => item.str).join(' ');
				text += pageText;
			}

			pendingFiles.push({ name: file.name, text });
			pendingFiles = pendingFiles;
			tick().then(() => {
				autoresizeTextarea();
			});
		} catch (error) {
			console.error(`Error processing PDF: ${error.message}`);
		}
	}

	export async function handleFileDrop(event) {
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

				if (item.type === 'application/pdf') {
					handlePDF(item.getAsFile());
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
	}

	async function handlePaste(event) {
		const items = (event.clipboardData || event.originalEvent.clipboardData).items;
		for (let i = 0; i < items.length; i++) {
			if (items[i].kind === 'file' && items[i].type.startsWith('image/')) {
				const file = items[i].getAsFile();
				const dataUrl = await readFileAsDataURL(file);
				pendingImages.push({ url: dataUrl, fidelity: 'high' });
				pendingImages = pendingImages;
				tick().then(() => {
					autoresizeTextarea();
				});
			} else if (items[i].kind === 'file' && items[i].type === 'application/pdf') {
				handlePDF(items[i].getAsFile());
			} else if (items[i].kind === 'string' && items[i].type === 'text/plain') {
				event.preventDefault();
				items[i].getAsString((text) => {
					if (text.split('\n').length >= 100) {
						pendingFiles.push({ name: 'Pasted.txt', text });
						pendingFiles = pendingFiles;
						tick().then(() => {
							autoresizeTextarea();
						});
					} else {
						// If text is less than 100 lines, manually insert it into the textarea
						// FIXME: An "undo" button should appear for pasted files
						const textarea = event.target;
						const start = textarea.selectionStart;
						const end = textarea.selectionEnd;
						const value = textarea.value;
						content = value.substring(0, start) + text + value.substring(end);
						textarea.selectionStart = textarea.selectionEnd = start + text.length;
						tick().then(() => {
							autoresizeTextarea();
						});
					}
				});
			}
		}
	}

	async function handleFileUpload(event) {
		const files = event.target.files;
		for (let i = 0; i < files.length; i++) {
			const file = files[i];
			if (file.type.startsWith('image/')) {
				const dataUrl = await readFileAsDataURL(file);
				pendingImages.push({ url: dataUrl, fidelity: 'high' });
				pendingImages = pendingImages;
				tick().then(() => {
					autoresizeTextarea();
				});
			} else if (file.type === 'application/pdf') {
				handlePDF(file);
			} else {
				const text = await file.text();
				pendingFiles.push({ name: file.name, text });
				pendingFiles = pendingFiles;
				tick().then(() => {
					autoresizeTextarea();
				});
			}
		}
	}
</script>

<div class="input-floating absolute bottom-4 left-1/2 z-[99] w-full -translate-x-1/2 px-5 ld:px-8">
	<div class="mx-auto flex w-full max-w-[680px] flex-col ld:max-w-[768px]">
		<div class="relative flex">
			{#if pendingImages.length > 0 || pendingFiles.length > 0}
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
								class="absolute -bottom-1 -right-1 flex h-4 w-4 rounded-full bg-black transition-[transform,background-color] hover:scale-110 hover:bg-red-400"
							>
								<Icon icon={feX} class="m-auto h-3 w-3 text-white" />
							</button>
							{#if file.name === 'Pasted.txt'}
								<button
									on:click={() => {
										content += file.text;
										pendingFiles.splice(i, 1);
										pendingFiles = pendingFiles;
										tick().then(() => {
											autoresizeTextarea();
										});
									}}
									class="absolute -bottom-1 -left-1 flex h-4 w-4 rounded-full bg-black transition-[transform,background-color] hover:scale-110 hover:bg-blue-400"
								>
									<Icon icon={feArrowUp} class="m-auto h-3 w-3 rotate-[-90deg] text-white" />
								</button>
							{/if}
						</div>
					{/each}
					{#each pendingImages as image, i}
						<div class="relative">
							<img
								src={image.url}
								alt=""
								class="h-20 w-20 rounded-lg border border-slate-200 object-cover"
							/>
							<button
								on:click={() => {
									pendingImages[i].fidelity = pendingImages[i].fidelity === 'high' ? 'low' : 'high';
									pendingImages = pendingImages;
								}}
								class="absolute -bottom-1 -left-1 flex h-4 px-1 rounded-full bg-black transition-[transform,background-color] hover:scale-110 hover:bg-blue-400"
								title="Toggle image fidelity"
							>
								<span class="m-auto text-[8px] font-bold text-white">
									{pendingImages[i].fidelity === 'high' ? 'High' : 'Low'}
								</span>
							</button>
							<button
								on:click={() => {
									pendingImages.splice(i, 1);
									pendingImages = pendingImages;
									imageUrlsBlacklist.push(image.url);
									tick().then(() => {
										autoresizeTextarea();
									});
								}}
								class="absolute -bottom-1 -right-1 flex h-4 w-4 rounded-full bg-black transition-[transform,background-color] hover:scale-110 hover:bg-red-400"
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
					<input type="file" class="hidden" bind:this={fileInputEl} on:change={handleFileUpload} />
					<Icon
						icon={fePaperclip}
						class="m-auto h-3.5 w-3.5 text-white transition-colors group-disabled:text-slate-400"
					/>
				</button>
			{/if}
			<textarea
				bind:this={inputTextareaEl}
				class="{isMultimodal ? '!pl-[58px]' : ''} {pendingImages.length > 0 ||
				pendingFiles.length > 0
					? '!pt-[112px]'
					: ''} max-h-[90dvh] w-full resize-none rounded-[18px] border border-slate-200 py-4 pl-5 pr-14 font-normal text-slate-800 shadow-sm transition-colors scrollbar-slim focus:border-slate-300 focus:outline-none"
				rows={1}
				bind:value={content}
				on:paste={handlePaste}
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
						if (!pendingImages.find(image => image.url === url) && !imageUrlsBlacklist.includes(url)) {
							pendingImages.push({ url, fidelity: 'high' });
							pendingImages = pendingImages;
							tick().then(() => {
								autoresizeTextarea();
							});
						}
					}
				}}
			/>
			{#if generating && convo.messages.filter((msg) => msg.generated).length > 0}
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
			{:else}
				<button
					transition:fly={{ x: 2, duration: 300 }}
					disabled={content.length === 0}
					class="group absolute bottom-[13px] right-4 flex h-8 w-8 rounded-full bg-slate-800 transition-transform hover:scale-110 disabled:bg-slate-400 disabled:hover:scale-100"
					on:click={sendMessage}
				>
					<Icon
						icon={feArrowUp}
						class="m-auto h-4 w-4 text-white transition-colors group-disabled:text-slate-100"
					/>
				</button>
			{/if}
		</div>
	</div>
</div>
