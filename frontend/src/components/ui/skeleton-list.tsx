import { Skeleton } from "./skeleton";

interface SkeletonListProps {
  count?: number;
}

export function SkeletonList({ count = 3 }: SkeletonListProps) {
  return (
    <div className="space-y-3 p-2">
      {[...Array(count)].map((_, i) => (
        <div key={i} className="flex items-center space-x-3 p-2">
          <Skeleton className="h-10 w-10 rounded-full" />
          <div className="space-y-1.5 flex-1">
            <Skeleton className="h-4 w-3/5" />
            <Skeleton className="h-3 w-2/5" />
          </div>
        </div>
      ))}
    </div>
  );
}
