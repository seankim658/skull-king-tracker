import { useNavigate } from "react-router-dom";
import type { UserSearchItem } from "@/lib/api/types";
import { SiteSummaryStats } from "@/components/explore/site-summary-stats";
import { UserSearch } from "@/components/explore/user-search";
import { Separator } from "@/components/ui/separator";
import { PageHeader } from "@/components/ui/page-header";

export function ExplorePage() {
  const navigate = useNavigate();

  const handleUserSelect = (user: UserSearchItem) => {
    navigate(`/users/${user.user_id}`);
  };

  return (
    <div className="container mx-auto p-4 md:p-6 space-y-10">
      <PageHeader title="Explore" />

      <SiteSummaryStats />

      <Separator />

      <section className="max-w-xl mx-auto">
        <h2 className="text-2xl font-semibold mb-2 text-center">
          Find Players
        </h2>
        <UserSearch onUserSelect={handleUserSelect} />
      </section>

      <Separator />

      {/* TODO : Placeholder for future user table/leaderboard */}
      <div className="text-center py-10">
        <h2 className="text-xl font-semibold mb-3">Global Leaderboard</h2>
        <p className="text-muted-foreground">(Coming Soon)</p>
      </div>
    </div>
  );
}
