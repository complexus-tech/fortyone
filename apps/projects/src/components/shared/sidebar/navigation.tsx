import { usePathname } from "next/navigation";
import { Flex } from "ui";
import {
  AiIcon,
  AnalyticsIcon,
  DashboardIcon,
  GridIcon,
  RoadmapIcon,
  UserIcon,
} from "icons";
import type { ReactNode } from "react";
import { NavLink } from "@/components/ui";
import {
  useWorkspacePath,
  useFeatures,
  useFeatureFlag,
  useUserRole,
} from "@/hooks";
import { useRunningSprints } from "@/modules/sprints/hooks/running-sprints";

type MenuItem = {
  name: string;
  icon: ReactNode;
  href: string;
  disabled?: boolean;
};

export const Navigation = () => {
  const pathname = usePathname();
  const { withWorkspace } = useWorkspacePath();
  const { data: runningSprints = [] } = useRunningSprints();
  const { userRole } = useUserRole();
  const isAnalyticsEnabled = useFeatureFlag("analytics_page");

  const features = useFeatures();

  const getSprintsItem = (): MenuItem | null => {
    if (runningSprints.length === 0 || userRole === "guest") return null;
    const sprint = runningSprints[0];
    return {
      name: `Active Board${runningSprints.length > 1 ? "s" : ""}`,
      icon: <GridIcon />,
      href:
        runningSprints.length > 1
          ? withWorkspace("/sprints")
          : withWorkspace(`/teams/${sprint.teamId}/sprints/${sprint.id}/stories`),
    };
  };

  const links: MenuItem[] = [
    {
      name: "My work",
      icon: <UserIcon />,
      href: withWorkspace("/my-work"),
    },
    {
      name: "Roadmap",
      icon: <RoadmapIcon strokeWidth={2} />,
      href: withWorkspace("/roadmaps"),
      disabled: !features.objectiveEnabled,
    },
    {
      name: "Summary",
      icon: <DashboardIcon />,
      href: withWorkspace("/summary"),
    },

    {
      name: "AI Assistant",
      icon: <AiIcon />,
      href: withWorkspace("/maya"),
    },
    ...(getSprintsItem() ? [getSprintsItem()!] : []),

    {
      name: "Analytics",
      icon: <AnalyticsIcon />,
      href: withWorkspace("/analytics") ,
      disabled: !isAnalyticsEnabled,
    },
  ];

  return (
    <Flex className="gap-1.5" direction="column">
      {links
        .filter(({ disabled }) => !disabled)
        .map(({ name, icon, href }) => {
          const isActive = pathname === href;
          return (
            <NavLink
              active={isActive}
              data-nav-my-work={href === withWorkspace("/my-work") ? "" : undefined}
              data-nav-summary={href === withWorkspace("/summary") ? "" : undefined}
              data-nav-ai-assistant={href === withWorkspace("/maya") ? "" : undefined}
              href={href}
              key={name}
            >
              <span className="shrink-0">{icon}</span>
              <span className="line-clamp-1 first-letter:capitalize">
                {name}
              </span>
            </NavLink>
          );
        })}
    </Flex>
  );
};
