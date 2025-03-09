import { usePathname } from "next/navigation";
import { Badge, Flex } from "ui";
import { cn } from "lib";
import {
  DashboardIcon,
  NotificationsIcon,
  ObjectiveIcon,
  UserIcon,
} from "icons";
import type { ReactNode } from "react";
import { NavLink } from "@/components/ui";

type MenuItem = {
  name: string;
  icon: ReactNode;
  href: string;
  messages?: number;
};

export const Navigation = () => {
  const pathname = usePathname();
  const links: MenuItem[] = [
    {
      name: "Inbox",
      icon: <NotificationsIcon className="h-[1.3rem]" />,
      href: "/notifications",
      // messages: 3,
    },

    {
      name: "My Work",
      icon: <UserIcon />,
      href: "/my-work",
    },
    {
      name: "Summary",
      icon: <DashboardIcon />,
      href: "/summary",
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
    // {
    //   name: "Roadmap",
    //   icon: <RoadmapIcon strokeWidth={2} />,
    //   href: "/roadmaps",
    // },
    {
      name: "Objectives",
      icon: <ObjectiveIcon className="relative -top-[0.5px] left-px" />,
      href: "/objectives",
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
