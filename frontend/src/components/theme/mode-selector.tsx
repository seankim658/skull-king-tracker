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

export function ModeSelector() {
  const { uiTheme, setUITheme } = useTheme();

  return (
    <div className="flex items-center gap-2">
      <Label htmlFor="theme-selector" className="sr-only">
        Theme
      </Label>
      <Select value={uiTheme} onValueChange={setUITheme}>
        <SelectTrigger
          id="theme-selector"
          className="w-full sm:w-auto min-w-[180px]"
        >
          <SelectValue placeholder="Select color theme" />
        </SelectTrigger>
        <SelectContent align="end">
          <SelectGroup>
            <SelectItem key="system" value="system">
              System
            </SelectItem>
            <SelectItem key="light" value="light">
              Light
            </SelectItem>
            <SelectItem key="dark" value="dark">
              Dark
            </SelectItem>
          </SelectGroup>
        </SelectContent>
      </Select>
    </div>
  );
}
