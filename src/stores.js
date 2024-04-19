import { persisted } from './localstorage.js';

export const openrouterAPIKey = persisted('openrouterkey', '');
export const toolSchema = persisted('toolSchema', []);
export const tools = persisted('tools', []);