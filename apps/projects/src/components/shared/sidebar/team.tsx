"use client";
import { cn } from "lib";
import { usePathname } from "next/navigation";
import { useState } from "react";
import { Box, ContextMenu, Flex } from "ui";
import {
  ArrowDownIcon,
  LogoutIcon,
  ObjectiveIcon,
  SettingsIcon,
  SprintsIcon,
  StoryIcon,
} from "icons";
import Link from "next/link";
import { useSession } from "next-auth/react";
import { useLocalStorage, useTerminology, useFeatures } from "@/hooks";
import { useRemoveMemberMutation } from "@/modules/teams/hooks/remove-member-mutation";
import { ConfirmDialog, NavLink, TeamColor } from "@/components/ui";
import type { Team as TeamType } from "@/modules/teams/types";

export const Team = ({
  id,
  name: teamName,
  color,
  isPrivate,
  totalTeams,
}: TeamType & {
  totalTeams: number;
}) => {
  const { getTermDisplay } = useTerminology();
  const features = useFeatures();
  const { data: session } = useSession();
  const [isLeaving, setIsLeaving] = useState(false);
  const [isOpen, setIsOpen] = useLocalStorage<boolean>(
    `teams:${id}:dropdown`,
    false,
  );
  const pathname = usePathname();
  const { mutate: removeMember, isPending } = useRemoveMemberMutation();

  const links = [
    {
      name: getTermDisplay("storyTerm", { variant: "plural" }),
      icon: <StoryIcon strokeWidth={2} />,
      href: `/teams/${id}/stories`,
    },
    {
      name: getTermDisplay("objectiveTerm", { variant: "plural" }),
      icon: <ObjectiveIcon strokeWidth={2} />,
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
          <Flex
            align="center"
            className="group h-[2.5rem] select-none rounded-lg pl-3 pr-2 outline-none transition hover:bg-gray-250/5 focus:bg-gray-250/5 hover:dark:bg-dark-50/60 focus:dark:bg-dark-50/60"
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
            <span className="flex items-center gap-2">
              <TeamColor color={color} />
              <span className="block max-w-[15ch] truncate">{teamName}</span>
            </span>
            <ArrowDownIcon
              className={cn(
                "h-3.5 w-auto -rotate-90 text-gray dark:text-gray-300",
                {
                  "rotate-0": isOpen,
                },
              )}
              strokeWidth={3.5}
              suppressHydrationWarning
            />
          </Flex>
          <Flex
            className={cn(
              "ml-5 h-0 overflow-hidden border-l border-gray-200/60 pl-2 transition-all duration-300 dark:border-dark-50",
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
          <ContextMenu.Item className="py-0">
            <Link
              className="flex items-center gap-1.5 py-1.5"
              href={`/settings/workspace/teams/${id}`}
            >
              <SettingsIcon />
              Team settings
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
        onClose={() => {
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
