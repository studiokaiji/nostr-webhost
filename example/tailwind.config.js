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
      backgroundImage: {
        "rainbow-gradient":
          "linear-gradient(to right,#e60000,#f39800,#fff100,#009944,#0068b7,#1d2088,#920783,#e60000)",
      },
    },
  },
  plugins: [],
};
