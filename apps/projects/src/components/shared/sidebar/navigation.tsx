import { usePathname } from "next/navigation";
import { Badge, Flex } from "ui";
import { cn } from "lib";
import {
  AnalyticsIcon,
  HomeIcon,
  NotificationsIcon,
  ObjectiveIcon,
  SprintsIcon,
  RoadmapIcon,
  WorkIcon,
} from "icons";
import { NavLink } from "@/components/ui";

type MenuItem = {
  name: string;
  icon: JSX.Element;
  href: string;
  messages?: number;
};

export const Navigation = () => {
  const pathname = usePathname();
  const links: MenuItem[] = [
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
      name: "My Work",
      icon: <WorkIcon className="h-[1.3rem] w-auto" strokeWidth={2} />,
      href: "/my-work",
    },
    {
      name: "Roadmaps",
      icon: <RoadmapIcon className="h-[1.3rem] w-auto" strokeWidth={2} />,
      href: "/roadmaps",
    },
    {
      name: "Objectives",
      icon: <ObjectiveIcon className="relative left-px h-[1.3rem] w-auto" />,
      href: "/objectives",
    },
    {
      name: "Notifications",
      icon: <NotificationsIcon className="h-[1.3rem] w-auto" strokeWidth={2} />,
      href: "/notifications",
      messages: 3,
    },
    {
      name: "Active Milestones",
      icon: <SprintsIcon className="h-[1.3rem] w-auto" />,
      href: "/sprints",
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
              <Badge color="primary" rounded="full" size="sm">
                {messages}
              </Badge>
            ) : null}
          </NavLink>
        );
      })}
    </Flex>
  );
};
