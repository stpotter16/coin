import { Form } from "react-router"
import { Button } from "./ui/button"
import { Separator } from "./ui/separator"
import { SidebarTrigger } from "./ui/sidebar"

export function SiteHeader() {
  return (
    <header className="flex h-20 shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear">
      <div className="flex w-full items-center gap-1 px-4 py-4 lg:gap-2 lg:px-6">
        <SidebarTrigger className="-ml-1" />
        <Separator
          orientation="vertical"
          className="mx-2 data-[orientation=vertical]:h-4"
        />
        <h1 className="text-base font-medium">Dashboard</h1>
        <div className="ml-auto flex items-center gap-2">
          <Form method="post" action="/signout">
            <Button 
              type="submit" 
              variant="ghost" 
              size="sm" 
              className="hidden sm:flex dark:text-foreground"
            >
              Logout
            </Button>
          </Form>
        </div>
      </div>
    </header>
  )
}

