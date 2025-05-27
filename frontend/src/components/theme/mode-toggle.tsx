import { Moon, Sun } from "lucide-react";
import { Button } from "../ui/button";
import { useTheme } from "@/hooks/use-theme";

export function ModeToggle() {
  const { uiTheme, setUITheme } = useTheme();

  const toggleTheme = () => {
    const nextTheme = uiTheme === "light" ? "dark" : "light";
    setUITheme(nextTheme);
  };

  return (
    <Button
      variant="outline"
      size="icon"
      className="size-8 cursor-pointer"
      onClick={toggleTheme}
    >
      <Sun className="h-[1.1rem] w-[1.1rem] rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
      <Moon className="absolute h-[1.1rem] w-[1.1rem] rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
      <span className="sr-only">Toggle theme</span>
    </Button>
  );
}
