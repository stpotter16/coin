import { Sidebar, SidebarContent } from "../ui/sidebar";
import { Banknote, DollarSign, Home, Landmark, Settings, User } from "lucide-react";
import { AppSidebarGroup, type AppSidebarGroupProps } from "./app-sidebar-group";

const finaceMenuGroupProps: AppSidebarGroupProps = {
  label: "Finances",
  menuItems: [
    {
      title: "Home",
      url: "/",
      icon: Home,
    },
    {
      title: "Transactions",
      url: "/transactions",
      icon: Banknote
    },
    {
      title: "Budget",
      url: "/budget",
      icon: DollarSign
    },
    {
      title: "Accounts",
      url: "/accounts",
      icon: Landmark,
    },
  ],
};

const userMenuGroupProps: AppSidebarGroupProps = {
  label: "User",
  menuItems: [
    {
      title: "Profile",
      url: "/profile",
      icon: User,
    },
    {
      title: "Settings",
      url: "/settings",
      icon: Settings,
    },
  ],
};

export function AppSidebar() {
  return (
    <Sidebar>
      <SidebarContent>
        <AppSidebarGroup {...finaceMenuGroupProps} />
        <AppSidebarGroup {...userMenuGroupProps} />
      </SidebarContent>
    </Sidebar>
  );
}
