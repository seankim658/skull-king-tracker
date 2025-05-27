export type UITheme = "dark" | "light" | "system";
export type ColorTheme = string;

export const DEFAULT_UI_THEME: UITheme = "system";
export const DEFAULT_COLOR_THEME: ColorTheme = "blue";

export const UI_THEME_STORAGE_KEY = "skullking-ui-theme";
export const COLOR_SCHEME_STORAGE_KEY = "skullking-color-theme";

export const COLOR_THEMES = [
  { name: "Default", value: "default" },
  { name: "Blue", value: "blue" },
  { name: "Green", value: "green" },
  { name: "Orange", value: "orange" },
  { name: "Red", value: "red" },
  { name: "Purple", value: "purple" },
];
