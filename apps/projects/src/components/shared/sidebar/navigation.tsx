import { usePathname } from "next/navigation";
import { Badge, Flex } from "ui";
import { cn } from "lib";
import {
  DashboardIcon,
  NotificationsIcon,
  ObjectiveIcon,
  RoadmapIcon,
  UserIcon,
} from "icons";
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

    // {
    //   name: "Analytics",
    //   icon: (
    //     <PriorityIcon
    //       className="h-5 w-auto text-gray dark:text-gray-300"
    //       priority="High"
    //     />
    //   ),
    //   href: "/analytics",
    // },
    {
      name: "Roadmap",
      icon: <RoadmapIcon strokeWidth={2} />,
      href: "/roadmaps",
    },
    {
      name: getTermDisplay("objectiveTerm", { variant: "plural" }),
      icon: <ObjectiveIcon className="relative -top-[0.5px] left-px" />,
      href: "/objectives",
      disabled: !features.objectiveEnabled,
    },
    {
      name: "Notifications",
      icon: <NotificationsIcon className="h-[1.3rem]" />,
      href: "/notifications",
      messages: unreadNotifications,
    },
    // {
    //   name: "Running sprints",
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
