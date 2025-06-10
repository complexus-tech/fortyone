import { usePathname } from "next/navigation";
import { Badge, Flex } from "ui";
import { cn } from "lib";
import {
  AnalyticsIcon,
  // AnalyticsIcon,
  DashboardIcon,
  NotificationsIcon,
  RoadmapIcon,
  SprintsIcon,
  UserIcon,
} from "icons";
import type { ReactNode } from "react";
import { NavLink } from "@/components/ui";
import { useUnreadNotifications } from "@/modules/notifications/hooks/unread";
import { useTerminology, useFeatures, useFeatureFlag } from "@/hooks";
import { useRunningSprints } from "@/modules/sprints/hooks/running-sprints";

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
  const { data: runningSprints = [] } = useRunningSprints();
  const isAnalyticsEnabled = useFeatureFlag("analytics_page");
  const { getTermDisplay } = useTerminology();
  const features = useFeatures();

  const getSprintsItem = (): MenuItem | null => {
    if (runningSprints.length === 0) return null;
    const sprint = runningSprints[0];
    return {
      name: `Current ${getTermDisplay("sprintTerm", { capitalize: true, variant: runningSprints.length > 1 ? "plural" : "singular" })}`,
      icon: <SprintsIcon />,
      href:
        runningSprints.length > 1
          ? "/sprints"
          : `/teams/${sprint.teamId}/sprints/${sprint.id}/stories`,
    };
  };

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
    {
      name: "Notifications",
      icon: <NotificationsIcon className="h-[1.3rem]" />,
      href: "/notifications",
      messages: unreadNotifications,
    },
    {
      name: "Analytics",
      icon: <AnalyticsIcon />,
      href: "/analytics",
    },
    ...(getSprintsItem() ? [getSprintsItem()!] : []),
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
              data-nav-notifications={
                href === "/notifications" ? "" : undefined
              }
              data-nav-summary={href === "/summary" ? "" : undefined}
              href={href}
              key={name}
            >
              <span className="flex items-center gap-2">
                <span className="shrink-0">{icon}</span>
                <span className="line-clamp-1 first-letter:capitalize">
                  {name}
                </span>
              </span>
              {messages ? (
                <Badge className="shrink-0" rounded="full" size="sm">
                  {messages}
                </Badge>
              ) : null}
            </NavLink>
          );
        })}
    </Flex>
  );
};
