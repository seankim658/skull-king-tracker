import path from "path";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig(({ command, mode }) => {
  const commonConfig = {
    plugins: [react(), tailwindcss()],
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
      },
    },
  };

  if (mode == "development") {
    return {
      ...commonConfig,
      server: {
        port: 5173,
        proxy: {
          "/api": {
            target: "http://localhost:8081",
            changeOrigin: true,
          },
        },
      },
    };
  } else {
    return {
      ...commonConfig,
    };
  }
});
