import type { ReactNode } from "react";
import { InfoTooltip } from "../ui/info-tooltip";

interface DashboardSectionHeaderProps {
  title: string;
  tooltipContent: ReactNode;
}

export function DashboardSectionHeader({
  title,
  tooltipContent,
}: DashboardSectionHeaderProps) {
  return (
    <div className="mb-4">
      <div className="relative inline-block">
        <h2 className="text-2xl font-semibold relative inline-block mr-2">{title}</h2>
        <InfoTooltip content={tooltipContent} />
      </div>
    </div>
  );
}
