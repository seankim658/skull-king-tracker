import type { ReactNode } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "./card";
import { InfoTooltip } from "./info-tooltip";

interface StatCardProps {
  title: string;
  value: ReactNode;
  icon?: ReactNode;
  description?: string;
  tooltip?: string;
}

export function StatCard({
  title,
  value,
  icon,
  description,
  tooltip,
}: StatCardProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        {icon}
      </CardHeader>
      <CardContent>
        <div className="relative inline-block">
          <div className="text-2xl font-bold pr-2">{value}</div>
          {tooltip && (
            <InfoTooltip
              content={tooltip}
              wrapperClassName="absolute top-0.5 right-0"
            />
          )}
        </div>
        {description && (
          <p className="text-xs text-muted-foreground">{description}</p>
        )}
      </CardContent>
    </Card>
  );
}
