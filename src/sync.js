import { setAPIKeysFromObject, syncServer } from './stores.js';
import { get } from 'svelte/store';

export const llumHostedAddress = 'http://localhost:8084';

export async function syncPull({
	conversationIds,
	messageIds,
	saveConversation,
	saveMessage,
	deleteConversation,
	deleteMessage,
}) {
	const clientMissingResult = await checkClientMissing(
		get(syncServer).address || llumHostedAddress,
		get(syncServer).token,
		conversationIds,
		messageIds
	);

	if (
		clientMissingResult.missingConversationIds.length > 0 ||
		clientMissingResult.missingMessageIds.length > 0
	) {
		const serverItems = await getMissingItems(
			get(syncServer).address || llumHostedAddress,
			get(syncServer).token,
			clientMissingResult.missingConversationIds,
			clientMissingResult.missingMessageIds
		);

		const messageArray = Object.values(serverItems.messages);
		const deletedMessages = messageArray.filter((m) => m.deleted);
		const newMessages = messageArray.filter((m) => !m.deleted);

		// Delete messages that were deleted on the server
		for (const message of deletedMessages) {
			deleteMessage(message, { syncToServer: false });
		}

		// First save all the messages we got from server
		for (const message of newMessages) {
			saveMessage(message, { syncToServer: false });
		}

		const conversationArray = Object.values(serverItems.conversations);
		const deletedConversations = conversationArray.filter((c) => c.deleted);
		const newConversations = conversationArray.filter((c) => !c.deleted);

		for (const conversation of deletedConversations) {
			deleteConversation(conversation, { syncToServer: false });
		}

		// Then save all conversations we got from server
		for (const conversation of newConversations) {
			saveConversation(conversation, { convert: false, syncToServer: false });
		}

		setAPIKeysFromObject(serverItems.apiKeys);

		return {
			newConversations,
			deletedConversations,
			newMessages,
			deletedMessages,
		};
	}
	return {
		newConversations: [],
		deletedConversations: [],
		newMessages: [],
		deletedMessages: [],
	};
}

export async function syncPush({ conversations, messages }) {
	// Check what server is missing from client
	const serverMissingResult = await checkServerMissing(
		get(syncServer).address || llumHostedAddress,
		get(syncServer).token,
		conversations.map((c) => c.id),
		messages.map((m) => m.id)
	);

	// Upload missing items if any
	if (
		serverMissingResult.missingConversationIds.length > 0 ||
		serverMissingResult.missingMessageIds.length > 0
	) {
		const conversationsToSend = {};
		for (const id of serverMissingResult.missingConversationIds) {
			conversationsToSend[id] = conversations.find((c) => c.id === id);
		}

		const messagesToSend = {};
		for (const id of serverMissingResult.missingMessageIds) {
			messagesToSend[id] = messages.find((m) => m.id === id);
		}

		// Send items to server
		await sendMissingItems(
			get(syncServer).address || llumHostedAddress,
			get(syncServer).token,
			conversationsToSend,
			messagesToSend
		);
	}
}

async function checkClientMissing(baseUrl, token, localConversationIds, localMessageIds, localAPIKeyIds) {
	const response = await fetch(`${baseUrl}/api/sync/check-client-missing`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			token,
			conversationIds: localConversationIds,
			messageIds: localMessageIds,
		}),
	});

	return await response.json();
}

async function getMissingItems(baseUrl, token, missingConversationIds, missingMessageIds) {
	const response = await fetch(`${baseUrl}/api/sync/get-items`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			token,
			conversationIds: missingConversationIds,
			messageIds: missingMessageIds,
		}),
	});

	return await response.json(); // Returns { conversations, messages, apiKeys }
}

async function checkServerMissing(baseUrl, token, allLocalConversationIds, allLocalMessageIds) {
	const response = await fetch(`${baseUrl}/api/sync/check-server-missing`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			token,
			conversationIds: allLocalConversationIds,
			messageIds: allLocalMessageIds,
		}),
	});

	return await response.json(); // Returns { missingConversationIds, missingMessageIds }
}

async function sendMissingItems(baseUrl, token, conversationsToSend, messagesToSend) {
	const response = await fetch(`${baseUrl}/api/sync/send-items`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			token,
			conversations: conversationsToSend,
			messages: messagesToSend,
		}),
	});

	return await response.json(); // Returns { success: true }
}

export async function sendSingleItem(baseUrl, token, item = { conversation: null, message: null, apiKeys: null }) {
	const response = await fetch(`${baseUrl}/api/sync/send-single-item`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			token,
			conversation: item.conversation || null,
			message: item.message || null,
			apiKeys: item.apiKeys || null,
		}),
	});

	return await response.json(); // Returns { success: true }
}

export async function deleteSingleItem(
	baseUrl,
	token,
	item = { conversationId: null, messageId: null }
) {
	const response = await fetch(`${baseUrl}/api/sync/delete-single-item`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			token,
			conversationId: item.conversationId || null,
			messageId: item.messageId || null,
		}),
	});

	return await response.json(); // Returns { success: true }
}
