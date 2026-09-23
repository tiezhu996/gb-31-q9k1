import type { Config } from 'tailwindcss';

const config: Config = {
  content: ['./src/**/*.{js,ts,jsx,tsx,mdx}'],
  theme: {
    extend: {
      colors: {
        brand: {
          50: '#fdf4f0',
          100: '#fbe6dc',
          500: '#f08c4a',
          600: '#e0702f',
          700: '#c25a22',
        },
      },
    },
  },
  plugins: [],
};
export default config;
