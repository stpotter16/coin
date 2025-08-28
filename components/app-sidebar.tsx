import { Sidebar, SidebarContent } from "@/components/ui/sidebar";
import { Home, Landmark, Settings, User } from "lucide-react";
import { AppSidebarGroup, AppSidebarGroupProps } from "./app-sidebar-group";

const finaceMenuGroupProps: AppSidebarGroupProps = {
  label: "Finances",
  menuItems: [
    {
      title: "Home",
      url: "#",
      icon: Home
    },
    {
      title: "Accounts",
      url: "#",
      icon: Landmark
    }
  ]
}

const userMenuGroupProps: AppSidebarGroupProps = {
  label: "User",
  menuItems: [
    {
      title: "Profile",
      url: "#",
      icon: User
    },
    {
      title: "Settings",
      url: "#",
      icon: Settings
    }
  ]
}

export function AppSidebar() {
  return (
    <Sidebar>
      <SidebarContent>
        <AppSidebarGroup {...finaceMenuGroupProps} />
        <AppSidebarGroup {...userMenuGroupProps} />
      </SidebarContent>
    </Sidebar>
  )
}
