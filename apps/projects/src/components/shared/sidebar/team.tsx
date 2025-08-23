"use client";
import { cn } from "lib";
import { usePathname } from "next/navigation";
import { useState } from "react";
import { Box, Button, ContextMenu, Flex, Menu } from "ui";
import {
  ArchiveIcon,
  ArrowRightIcon,
  BacklogIcon,
  DeleteIcon,
  LogoutIcon,
  MoreHorizontalIcon,
  ObjectiveIcon,
  SettingsIcon,
  SprintsIcon,
  StoryIcon,
} from "icons";
import Link from "next/link";
import { useSession } from "next-auth/react";
import {
  useLocalStorage,
  useTerminology,
  useFeatures,
  useUserRole,
} from "@/hooks";
import { useRemoveMemberMutation } from "@/modules/teams/hooks/remove-member-mutation";
import { ConfirmDialog, NavLink, TeamColor } from "@/components/ui";
import type { Team as TeamType } from "@/modules/teams/types";

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
  const { data: session } = useSession();
  const [isLeaving, setIsLeaving] = useState(false);
  const [isOpen, setIsOpen] = useLocalStorage<boolean>(
    `teams:${id}:dropdown`,
    idx === 0,
  );
  const pathname = usePathname();
  const { mutate: removeMember, isPending } = useRemoveMemberMutation();
  const { userRole } = useUserRole();

  const links = [
    {
      name: "Backlog",
      icon: <BacklogIcon className="h-[1.15rem]" />,
      href: `/teams/${id}/backlog`,
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
      disabled: !features.sprintEnabled,
    },
  ];

  return (
    <ContextMenu>
      <ContextMenu.Trigger>
        <Box>
          <Flex align="center" gap={1} justify="between">
            <Flex
              align="center"
              className="h-[2.5rem] flex-1 select-none rounded-[0.6rem] pl-3 pr-2 outline-none transition"
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
              <span className="flex items-center gap-1">
                <TeamColor color={color} />
                <span className="ml-1 block max-w-[15ch] truncate">
                  {teamName}
                </span>
                <ArrowRightIcon
                  className={cn("relative top-px h-3.5", {
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
                  className=""
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
                  <Menu.Item disabled={userRole !== "admin"}>
                    <Link href={`/teams/${id}/settings`}>
                      <SettingsIcon />
                      Team settings
                    </Link>
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
            disabled={totalTeams === 1}
            onClick={() => {
              setIsLeaving(true);
            }}
          >
            <LogoutIcon />
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
