import { usePathname } from "next/navigation";
import { Flex, Text } from "ui";
import {
  ActiveSprintIcon,
  AiIcon,
  AnalyticsIcon,
  DashboardIcon,
  ObjectiveIcon,
  OKRIcon,
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

const NavigationSection = ({
  children,
  name,
}: {
  children: ReactNode;
  name: string;
}) => (
  <Flex className="mt-3 gap-1.5" direction="column">
    <Text
      className="mb-1 flex h-7 items-center px-2.5"
      color="muted"
      fontWeight="medium"
    >
      {name}
    </Text>
    <Flex direction="column" gap={1}>
      {children}
    </Flex>
  </Flex>
);

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
      name: "Analytics",
      icon: <AnalyticsIcon />,
      href: withWorkspace("/analytics"),
    },
  ];

  const strategyLinks: MenuItem[] = [
    {
      name: getTermDisplay("objectiveTerm", {
        variant: "plural",
        capitalize: true,
      }),
      icon: <ObjectiveIcon />,
      href: withWorkspace("/objectives"),
      disabled: !features.objectiveEnabled,
    },
    {
      name: getTermDisplay("keyResultTerm", {
        variant: "plural",
        capitalize: true,
      }),
      icon: <OKRIcon />,
      href: withWorkspace("/key-results"),
      disabled: !features.objectiveEnabled || !features.keyResultEnabled,
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
    <>
      <Flex className="gap-1.5" direction="column">
        {renderLinks(primaryLinks)}
      </Flex>
      {strategyLinks.some(({ disabled }) => !disabled) ? (
        <NavigationSection name="Strategy">
          {renderLinks(strategyLinks)}
        </NavigationSection>
      ) : null}
    </>
  );
};
