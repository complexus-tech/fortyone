import { usePathname } from "next/navigation";
import { Flex } from "ui";
import { cn } from "lib";
import { NavLink } from "@/components/ui";
import {
  AnalyticsIcon,
  DashboardIcon,
  IssuesIcon,
  NotificationsIcon,
  ProjectsIcon,
  SprintsIcon,
} from "@/components/icons";

export const Navigation = () => {
  const pathname = usePathname();
  const links = [
    {
      name: "Dashboard",
      icon: <DashboardIcon className="h-[1.35rem] w-auto" />,
      href: "/",
    },
    {
      name: "Analytics",
      icon: <AnalyticsIcon className="h-[1.35rem] w-auto" />,
      href: "/analytics",
    },
    {
      name: "My issues",
      icon: <IssuesIcon className="h-[1.4rem] w-auto" />,
      href: "/my-issues",
    },
    {
      name: "Projects",
      icon: <ProjectsIcon className="h-5 w-auto" />,
      href: "/projects",
    },
    {
      name: "Notifications",
      icon: <NotificationsIcon className="h-[1.3rem] w-auto" />,
      href: "/notifications",
    },

    {
      name: "Sprints",
      icon: <SprintsIcon className="h-[1.3rem] w-auto" />,
      href: "/sprints",
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
