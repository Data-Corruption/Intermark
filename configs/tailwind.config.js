/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./data/**/*.{html,js}"],
  theme: {
    extend: {
      fontFamily: {},
    },
  },
  plugins: [require("@tailwindcss/typography"),require('daisyui')],
  daisyui: {
    themes: ["light", "dark"],
  },
};