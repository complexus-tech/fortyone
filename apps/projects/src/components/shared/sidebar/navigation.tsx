import { usePathname } from "next/navigation";
import { Flex } from "ui";
import {
  ActiveSprintIcon,
  AiIcon,
  DashboardIcon,
  OKRIcon,
  RoadmapIcon,
  UserIcon,
} from "icons";
import type { ReactNode } from "react";
import { NavLink } from "@/components/ui";
import {
  useWorkspacePath,
  useFeatures,
  useTerminology,
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
  const { getTermDisplay } = useTerminology();
  const { userRole } = useUserRole();

  const features = useFeatures();

  const getSprintsItem = (): MenuItem | null => {
    if (runningSprints.length === 0 || userRole === "guest") return null;
    const sprint = runningSprints[0];
    return {
      name: `Active ${getTermDisplay("sprintTerm", {
        variant: runningSprints.length > 1 ? "plural" : "singular",
      })}`,
      icon: <ActiveSprintIcon />,
      href:
        runningSprints.length > 1
          ? withWorkspace("/sprints")
          : withWorkspace(
              `/teams/${sprint.teamId}/sprints/${sprint.id}/stories`,
            ),
    };
  };

  const sprintItem = getSprintsItem();
  const primaryLinks: MenuItem[] = [
    {
      name: "My work",
      icon: <UserIcon />,
      href: withWorkspace("/my-work"),
    },
    {
      name: "AI Assistant",
      icon: <AiIcon />,
      href: withWorkspace("/maya"),
    },
    {
      name: "Summary",
      icon: <DashboardIcon />,
      href: withWorkspace("/summary"),
    },
    ...(sprintItem ? [sprintItem] : []),
    {
      name: "Roadmap",
      icon: <RoadmapIcon />,
      href: withWorkspace("/roadmap"),
      disabled: !features.objectiveEnabled,
    },
    {
      name: "Strategy Map",
      icon: <OKRIcon />,
      href: withWorkspace("/strategy-map"),
      disabled: !features.objectiveEnabled,
    },
  ];

  const renderLinks = (links: MenuItem[]) =>
    links.map(({ name, icon, href, disabled }) => {
      if (disabled) return null;

      const isActive = pathname === href || pathname.startsWith(`${href}/`);

      return (
        <NavLink
          active={isActive}
          className={isActive ? "text-foreground" : undefined}
          data-nav-ai-assistant={
            href === withWorkspace("/maya") ? "" : undefined
          }
          data-nav-my-work={href === withWorkspace("/my-work") ? "" : undefined}
          data-nav-summary={href === withWorkspace("/summary") ? "" : undefined}
          href={href}
          key={name}
        >
          <span className="shrink-0">{icon}</span>
          <span className="line-clamp-1 min-w-0 flex-1 first-letter:capitalize">
            {name}
          </span>
        </NavLink>
      );
    });

  return (
    <Flex className="gap-1.5" direction="column">
      {renderLinks(primaryLinks)}
    </Flex>
  );
};
