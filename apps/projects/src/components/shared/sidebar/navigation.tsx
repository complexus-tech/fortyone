import { usePathname } from "next/navigation";
import { Badge, Flex } from "ui";
import { cn } from "lib";
import {
  AnalyticsIcon,
  HomeIcon,
  StoryIcon,
  NotificationsIcon,
  ObjectiveIcon,
  SprintsIcon,
  RoadmapIcon,
} from "icons";
import { NavLink } from "@/components/ui";

export const Navigation = () => {
  const pathname = usePathname();
  const links = [
    {
      name: "Home",
      icon: <HomeIcon className="h-5 w-auto" />,
      href: "/",
    },
    {
      name: "Reports",
      icon: <AnalyticsIcon className="h-5 w-auto" />,
      href: "/reports",
    },
    {
      name: "My Work",
      icon: <StoryIcon className="h-5 w-auto" strokeWidth={2} />,
      href: "/my-work",
    },
    {
      name: "Messages",
      icon: <NotificationsIcon className="h-5 w-auto" strokeWidth={2} />,
      href: "/notifications",
      messages: 2,
    },
    {
      name: "Roadmaps",
      icon: <RoadmapIcon className="h-5 w-auto" strokeWidth={2} />,
      href: "/roadmaps",
    },
    {
      name: "Objectives",
      icon: <ObjectiveIcon className="relative left-px h-5 w-auto" />,
      href: "/objectives",
    },
    {
      name: "Current Sprints",
      icon: <SprintsIcon className="h-5 w-auto" />,
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
