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

interface ThemeProviderProps {
  children: ReactNode;
  initialUITheme?: UITheme | null;
  initialColorTheme?: ColorTheme | null;
}

export function ThemeProvider({
  children,
  initialUITheme,
  initialColorTheme,
}: ThemeProviderProps) {
  const { user, isAuthenticated, checkAuthStatus } = useAuth();

  const [uiTheme, setUIThemeState] = useState<UITheme>(() => {
    const storedUITheme = localStorage.getItem(UI_THEME_STORAGE_KEY) as UITheme;
    if (storedUITheme && ["light", "dark", "system"].includes(storedUITheme)) {
      return storedUITheme;
    }
    if (initialUITheme) {
      localStorage.setItem(UI_THEME_STORAGE_KEY, initialUITheme);
      return initialUITheme;
    }
    localStorage.setItem(UI_THEME_STORAGE_KEY, DEFAULT_UI_THEME);
    return DEFAULT_UI_THEME;
  });

  const [colorTheme, setColorThemeState] = useState<ColorTheme>(() => {
    const storedColorTheme = localStorage.getItem(
      COLOR_SCHEME_STORAGE_KEY,
    ) as ColorTheme;
    if (storedColorTheme) {
      return storedColorTheme;
    }
    if (initialColorTheme) {
      localStorage.setItem(COLOR_SCHEME_STORAGE_KEY, initialColorTheme);
      return initialColorTheme;
    }
    localStorage.setItem(COLOR_SCHEME_STORAGE_KEY, DEFAULT_COLOR_THEME);
    return DEFAULT_COLOR_THEME;
  });

  const setUITheme = useCallback(
    async (theme: UITheme) => {
      setUIThemeState(theme);
      localStorage.setItem(UI_THEME_STORAGE_KEY, theme);
      if (isAuthenticated && user) {
        try {
          const response = await userAPI.updateThemeSettings({
            ui_theme: theme,
            color_theme: colorTheme,
          });
          if (response.success && response.data?.user) {
            checkAuthStatus();
          }
          toast.success("UI theme updated");
        } catch (e) {
          const errMsg = errorExtract(e, "Failed to save UI theme preference");
          toast.error(errMsg);
          console.error(errMsg);
        }
      }
    },
    [isAuthenticated, user, colorTheme, checkAuthStatus],
  );

  const setColorTheme = useCallback(
    async (theme: ColorTheme) => {
      setColorThemeState(theme);
      localStorage.setItem(COLOR_SCHEME_STORAGE_KEY, theme);
      if (isAuthenticated && user) {
        try {
          const response = await userAPI.updateThemeSettings({
            ui_theme: uiTheme,
            color_theme: theme,
          });
          if (response.success && response.data?.user) {
            checkAuthStatus();
          }
          toast.success("Color theme updated");
        } catch (e) {
          const errMsg = errorExtract(
            e,
            "Failed to save color theme preference",
          );
          toast.error(errMsg);
          console.error(errMsg);
        }
      }
    },
    [isAuthenticated, user, uiTheme, checkAuthStatus],
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
