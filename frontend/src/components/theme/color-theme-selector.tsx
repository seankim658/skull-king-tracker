import { useTheme } from "@/hooks/use-theme";
import { Label } from "../ui/label";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { COLOR_THEMES } from "@/lib/themes";

export function ColorThemeSelector() {
  const { colorTheme, setColorTheme } = useTheme();

  return (
    <div className="flex items-center gap-2">
      <Label htmlFor="theme-selector" className="sr-only">
        Theme
      </Label>
      <Select value={colorTheme} onValueChange={setColorTheme}>
        <SelectTrigger
          id="theme-selector"
          className="w-full sm:w-auto min-w-[180px]"
        >
          <SelectValue placeholder="Select color theme" />
        </SelectTrigger>
        <SelectContent align="end">
          <SelectGroup>
            {COLOR_THEMES.map((theme) => (
              <SelectItem key={theme.name} value={theme.value}>
                {theme.name}
              </SelectItem>
            ))}
          </SelectGroup>
        </SelectContent>
      </Select>
    </div>
  );
}
