import { AppOverviewCards } from "@/components/app-overview-cards";
import { AppTransactionsTable } from "@/components/app-transactions-table";

export default function Home() {
  return (
    <div className="p-6 space-y-6">
      <AppOverviewCards />
      <AppTransactionsTable />
    </div>
  );
}
