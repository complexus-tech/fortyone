import { usePathname } from "next/navigation";
import { Badge, Flex } from "ui";
import { cn } from "lib";
import {
  AnalyticsIcon,
  HomeIcon,
  IssueIcon,
  NotificationsIcon,
  ProjectsIcon,
  SprintsIcon,
} from "icons";
import { NavLink } from "@/components/ui";

export const Navigation = () => {
  const pathname = usePathname();
  const links = [
    {
      name: "Home",
      icon: <HomeIcon className="h-[1.3rem] w-auto" />,
      href: "/",
    },
    {
      name: "Reports",
      icon: <AnalyticsIcon className="h-[1.3rem] w-auto" />,
      href: "/reports",
    },
    {
      name: "My issues",
      icon: <IssueIcon className="h-[1.3rem] w-auto" strokeWidth={2} />,
      href: "/my-issues",
    },
    {
      name: "Projects",
      icon: <ProjectsIcon className="relative left-px h-[1.1rem] w-auto" />,
      href: "/projects",
    },
    {
      name: "Active sprints",
      icon: <SprintsIcon className="h-[1.25rem] w-auto" />,
      href: "/sprints",
    },
    {
      name: "Notifications",
      icon: (
        <NotificationsIcon className="h-[1.25rem] w-auto" strokeWidth={2} />
      ),
      href: "/notifications",
      messages: 2,
    },
  ];

  return (
    <Flex direction="column" gap={2}>
      {links.map(({ name, icon, href, messages }) => {
        const isActive = pathname === href;
        return (
          <NavLink
            active={isActive}
            className={cn({
              "justify-between": messages,
            })}
            href={href}
            key={name}
          >
            <span className="flex items-center gap-2">
              <span
                className={cn(
                  "text-gray-300/80 group-hover:text-gray-300 dark:text-gray dark:group-hover:text-gray-200",
                  {
                    "text-gray-300 dark:text-gray-200": isActive,
                  },
                )}
              >
                {icon}
              </span>
              {name}
            </span>
            {messages ? (
              <Badge color="tertiary" rounded="full">
                {messages}
              </Badge>
            ) : null}
          </NavLink>
        );
      })}
    </Flex>
  );
};
