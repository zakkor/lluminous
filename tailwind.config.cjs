const typography = require('@tailwindcss/typography');
const forms = require('@tailwindcss/forms');

/** @type {import('tailwindcss').Config}*/
const config = {
	content: ['./src/**/*.{html,js,svelte,ts}'],

	theme: {
		extend: {
			screens: {
				xl: '1410px',
			},
			transitionTimingFunction: {
        'in-out': 'linear(0, 0.005, 0.02 2.2%, 0.045, 0.081 4.9%, 0.16 7.3%, 0.465 16.2%, 0.561, 0.642,0.713 25.8%, 0.773, 0.825 32.7%, 0.868 36.5%, 0.905 40.9%, 0.935 45.7%,0.958 51.1%, 0.975 57.4%, 0.986 64.4%, 0.993 73.1%, 0.997 84.1%, 0.999)',
      },
		},
	},

	plugins: [forms, typography],
};

module.exports = config;
