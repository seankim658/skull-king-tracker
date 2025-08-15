import { PageHeader } from "@/components/ui/page-header";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Trophy } from "lucide-react";
import { AWARD_DEFINITIONS } from "@/lib/award-definitions";

export function AwardsDefinitionsPage() {
  return (
    <div className="container mx-auto p-4 md:p-6 space-y-8">
      <PageHeader
        title="Award Definitions"
        description="Learn about the various awards you can earn at the end of a game. Awards are only assigned for games with four or more players."
      />
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {AWARD_DEFINITIONS.map((award) => (
          <Card key={award.title} className="flex flex-col">
            <CardHeader>
              <div className="flex items-start gap-4">
                <Trophy className="h-8 w-8 text-amber-500 flex-shrink-0 mt-1" />
                <div>
                  <CardTitle>{award.title}</CardTitle>
                  <CardDescription className="mt-1">
                    {award.description}
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            {award.calculation && (
              <CardContent className="pt-0">
                <div className="border-t pt-3">
                  <p className="text-xs text-muted-foreground italic">
                    <span className="font-semibold">How it's calculated:</span>{" "}
                    {award.calculation}
                  </p>
                </div>
              </CardContent>
            )}
          </Card>
        ))}
      </div>
    </div>
  );
}
