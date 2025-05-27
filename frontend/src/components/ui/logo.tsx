import { useState, useEffect } from "react";
import LogoLight from "@/assets/logo_light.png";
import LogoDark from "@/assets/logo_dark.png";
import { cn } from "@/lib/utils";

interface AppLogoProps {
  className?: string;
  alt?: string;
  width?: number | string;
  height?: number | string;
}

export function AppLogo({
  className,
  alt = "Skull King Logo",
  width = 180,
  height,
  ...props
}: AppLogoProps) {
  const [currentSrc, setCurrentSrc] = useState<string>(LogoLight);

  useEffect(() => {
    const root = document.documentElement;

    const updateLogoSrc = () => {
      if (root.classList.contains("dark")) {
        setCurrentSrc(LogoDark);
      } else {
        setCurrentSrc(LogoLight);
      }
    };

    updateLogoSrc();

    const observer = new MutationObserver(updateLogoSrc);
    observer.observe(root, { attributes: true, attributeFilter: ["class"] });

    return () => observer.disconnect();
  }, []);

  return (
    <img
      src={currentSrc}
      alt={alt}
      className={cn("select-none", className)}
      width={width}
      height={height || "auto"}
      {...props}
    />
  );
}
