import { usePathname } from "next/navigation";
import { Badge, Flex } from "ui";
import { cn } from "lib";
import { DashboardIcon, NotificationsIcon, RoadmapIcon, UserIcon } from "icons";
import type { ReactNode } from "react";
import { NavLink } from "@/components/ui";
import { useUnreadNotifications } from "@/modules/notifications/hooks/unread";
import { useTerminology, useFeatures } from "@/hooks";

type MenuItem = {
  name: string;
  icon: ReactNode;
  href: string;
  messages?: number;
  disabled?: boolean;
};

export const Navigation = () => {
  const pathname = usePathname();
  const { data: unreadNotifications = 0 } = useUnreadNotifications();
  const { getTermDisplay } = useTerminology();
  const features = useFeatures();
  const links: MenuItem[] = [
    {
      name: "Summary",
      icon: <DashboardIcon />,
      href: "/summary",
    },
    {
      name: `My ${getTermDisplay("storyTerm", { variant: "plural" })}`,
      icon: <UserIcon />,
      href: "/my-work",
    },
    {
      name: "Roadmap",
      icon: <RoadmapIcon strokeWidth={2} />,
      href: "/roadmaps",
      disabled: !features.objectiveEnabled,
    },
    // {
    //   name: "Analytics",
    //   icon: <AnalyticsIcon />,
    //   href: "/analytics",
    // },
    {
      name: "Notifications",
      icon: <NotificationsIcon className="h-[1.3rem]" />,
      href: "/notifications",
      messages: unreadNotifications,
    },
    // {
    //   name: "Active sprints",
    //   icon: <SprintsIcon />,
    //   href: "/sprints",
    // },
  ];

  return (
    <Flex direction="column" gap={2}>
      {links
        .filter(({ disabled }) => !disabled)
        .map(({ name, icon, href, messages }) => {
          const isActive = pathname === href;
          return (
            <NavLink
              active={isActive}
              className={cn({
                "justify-between": messages,
              })}
              data-nav-my-work={href === "/my-work" ? "" : undefined}
              href={href}
              key={name}
            >
              <span className="flex items-center gap-2">
                <span>{icon}</span>
                <span className="first-letter:capitalize">{name}</span>
              </span>
              {messages ? (
                <Badge rounded="full" size="sm">
                  {messages}
                </Badge>
              ) : null}
            </NavLink>
          );
        })}
    </Flex>
  );
};
