import { get, writable } from 'svelte/store';
import { persisted } from './localstorage.js';

export const controller = writable(null);

export const params = persisted('params', {
	temperature: 0.3,
});

export const openaiAPIKey = persisted('openaiAPIKey', '');
export const openrouterAPIKey = persisted('openrouterkey', '');
export const groqAPIKey = persisted('groqAPIKey', '');

export const remoteServer = persisted('remoteServer', { address: '', password: '' });
export const toolSchema = persisted('toolSchema', []);
export const tools = persisted('tools', []);

export function pick(parentStore, getFn) {
	const { subscribe, set } = writable(getFn(get(parentStore)));

	const unsubscribe = parentStore.subscribe((value) => {
		set(getFn(value));
	});

	return {
		subscribe,
		set: (value) => {
			parentStore.update((current) => {
				const parent = { ...current };
				const updatedValue = getFn(parent);
				Object.assign(updatedValue, value);
				return parent;
			});
		},
		update: (updateFn) => {
			parentStore.update((current) => {
				const parent = { ...current };
				const updatedValue = getFn(parent);
				Object.assign(updatedValue, updateFn(updatedValue));
				return parent;
			});
		},
		unsubscribe,
	};
}
