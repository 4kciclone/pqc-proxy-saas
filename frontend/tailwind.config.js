/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        dark: "#0F172A",     // Fundo Profundo (Slate 900)
        card: "#1E293B",     // Cartões (Slate 800)
        primary: "#10B981",  // Verde Emerald (Sucesso/Online)
        accent: "#38BDF8",   // Azul Sky (Acentos)
        danger: "#EF4444",   // Vermelho (Erro/Ataque)
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'],
      }
    },
  },
  plugins: [],
}