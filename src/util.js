export function readFileAsDataURL(file) {
	return new Promise((resolve, reject) => {
		const reader = new FileReader();
		reader.onload = () => resolve(reader.result);
		reader.onerror = () => reject(reader.error);
		reader.readAsDataURL(file);
	});
}

export function debounce(func, wait) {
	const timers = new Map();

	return function (...args) {
		const id = args[0].id; // Assuming the first argument has an `id` property

		if (timers.has(id)) {
			clearTimeout(timers.get(id));
		}

		const timer = setTimeout(() => {
			func.apply(this, args);
			timers.delete(id);
		}, wait);

		timers.set(id, timer);
	};
}
