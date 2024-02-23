import { usePathname } from "next/navigation";
import { Badge, Flex } from "ui";
import { cn } from "lib";
import {
  AnalyticsIcon,
  DashboardIcon,
  IssuesIcon,
  NotificationsIcon,
  ProjectsIcon,
  SprintsIcon,
} from "icons";
import { NavLink } from "@/components/ui";

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
      name: "Sprints",
      icon: <SprintsIcon className="h-[1.3rem] w-auto" />,
      href: "/sprints",
    },
    {
      name: "Notifications",
      icon: <NotificationsIcon className="h-[1.3rem] w-auto" />,
      href: "/notifications",
      messages: 2,
    },
  ];

  return (
    <Flex direction="column" gap={2}>
      {links.map(({ name, icon, href, messages }) => (
        <NavLink
          active={pathname === href}
          className={cn({
            "justify-between": messages,
          })}
          href={href}
          key={name}
        >
          <span className="flex items-center gap-2">
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
          </span>
          {messages ? <Badge color="tertiary">{messages}</Badge> : null}
        </NavLink>
      ))}
    </Flex>
  );
};
