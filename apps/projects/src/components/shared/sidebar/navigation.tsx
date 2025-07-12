import { usePathname } from "next/navigation";
import { Flex } from "ui";
import {
  AnalyticsIcon,
  DashboardIcon,
  GridIcon,
  RoadmapIcon,
  UserIcon,
} from "icons";
import type { ReactNode } from "react";
import { NavLink } from "@/components/ui";
import {
  useTerminology,
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
  const { data: runningSprints = [] } = useRunningSprints();
  const { userRole } = useUserRole();
  const isAnalyticsEnabled = useFeatureFlag("analytics_page");
  const { getTermDisplay } = useTerminology();
  const features = useFeatures();

  const getSprintsItem = (): MenuItem | null => {
    if (runningSprints.length === 0 || userRole !== "admin") return null;
    const sprint = runningSprints[0];
    return {
      name: `Active Board${runningSprints.length > 1 ? "s" : ""}`,
      icon: <GridIcon />,
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
    ...(getSprintsItem() ? [getSprintsItem()!] : []),
    {
      name: "Analytics",
      icon: <AnalyticsIcon />,
      href: "/analytics",
      disabled: !isAnalyticsEnabled,
    },
    {
      name: "Roadmap",
      icon: <RoadmapIcon strokeWidth={2} />,
      href: "/roadmaps",
      disabled: !features.objectiveEnabled,
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
              data-nav-my-work={href === "/my-work" ? "" : undefined}
              data-nav-summary={href === "/summary" ? "" : undefined}
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
