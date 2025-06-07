import { SiteSummaryStats } from "@/components/explore/site-summary-stats";
import { UserSearch } from "@/components/explore/user-search";
import { Separator } from "@/components/ui/separator";

export function ExplorePage() {
  return (
    <div className="container mx-auto p-4 md:p-6 space-y-10">
      <div>
        <h1 className="text-3xl font-bold tracking-tight mb-2">Explore</h1>
      </div>

      <SiteSummaryStats />

      <Separator />

      <div>
        <h2 className="text-2xl font-semibold mb-4 text-center">
          Find Players
        </h2>
        <UserSearch />
      </div>

      <Separator />

      {/* TODO : Placeholder for future user table/leaderboard */}
      <div className="text-center py-10">
        <h2 className="text-xl font-semibold mb-3">Global Leaderboard</h2>
        <p className="text-muted-foreground">(Coming Soon)</p>
      </div>
    </div>
  );
}
