/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    fontFamily: {
      body: ["Josefin Sans", "ui-sans-serif", "system-ui"],
    },
    extend: {
      colors: {
        primary: "#8e30eb",
      },
    },
  },
  plugins: [],
};
