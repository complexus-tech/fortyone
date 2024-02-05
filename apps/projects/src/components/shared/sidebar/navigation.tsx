import { usePathname } from "next/navigation";
import { Flex } from "ui";
import { cn } from "lib";
import { Activity, Bell, Columns3, ListTodo, TimerReset } from "lucide-react";
import { NavLink } from "@/components/ui";

export const Navigation = () => {
  const pathname = usePathname();
  const links = [
    {
      name: "Dashboard",
      icon: <Columns3 className="h-5 w-auto" />,
      href: "/",
    },
    {
      name: "Analytics",
      icon: <Activity className="h-5 w-auto" />,
      href: "/analytics",
    },
    {
      name: "Notifications",
      icon: <Bell className="h-5 w-auto" />,
      href: "/notifications",
    },
    {
      name: "My issues",
      icon: <ListTodo className="h-5 w-auto" />,
      href: "/my-issues",
    },
    {
      name: "Sprints",
      icon: <TimerReset className="h-5 w-auto" />,
      href: "/projects",
    },
  ];

  return (
    <Flex direction="column" gap={2}>
      {links.map(({ name, icon, href }) => (
        <NavLink active={pathname === href} href={href} key={name}>
          <span
            className={cn(
              "text-gray-300/80 group-hover:text-gray-300 dark:text-gray dark:group-hover:text-white",
              {
                "text-gray-300 dark:text-white": pathname === href,
              },
            )}
          >
            {icon}
          </span>
          {name}
        </NavLink>
      ))}
    </Flex>
  );
};
