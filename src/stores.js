import { persisted } from './localstorage.js';

export const openrouterAPIKey = persisted('openrouterkey', '');
export const groqAPIKey = persisted('groqAPIKey', '');

export const toolSchema = persisted('toolSchema', []);
export const tools = persisted('tools', []);