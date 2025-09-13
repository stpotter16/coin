"use client";

import React from "react";
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  flexRender,
  type ColumnDef,
  type SortingState,
  type ColumnFiltersState,
} from "@tanstack/react-table";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import { ChevronUpIcon, ChevronDownIcon } from "lucide-react";

export interface Transaction {
  id: string;
  name: string;
  amount: number;
  date: string;
  account: string;
  category: string;
  type: "income" | "expense";
  status: "completed" | "pending" | "failed";
}

const sampleData: Transaction[] = [
  {
    id: "1",
    name: "Grocery Store Purchase",
    amount: -125.50,
    date: "2024-01-15",
    account: "Checking",
    category: "Food & Dining",
    type: "expense",
    status: "completed",
  },
  {
    id: "2",
    name: "Salary Deposit",
    amount: 3500.00,
    date: "2024-01-15",
    account: "Checking",
    category: "Income",
    type: "income",
    status: "completed",
  },
  {
    id: "3",
    name: "Netflix Subscription",
    amount: -15.99,
    date: "2024-01-14",
    account: "Credit Card",
    category: "Entertainment",
    type: "expense",
    status: "completed",
  },
  {
    id: "4",
    name: "Gas Station",
    amount: -65.00,
    date: "2024-01-13",
    account: "Credit Card",
    category: "Transportation",
    type: "expense",
    status: "pending",
  },
  {
    id: "5",
    name: "Freelance Payment",
    amount: 850.00,
    date: "2024-01-12",
    account: "Savings",
    category: "Income",
    type: "income",
    status: "completed",
  },
  {
    id: "6",
    name: "Coffee Shop",
    amount: -12.75,
    date: "2024-01-12",
    account: "Credit Card",
    category: "Food & Dining",
    type: "expense",
    status: "completed",
  },
  {
    id: "7",
    name: "Electric Bill",
    amount: -89.43,
    date: "2024-01-11",
    account: "Checking",
    category: "Utilities",
    type: "expense",
    status: "completed",
  },
  {
    id: "8",
    name: "Online Transfer",
    amount: -500.00,
    date: "2024-01-10",
    account: "Checking",
    category: "Transfer",
    type: "expense",
    status: "completed",
  },
  {
    id: "9",
    name: "Amazon Purchase",
    amount: -34.99,
    date: "2024-01-10",
    account: "Credit Card",
    category: "Shopping",
    type: "expense",
    status: "completed",
  },
  {
    id: "10",
    name: "Gym Membership",
    amount: -45.00,
    date: "2024-01-09",
    account: "Checking",
    category: "Health & Fitness",
    type: "expense",
    status: "completed",
  },
  {
    id: "11",
    name: "Restaurant",
    amount: -67.80,
    date: "2024-01-08",
    account: "Credit Card",
    category: "Food & Dining",
    type: "expense",
    status: "completed",
  },
  {
    id: "12",
    name: "ATM Withdrawal",
    amount: -100.00,
    date: "2024-01-08",
    account: "Checking",
    category: "Cash",
    type: "expense",
    status: "completed",
  },
  {
    id: "13",
    name: "Dividend Payment",
    amount: 125.30,
    date: "2024-01-07",
    account: "Investment",
    category: "Investment",
    type: "income",
    status: "completed",
  },
  {
    id: "14",
    name: "Spotify Subscription",
    amount: -9.99,
    date: "2024-01-06",
    account: "Credit Card",
    category: "Entertainment",
    type: "expense",
    status: "completed",
  },
  {
    id: "15",
    name: "Uber Ride",
    amount: -23.45,
    date: "2024-01-05",
    account: "Credit Card",
    category: "Transportation",
    type: "expense",
    status: "completed",
  },
  {
    id: "16",
    name: "Interest Earned",
    amount: 15.67,
    date: "2024-01-05",
    account: "Savings",
    category: "Interest",
    type: "income",
    status: "completed",
  },
  {
    id: "17",
    name: "Pharmacy",
    amount: -28.99,
    date: "2024-01-04",
    account: "Credit Card",
    category: "Healthcare",
    type: "expense",
    status: "completed",
  },
  {
    id: "18",
    name: "Gas Station",
    amount: -58.20,
    date: "2024-01-03",
    account: "Credit Card",
    category: "Transportation",
    type: "expense",
    status: "completed",
  },
  {
    id: "19",
    name: "Bookstore",
    amount: -42.50,
    date: "2024-01-02",
    account: "Credit Card",
    category: "Education",
    type: "expense",
    status: "completed",
  },
  {
    id: "20",
    name: "Client Payment",
    amount: 1200.00,
    date: "2024-01-01",
    account: "Checking",
    category: "Income",
    type: "income",
    status: "completed",
  },
];

