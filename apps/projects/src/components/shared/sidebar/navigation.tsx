import { usePathname } from "next/navigation";
import { Badge, Flex, Text } from "ui";
import { cn } from "lib";
import {
  AnalyticsIcon,
  HomeIcon,
  NotificationsIcon,
  ObjectiveIcon,
  SprintsIcon,
  RoadmapIcon,
  WorkIcon,
  ChatIcon,
  UserIcon,
} from "icons";
import { NavLink, PriorityIcon } from "@/components/ui";

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
      icon: <HomeIcon />,
      href: "/",
    },
    {
      name: "My Work",
      icon: <UserIcon strokeWidth={2} />,
      href: "/my-work",
    },
    {
      name: "Analytics",
      icon: (
        <PriorityIcon
          priority="High"
          className="h-5 w-auto text-gray dark:text-gray-300"
        />
      ),
      href: "/reports",
    },

    // {
    //   name: "Messages",
    //   icon: <ChatIcon  strokeWidth={2} />,
    //   href: "/messages",
    //   messages: 1,
    // },
    {
      name: "Roadmap",
      icon: <RoadmapIcon strokeWidth={2} />,
      href: "/roadmaps",
    },
    // {
    //   name: "Objectives",
    //   icon: <ObjectiveIcon className="relative left-px h-5 w-auto text-gray dark:text-gray-300" />,
    //   href: "/objectives",
    // },
    {
      name: "Notifications",
      icon: <NotificationsIcon className="h-[1.3rem]" />,
      href: "/notifications",
      messages: 3,
    },
    // {
    //   name: "Running Sprints",
    //   icon: <SprintsIcon  />,
    //   href: "/sprints",
    // },
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
              <span>{icon}</span>
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
