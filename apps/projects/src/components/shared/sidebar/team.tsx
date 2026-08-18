"use client";
import { cn } from "lib";
import { usePathname } from "next/navigation";
import { useState } from "react";
import { Box, Button, Collapsible, ContextMenu, Flex, Menu, Tooltip } from "ui";
import {
  ArchiveIcon,
  // BacklogIcon,
  ChevronRightIcon,
  DeleteIcon,
  DragIcon,
  IntakeIcon,
  LogoutIcon,
  MoreHorizontalIcon,
  ObjectiveIcon,
  RequestsIcon,
  SettingsIcon,
  SprintsIcon,
  StoryIcon,
} from "icons";
import Link from "next/link";
import { useSortable } from "@dnd-kit/sortable";
import { useSession } from "@/lib/auth/client";
import {
  useTerminology,
  useFeatures,
  useUserRole,
  useSprintsEnabled,
  useWorkspacePath,
} from "@/hooks";
import { useRemoveMemberMutation } from "@/modules/teams/hooks/remove-member-mutation";
import { ConfirmDialog, NavLink, TeamColor } from "@/components/ui";
import type { Team as TeamType } from "@/modules/teams/types";
import type { TeamFeedbackSummary } from "@/modules/team-feedback/types";
// import { useTeamStatuses } from "@/lib/hooks/statuses";
import { useTeamIntegrationRequests } from "@/modules/integration-requests/hooks/use-team-requests";
import { NavCount } from "./nav-count";

