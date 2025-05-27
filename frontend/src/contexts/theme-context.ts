import { createContext } from "react";
import type { UITheme, ColorTheme } from "@/lib/themes";

export interface ThemeContextType {
  uiTheme: UITheme;
  colorTheme: ColorTheme;
  setUITheme: (theme: UITheme) => void;
  setColorTheme: (theme: ColorTheme) => void;
  availableColorThemes: { name: string; value: string }[];
}

export const ThemeContext = createContext<ThemeContextType | undefined>(
  undefined,
);
