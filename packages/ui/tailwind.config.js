/** @type {import('tailwindcss').Config} */
const config = require('tailwind-config/tailwind.config.js');

// module.exports = {
//   // prefix ui lib classes to avoid conflicting with the app
//   // prefix: 'ui-',
//   presets: [config],
// };

module.exports = {
  ...config,
  content: ['./**/*.{js,ts,jsx,tsx,mdx}'],
};