export const Team = ({
  id,
  name: teamName,
  color,
  isPrivate,
  feedbackSummary,
  totalTeams,
  isOpen,
  isCollapsed,
  onOpenChange,
  onExpand,
  sortingDisabled,
}: Pick<TeamType, "id" | "name" | "color" | "isPrivate"> & {
  feedbackSummary?: TeamFeedbackSummary;
  totalTeams: number;
  isOpen: boolean;
  isCollapsed: boolean;
  onOpenChange: (open: boolean) => void;
  onExpand?: () => void;
  sortingDisabled: boolean;
}) => {
  const { getTermDisplay } = useTerminology();
  const features = useFeatures();
  const sprintsEnabled = useSprintsEnabled(id);
  const { data: session } = useSession();
  const [isLeaving, setIsLeaving] = useState(false);
  // const { data: statuses } = useTeamStatuses(id);
  const { data: pendingRequestsPage } = useTeamIntegrationRequests(id);
  const pathname = usePathname();
  const { withWorkspace } = useWorkspacePath();
  const { mutate: removeMember, isPending } = useRemoveMemberMutation();
  const { userRole } = useUserRole();
  // const hasBacklog = statuses?.some((status) => status.category === "backlog");
  const intakeCount = pendingRequestsPage?.pagination.totalCount ?? 0;
  const hasIntake = intakeCount > 0;

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    disabled: sortingDisabled,
    id,
  });

  const style = {
    transform: transform
      ? `translate3d(${transform.x}px, ${transform.y}px, 0)`
      : undefined,
    transition,
  };

  const links = [
    {
      name: "Intake",
      icon: <IntakeIcon className="h-[1.15rem]" />,
      href: withWorkspace(`/teams/${id}/requests`),
      count: intakeCount,
      disabled: !hasIntake,
    },
    {
      name: "Feedback",
      icon: <RequestsIcon className="h-[1.15rem]" />,
      href: withWorkspace(`/teams/${id}/feedback`),
      count: feedbackSummary?.unreadCount ?? 0,
      disabled: !feedbackSummary?.enabled,
    },
    // {
    //   name: "Backlog",
    //   icon: <BacklogIcon className="h-[1.15rem]" />,
    //   href: withWorkspace(`/teams/${id}/backlog`),
    //   disabled: !hasBacklog,
    // },
    {
      name: getTermDisplay("storyTerm", { variant: "plural" }),
      icon: <StoryIcon strokeWidth={2} />,
      href: withWorkspace(`/teams/${id}/stories`),
    },
    {
      name: getTermDisplay("objectiveTerm", { variant: "plural" }),
      icon: <ObjectiveIcon />,
      href: withWorkspace(`/teams/${id}/objectives`),
      disabled: !features.objectiveEnabled,
    },
    {
      name: getTermDisplay("sprintTerm", { variant: "plural" }),
      icon: <SprintsIcon />,
      href: withWorkspace(`/teams/${id}/sprints`),
      disabled: !sprintsEnabled,
    },
  ];

  return (
    <ContextMenu>
      <ContextMenu.Trigger>
        <div
          className={cn("group", {
            "opacity-80 backdrop-blur": isDragging,
          })}
          ref={setNodeRef}
          style={style}
        >
          <Collapsible onOpenChange={onOpenChange} open={isOpen}>
            <Box>
              <Flex
                align="center"
                className={cn("relative", isCollapsed && "justify-center")}
                gap={1}
                justify="between"
              >
                {!sortingDisabled && !isCollapsed ? (
                  <DragIcon
                    className={cn(
                      "absolute top-1/2 bottom-1/2 -left-2.5 h-[1.1rem] -translate-y-1/2 opacity-0 transition-opacity outline-none group-hover:opacity-100",
                      {
                        "cursor-grab": !isDragging,
                        "cursor-grabbing": isDragging,
                        "pointer-events-none cursor-default opacity-0!":
                          isOpen || totalTeams === 1,
                      },
                    )}
                    strokeWidth={3.5}
                    {...attributes}
                    {...listeners}
                  />
                ) : null}
                <Collapsible.Trigger asChild>
                  <button
                    aria-label={
                      isCollapsed
                        ? `Expand sidebar and open ${teamName} team`
                        : `${isOpen ? "Collapse" : "Expand"} ${teamName} team`
                    }
                    className={cn(
                      "focus-visible:ring-primary/40 flex h-10 min-w-0 flex-1 items-center justify-between rounded-lg pr-2 pl-3 text-left transition outline-none select-none focus-visible:ring-2",
                      isCollapsed && "w-10 flex-none justify-center px-0",
                    )}
                    onClick={(event) => {
                      if (isCollapsed) {
                        event.preventDefault();
                        onExpand?.();
                        onOpenChange(true);
                      }
                    }}
                    suppressHydrationWarning
                    type="button"
                  >
                    <Tooltip side="right" title={isCollapsed ? teamName : null}>
                      <span className="flex min-w-0 items-center gap-1.5">
                        <TeamColor
                          className={isCollapsed ? "size-4" : undefined}
                          color={color}
                        />
                        <span
                          className={cn(
                            "ml-0.5 block max-w-[15ch] truncate",
                            isCollapsed && "hidden",
                          )}
                        >
                          {teamName}
                        </span>
                        <ChevronRightIcon
                          className={cn(
                            "relative top-[0.5px] h-3.5",
                            {
                              "rotate-90": isOpen,
                            },
                            isCollapsed && "hidden",
                          )}
                          strokeWidth={3.5}
                          suppressHydrationWarning
                        />
                      </span>
                    </Tooltip>
                  </button>
                </Collapsible.Trigger>
                {!isCollapsed ? (
                  <Menu>
                    <Menu.Button>
                      <Button
                        asIcon
                        className="opacity-0 transition-opacity group-hover:opacity-100"
                        color="tertiary"
                        leftIcon={<MoreHorizontalIcon />}
                        size="sm"
                        variant="naked"
                      >
                        <span className="sr-only">Team menu</span>
                      </Button>
                    </Menu.Button>
                    <Menu.Items>
                      <Menu.Group>
                        <Menu.Item
                          className="py-0"
                          disabled={userRole !== "admin"}
                        >
                          <Link
                            className="flex items-center gap-1.5 py-1.5"
                            href={withWorkspace(
                              `/settings/workspace/teams/${id}`,
                            )}
                          >
                            <SettingsIcon />
                            Team settings
                          </Link>
                        </Menu.Item>
                        <Menu.Item
                          className="py-0"
                          disabled={userRole !== "admin"}
                        >
                          <Link
                            className="flex items-center gap-1.5 py-1.5"
                            href={withWorkspace(`/teams/${id}/archived`)}
                          >
                            <ArchiveIcon />
                            Archived{" "}
                            {getTermDisplay("storyTerm", { variant: "plural" })}
                          </Link>
                        </Menu.Item>
                        <Menu.Item
                          className="py-0"
                          disabled={userRole !== "admin"}
                        >
                          <Link
                            className="flex items-center gap-1.5 py-1.5"
                            href={withWorkspace(`/teams/${id}/deleted`)}
                          >
                            <DeleteIcon />
                            Deleted{" "}
                            {getTermDisplay("storyTerm", { variant: "plural" })}
                          </Link>
                        </Menu.Item>
                      </Menu.Group>
                      <Menu.Separator />
                      <Menu.Group>
                        <Menu.Item
                          className="text-danger"
                          disabled={totalTeams === 1}
                          onClick={() => {
                            setIsLeaving(true);
                          }}
                        >
                          <LogoutIcon className="text-danger" />
                          Leave team
                        </Menu.Item>
                      </Menu.Group>
                    </Menu.Items>
                  </Menu>
                ) : null}
              </Flex>
              {!isCollapsed ? (
                <Collapsible.Content>
                  <Flex
                    className="border-border mt-2 ml-5 border-l-[0.5px] pl-2"
                    direction="column"
                    gap={1}
                  >
                    {links.map(({ name, icon, href, count, disabled }) => {
                      if (disabled) return null;

                      const isActive =
                        href === withWorkspace("/")
                          ? pathname === href ||
                            pathname.startsWith(withWorkspace("/dashboard"))
                          : pathname.startsWith(href);
                      return (
                        <NavLink
                          active={isActive}
                          className={isActive ? "text-foreground" : undefined}
                          href={href}
                          key={name}
                        >
                          {icon}
                          <span className="flex min-w-0 flex-1 items-center justify-between gap-2">
                            <span className="capitalize">{name}</span>
                            <NavCount count={count ?? 0} />
                          </span>
                        </NavLink>
                      );
                    })}
                  </Flex>
                </Collapsible.Content>
              ) : null}
            </Box>
          </Collapsible>
        </div>
      </ContextMenu.Trigger>
      <ContextMenu.Items>
        <ContextMenu.Group>
          <ContextMenu.Item className="py-0" disabled={userRole !== "admin"}>
            <Link
              className="flex items-center gap-1.5 py-1.5"
              href={withWorkspace(`/settings/workspace/teams/${id}`)}
            >
              <SettingsIcon />
              Team settings
            </Link>
          </ContextMenu.Item>
          <ContextMenu.Item className="py-0" disabled={userRole !== "admin"}>
            <Link
              className="flex items-center gap-1.5 py-1.5"
              href={withWorkspace(`/teams/${id}/archived`)}
            >
              <ArchiveIcon />
              Archived {getTermDisplay("storyTerm", { variant: "plural" })}
            </Link>
          </ContextMenu.Item>
          <ContextMenu.Item className="py-0" disabled={userRole !== "admin"}>
            <Link
              className="flex items-center gap-1.5 py-1.5"
              href={withWorkspace(`/teams/${id}/deleted`)}
            >
              <DeleteIcon />
              Deleted {getTermDisplay("storyTerm", { variant: "plural" })}
            </Link>
          </ContextMenu.Item>
        </ContextMenu.Group>
        <ContextMenu.Separator />
        <ContextMenu.Group>
          <ContextMenu.Item
            className="text-danger"
            disabled={totalTeams === 1}
            onClick={() => {
              setIsLeaving(true);
            }}
          >
            <LogoutIcon className="text-danger" />
            Leave team
          </ContextMenu.Item>
        </ContextMenu.Group>
      </ContextMenu.Items>
      <ConfirmDialog
        description={
          isPrivate
            ? "Once you leave this team, you will not be able to rejoin later, you will need to be invited again by an admin."
            : "You can rejoin the team later from the sidebar."
        }
        isLoading={isPending}
        isOpen={isLeaving}
        loadingText="Leaving team..."
        onCancel={() => {
          setIsLeaving(false);
        }}
        onConfirm={() => {
          removeMember({
            teamId: id,
            memberId: session?.user.id ?? "",
          });
          setIsLeaving(false);
        }}
        title={`Leave ${teamName} team?`}
      />
    </ContextMenu>
  );
};
