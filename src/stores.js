import { persisted } from './localstorage.js';

export const params = persisted('params', {
	temperature: 0.3,
});

export const openrouterAPIKey = persisted('openrouterkey', '');
export const groqAPIKey = persisted('groqAPIKey', '');

export const remoteServer = persisted('remoteServer', { address: '', password: '' });
export const toolSchema = persisted('toolSchema', []);
export const tools = persisted('tools', []);
