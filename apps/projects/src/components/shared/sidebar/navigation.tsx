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
      name: "Overview",
      icon: <HomeIcon className="h-[1.3rem] w-auto" />,
      href: "/",
    },
    {
      name: "My Work",
      icon: <WorkIcon className="h-[1.3rem] w-auto" strokeWidth={2} />,
      href: "/my-work",
    },
    {
      name: "Reporting",
      icon: <AnalyticsIcon className="h-[1.3rem] w-auto" />,
      href: "/reports",
    },

    // {
    //   name: "Messages",
    //   icon: <ChatIcon className="h-[1.3rem] w-auto" strokeWidth={2} />,
    //   href: "/messages",
    //   messages: 1,
    // },
    {
      name: "Roadmap",
      icon: <RoadmapIcon className="h-[1.3rem] w-auto" strokeWidth={2} />,
      href: "/roadmaps",
    },
    // {
    //   name: "Objectives",
    //   icon: <ObjectiveIcon className="relative left-px h-[1.3rem] w-auto" />,
    //   href: "/objectives",
    // },
    {
      name: "Notifications",
      icon: <NotificationsIcon className="h-[1.3rem] w-auto" strokeWidth={2} />,
      href: "/notifications",
      messages: 3,
    },
    // {
    //   name: "Running Sprints",
    //   icon: <SprintsIcon className="h-[1.3rem] w-auto" />,
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
