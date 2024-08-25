/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./tmpl/**/*.{go,templ}", "./assets/images/logo.svg"],
  theme: {
    extend: {
      fontFamily: {
        'serif': ['Roboto Slab', 'ui-serif']
      }
    },
  },
  plugins: [],
}

