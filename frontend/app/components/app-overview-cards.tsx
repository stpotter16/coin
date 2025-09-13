import { 
  Card, 
  CardContent, 
  CardHeader, 
  CardTitle, 
  CardAction 
} from "~/components/ui/card";
import { Badge } from "~/components/ui/badge";
import { 
  FaDollarSign, 
  FaChartLine, 
  FaPiggyBank, 
  FaCreditCard 
} from "react-icons/fa";

export function AppOverviewCards() {
  return (
    <div className="p-6 border-b">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card>
          <CardHeader>
            <CardAction>
              <FaDollarSign className="h-4 w-4 text-green-600" />
            </CardAction>
            <CardTitle>Total Balance</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">$12,345.67</div>
            <Badge variant="secondary" className="mt-2">
              +2.5% from last month
            </Badge>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardAction>
              <FaChartLine className="h-4 w-4 text-blue-600" />
            </CardAction>
            <CardTitle>Investments</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">$8,230.45</div>
            <Badge variant="default" className="mt-2">
              +5.2% this week
            </Badge>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardAction>
              <FaPiggyBank className="h-4 w-4 text-purple-600" />
            </CardAction>
            <CardTitle>Savings</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">$3,456.78</div>
            <Badge variant="outline" className="mt-2">
              Goal: 75% reached
            </Badge>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardAction>
              <FaCreditCard className="h-4 w-4 text-red-600" />
            </CardAction>
            <CardTitle>Monthly Expenses</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">$2,134.89</div>
            <Badge variant="destructive" className="mt-2">
              +12% vs budget
            </Badge>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}