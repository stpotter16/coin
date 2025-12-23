import { AppOverviewCards } from "../components/app-overview-cards";
import { DashboardTable } from "../components/dashboard-table";

export function Dashboard() {
  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <AppOverviewCards />
      <div className="space-y-4">
        <h3 className="text-lg font-medium">Recent Transactions</h3>
        <DashboardTable />
      </div>
    </div>
  );
}
