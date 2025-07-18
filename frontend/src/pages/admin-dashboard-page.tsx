import { useState } from "react";
import { PageHeader } from "@/components/ui/page-header";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ReportsTable } from "@/components/admin/reports-table";
import { AlertsTable } from "@/components/admin/alerts-table";
import { Button } from "@/components/ui/button";
import { Send } from "lucide-react";
import { SendNotificationModal } from "@/components/admin/send-notification-modal";

export function AdminDashboardPage() {
  const [isNotificationModalOpen, setNotificationModalOpen] = useState(false);

  return (
    <div className="container mx-auto p-4 md:p-6 space-y-8">
      <div className="flex justify-between items-center">
        <PageHeader
          title="Admin Panel"
          description="Review reports and manage site-wide settings."
        />
        <Button
          onClick={() => setNotificationModalOpen(true)}
          className="cursor-pointer"
        >
          <Send className="mr-2 h-4 w-4" />
          Send Notification
        </Button>
      </div>

      <Tabs defaultValue="reports">
        <TabsList>
          <TabsTrigger value="reports">User Reports</TabsTrigger>
          <TabsTrigger value="alerts">Site Alerts</TabsTrigger>
        </TabsList>
        <TabsContent value="reports" className="mt-4">
          <ReportsTable />
        </TabsContent>
        <TabsContent value="alerts" className="mt-4">
          <AlertsTable />
        </TabsContent>
      </Tabs>

      <SendNotificationModal
        isOpen={isNotificationModalOpen}
        onClose={() => setNotificationModalOpen(false)}
      />
    </div>
  );
}
