/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./views/*.templ",
    "./views/*/*.templ",
    "./views/**/**/*.templ",
    "./static/src/js/*.js",
  ],
  plugins: [],
};
