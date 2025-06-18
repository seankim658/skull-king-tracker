import type { ReactNode } from "react";
import { useEffect, useState, useCallback, useMemo } from "react";
import { useAuth } from "@/hooks/use-auth";
import type { UITheme, ColorTheme } from "@/lib/themes";
import {
  DEFAULT_UI_THEME,
  DEFAULT_COLOR_THEME,
  UI_THEME_STORAGE_KEY,
  COLOR_SCHEME_STORAGE_KEY,
  COLOR_THEMES,
} from "@/lib/themes";
import { ThemeContext } from "@/contexts/theme-context";
import type { ThemeContextType } from "@/contexts/theme-context";
import { toast } from "sonner";
import { userAPI } from "@/lib/api/service/user";
import { errorExtract } from "@/lib/utils";
import { useSubmit } from "@/hooks/use-submit";

interface ThemeProviderProps {
  children: ReactNode;
}

export function ThemeProvider({ children }: ThemeProviderProps) {
  const { isAuthenticated, checkAuthStatus } = useAuth();

  const [uiTheme, setUIThemeState] = useState<UITheme>(() => {
    return (
      (localStorage.getItem(UI_THEME_STORAGE_KEY) as UITheme) ||
      DEFAULT_UI_THEME
    );
  });

  const [colorTheme, setColorThemeState] = useState<ColorTheme>(() => {
    return (
      (localStorage.getItem(COLOR_SCHEME_STORAGE_KEY) as ColorTheme) ||
      DEFAULT_COLOR_THEME
    );
  });

  const { submit: saveThemeToServer } = useSubmit(userAPI.updateThemeSettings, {
    actionVerb: "Saving theme preference",
    successMessage: "Theme updated successfully",
    onSuccess: () => checkAuthStatus(),
  });

  const setUITheme = useCallback(
    (theme: UITheme) => {
      setUIThemeState(theme);
      localStorage.setItem(UI_THEME_STORAGE_KEY, theme);
      if (isAuthenticated) {
        saveThemeToServer({ ui_theme: theme, color_theme: colorTheme });
      }
    },
    [isAuthenticated, colorTheme, saveThemeToServer],
  );

  const setColorTheme = useCallback(
    (theme: ColorTheme) => {
      setColorThemeState(theme);
      localStorage.setItem(COLOR_SCHEME_STORAGE_KEY, theme);
      if (isAuthenticated) {
        saveThemeToServer({ ui_theme: uiTheme, color_theme: theme });
      }
    },
    [isAuthenticated, uiTheme, saveThemeToServer],
  );

  useEffect(() => {
    const root = window.document.documentElement;
    root.classList.remove("light", "dark");
    const effectiveTheme: "light" | "dark" =
      uiTheme === "system"
        ? window.matchMedia("(prefers-color-scheme: dark)").matches
          ? "dark"
          : "light"
        : uiTheme;
    root.classList.add(effectiveTheme);

    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const handleChange = () => {
      if (uiTheme === "system") {
        const newEffectiveTheme = mediaQuery.matches ? "dark" : "light";
        root.classList.remove("light", "dark");
        root.classList.add(newEffectiveTheme);
      }
    };
    mediaQuery.addEventListener("change", handleChange);
    return () => mediaQuery.removeEventListener("change", handleChange);
  }, [uiTheme]);

  useEffect(() => {
    const body = document.body;
    const currentThemeBase = colorTheme.replace("-scaled", "");
    const isScaled = colorTheme.endsWith("-scaled");

    Array.from(body.classList)
      .filter((className) => className.startsWith("theme-"))
      .forEach((className) => {
        body.classList.remove(className);
      });
    body.classList.remove("theme-scaled");

    body.classList.add(`theme-${currentThemeBase}`);
    if (isScaled) {
      body.classList.add("theme-scaled");
    }
  }, [colorTheme]);

  const contextValue = useMemo<ThemeContextType>(
    () => ({
      uiTheme,
      colorTheme,
      setUITheme,
      setColorTheme,
      availableColorThemes: COLOR_THEMES.map((ct) => ({
        name: ct.name,
        value: ct.value,
      })),
    }),
    [uiTheme, colorTheme, setUITheme, setColorTheme],
  );

  return (
    <ThemeContext.Provider value={contextValue}>
      {children}
    </ThemeContext.Provider>
  );
}
