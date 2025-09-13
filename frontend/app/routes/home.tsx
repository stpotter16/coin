import type { Route } from "./+types/home";
import { Dashboard } from "~/dashboard/dashboard";

export function meta({ }: Route.MetaArgs) {
  return [
    { title: "Coin" },
    { name: "description", content: "Coin dashboard" },
  ];
}

export default function Home() {
  return <Dashboard />;
}