export function DashboardTable() {
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([]);

  const columns: ColumnDef<Transaction>[] = [
    {
      accessorKey: "name",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="h-auto p-0 hover:bg-transparent"
          >
            Transaction
            {column.getIsSorted() === "asc" ? (
              <ChevronUpIcon className="ml-2 h-4 w-4" />
            ) : column.getIsSorted() === "desc" ? (
              <ChevronDownIcon className="ml-2 h-4 w-4" />
            ) : null}
          </Button>
        );
      },
      cell: ({ row }) => (
        <div className="font-medium">{row.getValue("name")}</div>
      ),
    },
    {
      accessorKey: "amount",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="h-auto p-0 hover:bg-transparent"
          >
            Amount
            {column.getIsSorted() === "asc" ? (
              <ChevronUpIcon className="ml-2 h-4 w-4" />
            ) : column.getIsSorted() === "desc" ? (
              <ChevronDownIcon className="ml-2 h-4 w-4" />
            ) : null}
          </Button>
        );
      },
      cell: ({ row }) => {
        const amount = parseFloat(row.getValue("amount"));
        const formatted = new Intl.NumberFormat("en-US", {
          style: "currency",
          currency: "USD",
        }).format(amount);

        return (
          <div className={`font-medium ${amount >= 0 ? "text-green-600" : "text-red-600"}`}>
            {formatted}
          </div>
        );
      },
    },
    {
      accessorKey: "date",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="h-auto p-0 hover:bg-transparent"
          >
            Date
            {column.getIsSorted() === "asc" ? (
              <ChevronUpIcon className="ml-2 h-4 w-4" />
            ) : column.getIsSorted() === "desc" ? (
              <ChevronDownIcon className="ml-2 h-4 w-4" />
            ) : null}
          </Button>
        );
      },
      cell: ({ row }) => {
        const date = new Date(row.getValue("date"));
        return <div>{date.toLocaleDateString()}</div>;
      },
    },
    {
      accessorKey: "account",
      header: "Account",
      cell: ({ row }) => (
        <Badge variant="outline">{row.getValue("account")}</Badge>
      ),
    },
    {
      accessorKey: "category",
      header: "Category",
      cell: ({ row }) => (
        <div className="text-sm text-muted-foreground">{row.getValue("category")}</div>
      ),
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => {
        const status = row.getValue("status") as string;
        return (
          <Badge
            variant={
              status === "completed"
                ? "default"
                : status === "pending"
                  ? "secondary"
                  : "destructive"
            }
          >
            {status}
          </Badge>
        );
      },
    },
  ];

  const table = useReactTable({
    data: sampleData,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    state: {
      sorting,
      columnFilters,
    },
  });

  return (
    <div className="w-full">
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id}>
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                    </TableHead>
                  );
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && "selected"}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center"
                >
                  No transactions found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-end space-x-2 py-4">
        <div className="flex-1 text-sm text-muted-foreground">
          {table.getFilteredRowModel().rows.length} transaction(s) total.
        </div>
        <div className="space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
