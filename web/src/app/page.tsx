import { LayoutDashboard } from "lucide-react";
import { KpiCards } from "./(dashboard)/kpi-cards";
import { ServiceHealthGrid } from "./(dashboard)/service-health-grid";
import { RecentActivity } from "./(dashboard)/recent-activity";

export default function DashboardPage() {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center gap-3">
        <LayoutDashboard className="size-5 text-primary" />
        <h2 className="text-lg font-semibold text-foreground">Dashboard</h2>
      </div>

      <KpiCards />

      <div className="flex flex-col gap-2">
        <h3 className="text-sm font-medium text-muted-foreground">
          Service Health
        </h3>
        <ServiceHealthGrid />
      </div>

      <div className="flex flex-col gap-2">
        <RecentActivity />
      </div>
    </div>
  );
}
