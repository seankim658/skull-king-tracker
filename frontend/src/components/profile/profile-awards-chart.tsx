import { Bar, BarChart, CartesianGrid, LabelList, XAxis } from "recharts";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../ui/card";
import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "../ui/chart";
import { useQuery } from "@tanstack/react-query";
import { statsAPI } from "@/lib/api/service/stat";
import { Skeleton } from "../ui/skeleton";

interface ProfileAwardsChartProps {
  userId: string;
}

const chartConfig = {
  count: {
    label: "Count",
    color: "hsl(var(--chart-1))",
  },
} satisfies ChartConfig;

export function ProfileAwardsChart({ userId }: ProfileAwardsChartProps) {
  const {
    data: awardsData,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ["userAwardsStats", userId],
    queryFn: async () => {
      const response = await statsAPI.getUserAwardsStats(userId);
      if (!response.success || !response.data) {
        throw new Error(response.message || "Failed to fetch awards stats");
      }
      return response.data;
    },
    enabled: !!userId,
  });

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-7 w-32" />
          <Skeleton className="h-4 w-48" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-64 w-full" />
        </CardContent>
      </Card>
    );
  }

  if (isError || !awardsData || awardsData.length === 0) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Trophy Case</CardTitle>
        <CardDescription>A summary of all awards won.</CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-64 w-full">
          <BarChart
            accessibilityLayer
            data={awardsData}
            layout="vertical"
            margin={{ left: 10, right: 10 }}
          >
            <CartesianGrid horizontal={false} />
            <XAxis type="number" dataKey="count" hide />
            <ChartTooltip
              cursor={false}
              content={<ChartTooltipContent hideLabel />}
            />
            <Bar dataKey="count" fill="var(--color-count)" radius={4}>
              <LabelList
                dataKey="award_title"
                position="insideLeft"
                offset={8}
                className="fill-background"
                fontSize={12}
              />
              <LabelList
                dataKey="count"
                position="right"
                offset={8}
                className="fill-foreground"
                fontSize={12}
              />
            </Bar>
          </BarChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
