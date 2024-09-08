import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        background: "var(--background)",
        foreground: "var(--foreground)",
        'custom-black': '#333', // A softer black
        'custom-white': '#f7f7f7' // A softer white
      },
      backgroundImage: theme => ({
        'pattern-dot': "url('/path-to-dot-pattern.png')", // Ensure you have this image in your public folder
        'pattern-stripe': "url('/path-to-stripe-pattern.svg')",
      }),
    },
  },
  plugins: [],
};
export default config;
