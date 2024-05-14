import { writable, get } from 'svelte/store';

function openDB() {
	return new Promise((resolve, reject) => {
		const request = indexedDB.open('lluminous', 1);

		request.onupgradeneeded = (event) => {
			const db = event.target.result;
			if (!db.objectStoreNames.contains('store')) {
				db.createObjectStore('store');
			}
		};

		request.onsuccess = (event) => {
			resolve(event.target.result);
		};

		request.onerror = (event) => {
			reject(event.target.error);
		};
	});
}

function getDB(db, key) {
	return new Promise((resolve, reject) => {
		const transaction = db.transaction(['store'], 'readonly');
		const objectStore = transaction.objectStore('store');
		const request = objectStore.get(key);

		request.onsuccess = (event) => {
			resolve(event.target.result);
		};

		request.onerror = (event) => {
			reject(event.target.error);
		};
	});
}

function putDB(db, key, value) {
	return new Promise((resolve, reject) => {
		const transaction = db.transaction(['store'], 'readwrite');
		const objectStore = transaction.objectStore('store');
		const request = objectStore.put(value, key);

		request.onsuccess = () => {
			resolve();
		};

		request.onerror = (event) => {
			reject(event.target.error);
		};
	});
}

export async function persisted(key, initial) {
	const db = await openDB();
	const store = writable(initial);
	const { subscribe, set } = store;

	const persisted = await getDB(db, key);
	if (persisted === undefined && initial !== undefined) {
		set(initial);
		// If initial data is passed in, use it to initialize only if there is no persisted data.
		putDB(db, key, initial);
	} else if (persisted !== undefined) {
		// Otherwise, read persisted data from IndexedDB and set value of store to that.
		set(persisted);
	}

	return {
		subscribe,
		set: (data) => {
			// On set persist data, then update store value.
			set(data);
			putDB(db, key, data);
		},
		update: (updateFn) => {
			const updatedData = updateFn(get(store));
			set(updatedData);
			putDB(db, key, updatedData);
		},
	};
}
