"use client";
import { cn } from "lib";
import { usePathname } from "next/navigation";
import { useState } from "react";
import { Box, Button, ContextMenu, Flex, Menu } from "ui";
import {
  ArchiveIcon,
  ArrowRight2Icon,
  BacklogIcon,
  DeleteIcon,
  DragIcon,
  LogoutIcon,
  MoreHorizontalIcon,
  ObjectiveIcon,
  SettingsIcon,
  SprintsIcon,
  StoryIcon,
} from "icons";
import Link from "next/link";
import { useSession } from "next-auth/react";
import { useSortable } from "@dnd-kit/sortable";
import {
  useLocalStorage,
  useTerminology,
  useFeatures,
  useUserRole,
  useSprintsEnabled,
} from "@/hooks";
import { useRemoveMemberMutation } from "@/modules/teams/hooks/remove-member-mutation";
import { ConfirmDialog, NavLink, TeamColor } from "@/components/ui";
import type { Team as TeamType } from "@/modules/teams/types";
import { useTeamStatuses } from "@/lib/hooks/statuses";

export const Team = ({
  id,
  name: teamName,
  color,
  isPrivate,
  totalTeams,
  idx,
}: TeamType & {
  totalTeams: number;
  idx: number;
}) => {
  const { getTermDisplay } = useTerminology();
  const features = useFeatures();
  const sprintsEnabled = useSprintsEnabled(id);
  const { data: session } = useSession();
  const [isLeaving, setIsLeaving] = useState(false);
  const [isOpen, setIsOpen] = useLocalStorage<boolean>(
    `teams:${id}:dropdown`,
    idx === 0,
  );
  const { data: statuses } = useTeamStatuses(id);
  const pathname = usePathname();
  const { mutate: removeMember, isPending } = useRemoveMemberMutation();
  const { userRole } = useUserRole();
  const hasBacklog = statuses?.some((status) => status.category === "backlog");

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
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
      name: "Backlog",
      icon: <BacklogIcon className="h-[1.15rem]" />,
      href: `/teams/${id}/backlog`,
      disabled: !hasBacklog,
    },
    {
      name: getTermDisplay("storyTerm", { variant: "plural" }),
      icon: <StoryIcon strokeWidth={2} />,
      href: `/teams/${id}/stories`,
    },
    {
      name: getTermDisplay("objectiveTerm", { variant: "plural" }),
      icon: <ObjectiveIcon />,
      href: `/teams/${id}/objectives`,
      disabled: !features.objectiveEnabled,
    },
    {
      name: getTermDisplay("sprintTerm", { variant: "plural" }),
      icon: <SprintsIcon />,
      href: `/teams/${id}/sprints`,
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
          <Box>
            <Flex align="center" className="relative" gap={1} justify="between">
              <DragIcon
                className={cn(
                  "absolute -left-2.5 bottom-1/2 top-1/2 h-[1.1rem] -translate-y-1/2 opacity-0 outline-none transition-opacity group-hover:opacity-100",
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
              <Flex
                align="center"
                className="h-10 flex-1 select-none rounded-[0.6rem] pl-3 pr-2 outline-none transition"
                justify="between"
                onClick={() => {
                  setIsOpen(!isOpen);
                }}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    setIsOpen(!isOpen);
                  }
                }}
                role="button"
                tabIndex={0}
              >
                <span className="flex items-center gap-1.5">
                  <TeamColor color={color} />
                  <span className="ml-0.5 block max-w-[15ch] truncate">
                    {teamName}
                  </span>
                  <ArrowRight2Icon
                    className={cn("relative top-[0.5px] h-3.5", {
                      "rotate-90": isOpen,
                    })}
                    strokeWidth={3.5}
                    suppressHydrationWarning
                  />
                </span>
              </Flex>
              <Menu>
                <Menu.Button>
                  <Button
                    asIcon
                    className="opacity-0 transition-opacity group-hover:opacity-100"
                    color="tertiary"
                    leftIcon={<MoreHorizontalIcon />}
                    rounded="full"
                    size="sm"
                    variant="naked"
                  >
                    <span className="sr-only">Team menu</span>
                  </Button>
                </Menu.Button>
                <Menu.Items>
                  <Menu.Group>
                    <Menu.Item className="py-0" disabled={userRole !== "admin"}>
                      <Link
                        className="flex items-center gap-1.5 py-1.5"
                        href={`/settings/workspace/teams/${id}`}
                      >
                        <SettingsIcon />
                        Team settings
                      </Link>
                    </Menu.Item>
                    <Menu.Item className="py-0" disabled={userRole !== "admin"}>
                      <Link
                        className="flex items-center gap-1.5 py-1.5"
                        href={`/teams/${id}/archived`}
                      >
                        <ArchiveIcon />
                        Archived{" "}
                        {getTermDisplay("storyTerm", { variant: "plural" })}
                      </Link>
                    </Menu.Item>
                    <Menu.Item className="py-0" disabled={userRole !== "admin"}>
                      <Link
                        className="flex items-center gap-1.5 py-1.5"
                        href={`/teams/${id}/deleted`}
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
                      className="text-danger dark:text-danger"
                      disabled={totalTeams === 1}
                      onClick={() => {
                        setIsLeaving(true);
                      }}
                    >
                      <LogoutIcon className="text-danger dark:text-danger" />
                      Leave team
                    </Menu.Item>
                  </Menu.Group>
                </Menu.Items>
              </Menu>
            </Flex>
            <Flex
              className={cn(
                "ml-5 h-0 overflow-hidden border-l border-dashed border-gray-200/80 pl-2 transition-all duration-300 dark:border-dark-50",
                {
                  "mt-2 h-max": isOpen,
                },
              )}
              direction="column"
              gap={1}
              suppressHydrationWarning
            >
              {links
                .filter(({ disabled }) => !disabled)
                .map(({ name, icon, href }) => {
                  const isActive =
                    href === "/"
                      ? pathname === href || pathname.startsWith("/dashboard")
                      : pathname.startsWith(href);
                  return (
                    <NavLink active={isActive} href={href} key={name}>
                      {icon}
                      <span className="capitalize">{name}</span>
                    </NavLink>
                  );
                })}
            </Flex>
          </Box>
        </div>
      </ContextMenu.Trigger>
      <ContextMenu.Items>
        <ContextMenu.Group>
          <ContextMenu.Item className="py-0" disabled={userRole !== "admin"}>
            <Link
              className="flex items-center gap-1.5 py-1.5"
              href={`/settings/workspace/teams/${id}`}
            >
              <SettingsIcon />
              Team settings
            </Link>
          </ContextMenu.Item>
          <ContextMenu.Item className="py-0" disabled={userRole !== "admin"}>
            <Link
              className="flex items-center gap-1.5 py-1.5"
              href={`/teams/${id}/archived`}
            >
              <ArchiveIcon />
              Archived {getTermDisplay("storyTerm", { variant: "plural" })}
            </Link>
          </ContextMenu.Item>
          <ContextMenu.Item className="py-0" disabled={userRole !== "admin"}>
            <Link
              className="flex items-center gap-1.5 py-1.5"
              href={`/teams/${id}/deleted`}
            >
              <DeleteIcon />
              Deleted {getTermDisplay("storyTerm", { variant: "plural" })}
            </Link>
          </ContextMenu.Item>
        </ContextMenu.Group>
        <ContextMenu.Separator />
        <ContextMenu.Group>
          <ContextMenu.Item
            className="text-danger dark:text-danger"
            disabled={totalTeams === 1}
            onClick={() => {
              setIsLeaving(true);
            }}
          >
            <LogoutIcon className="text-danger dark:text-danger" />
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
            memberId: session?.user?.id ?? "",
          });
          setIsLeaving(false);
        }}
        title={`Leave ${teamName} team?`}
      />
    </ContextMenu>
  );
};
