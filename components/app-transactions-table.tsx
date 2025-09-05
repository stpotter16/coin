"use client";

import { useState } from "react";
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  ColumnDef,
} from "@tanstack/react-table";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

type Transaction = {
  id: string;
  name: string;
  amount: number;
  date: string;
};

const mockData: Transaction[] = [
  { id: "1", name: "Grocery Shopping", amount: -85.32, date: "2024-01-15" },
  { id: "2", name: "Salary Deposit", amount: 3200.00, date: "2024-01-14" },
  { id: "3", name: "Electric Bill", amount: -124.50, date: "2024-01-13" },
  { id: "4", name: "Coffee Shop", amount: -5.75, date: "2024-01-12" },
  { id: "5", name: "Investment Return", amount: 450.25, date: "2024-01-11" },
  { id: "6", name: "Gas Station", amount: -42.10, date: "2024-01-10" },
  { id: "7", name: "Netflix Subscription", amount: -15.99, date: "2024-01-09" },
  { id: "8", name: "Freelance Payment", amount: 850.00, date: "2024-01-08" },
  { id: "9", name: "Restaurant", amount: -67.45, date: "2024-01-07" },
  { id: "10", name: "ATM Withdrawal", amount: -100.00, date: "2024-01-06" },
  { id: "11", name: "Rent Payment", amount: -1200.00, date: "2024-01-05" },
  { id: "12", name: "Bonus", amount: 500.00, date: "2024-01-04" },
  { id: "13", name: "Uber Ride", amount: -18.25, date: "2024-01-03" },
  { id: "14", name: "Amazon Purchase", amount: -89.99, date: "2024-01-02" },
  { id: "15", name: "Bank Interest", amount: 12.50, date: "2024-01-01" },
  { id: "16", name: "Spotify Premium", amount: -9.99, date: "2023-12-31" },
  { id: "17", name: "Pharmacy", amount: -24.75, date: "2023-12-30" },
  { id: "18", name: "Cash Back Reward", amount: 35.00, date: "2023-12-29" },
  { id: "19", name: "Internet Bill", amount: -79.99, date: "2023-12-28" },
  { id: "20", name: "Dividend Payment", amount: 125.80, date: "2023-12-27" },
];

const columns: ColumnDef<Transaction>[] = [
  {
    accessorKey: "name",
    header: "Name",
  },
  {
    accessorKey: "amount",
    header: "Amount",
    cell: ({ row }) => {
      const amount = row.getValue("amount") as number;
      const formatted = new Intl.NumberFormat("en-US", {
        style: "currency",
        currency: "USD",
      }).format(amount);
      
      return (
        <span className={amount >= 0 ? "text-green-600" : "text-red-600"}>
          {formatted}
        </span>
      );
    },
  },
  {
    accessorKey: "date",
    header: "Date",
    cell: ({ row }) => {
      const date = new Date(row.getValue("date") as string);
      return date.toLocaleDateString();
    },
  },
];

export function AppTransactionsTable() {
  const [data] = useState(mockData);

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <TableHead key={header.id}>
                  {header.isPlaceholder
                    ? null
                    : flexRender(
                        header.column.columnDef.header,
                        header.getContext()
                      )}
                </TableHead>
              ))}
            </TableRow>
          ))}
        </TableHeader>
        <TableBody>
          {table.getRowModel().rows?.length ? (
            table.getRowModel().rows.map((row) => (
              <TableRow key={row.id}>
                {row.getVisibleCells().map((cell) => (
                  <TableCell key={cell.id}>
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))
          ) : (
            <TableRow>
              <TableCell colSpan={columns.length} className="h-24 text-center">
                No results.
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  );
}