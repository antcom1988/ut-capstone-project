/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#f0f7fc',
          100: '#e1eff9',
          500: '#1f5f8b',
          600: '#174c71',
          700: '#123d5c',
        }
      }
    },
  },
  plugins: [],
}
