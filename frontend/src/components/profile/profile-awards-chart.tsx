import { Bar, BarChart, CartesianGrid, Cell, LabelList, XAxis } from "recharts";
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

const categoryColors = [
  "var(--category-color-blue)",
  "var(--category-color-rose)",
  "var(--category-color-violet)",
  "var(--category-color-amber)",
  "var(--category-color-red)",
  "var(--category-color-orange)",
  "var(--category-color-fuchsia)",
  "var(--category-color-yellow)",
  "var(--category-color-green)",
  "var(--category-color-sky)",
  "var(--category-color-indigo)",
  "var(--category-color-lime)",
  "var(--category-color-purple)",
  "var(--category-color-pink)",
];

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
    return (
      <Card>
        <CardHeader>
          <CardTitle>Trophy Case </CardTitle>
          <CardDescription>A summary of all awards won.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-10">
            <p className="text-muted-foreground">
              No awards yet. Keep playing to earn some!
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  const awardDataWithColors = awardsData?.map((award, index) => ({
    ...award,
    fill: categoryColors[index % categoryColors.length],
  }));

  return (
    <Card>
      <CardHeader>
        <CardTitle>Trophy Case</CardTitle>
        <CardDescription>A summary of all awards won.</CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-80 w-full">
          <BarChart
            accessibilityLayer
            data={awardDataWithColors}
            margin={{ top: 20, bottom: 75 }}
          >
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="award_title"
              tickLine={false}
              axisLine={false}
              tickMargin={10}
              angle={-45}
              textAnchor="end"
            />
            <ChartTooltip
              cursor={false}
              content={
                <ChartTooltipContent
                  formatter={(value, _, item) => {
                    const { percentile } = item.payload;
                    return (
                      <div className="min-w-[150px] space-y-1.5 p-1">
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">Count:</span>
                          <span className="font-medium">{value}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">Rank:</span>
                          <span className="font-medium">
                            Top {percentile}%
                          </span>
                        </div>
                      </div>
                    );
                  }}
                />
              }
            />
            <Bar dataKey="count" radius={4}>
              {awardDataWithColors.map((entry) => (
                <Cell key={entry.award_type} fill={entry.fill} />
              ))}
              <LabelList
                position="top"
                offset={10}
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
