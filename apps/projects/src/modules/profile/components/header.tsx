"use client";
import { Avatar, BreadCrumbs, Flex } from "ui";
import { UserIcon } from "icons";
import { useHotkeys } from "react-hotkeys-hook";
import { useParams } from "next/navigation";
import { HeaderContainer } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import { StoriesViewOptionsButton, LayoutSwitcher } from "@/components/ui";
import { useMembers } from "@/lib/hooks/members";
import { useProfile } from "./provider";

export const Header = ({
  layout,
  setLayout,
}: {
  layout: StoriesLayout;
  setLayout: (value: StoriesLayout) => void;
}) => {
  const { viewOptions, setViewOptions } = useProfile();
  const { userId } = useParams<{
    userId: string;
  }>();
  const { data: members = [] } = useMembers();
  const member = members.find((member) => member.id === userId);

  useHotkeys("v+l", () => {
    setLayout("list");
  });

  useHotkeys("v+k", () => {
    setLayout("kanban");
  });
  return (
    <HeaderContainer className="justify-between">
      <Flex align="center" gap={2}>
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Profile",
              icon: <UserIcon className="h-4 w-auto" />,
            },
            {
              name: member?.fullName || member?.username || "",
              icon: (
                <Avatar
                  name={member?.fullName || member?.username}
                  size="xs"
                  src={member?.avatarUrl}
                />
              ),
            },
          ]}
        />
      </Flex>
      <Flex align="center" gap={2}>
        <LayoutSwitcher
          layout={layout}
          options={["list", "kanban"]}
          setLayout={setLayout}
        />
        <StoriesViewOptionsButton
          groupByOptions={["status", "priority"]}
          layout={layout}
          setViewOptions={setViewOptions}
          viewOptions={viewOptions}
        />
      </Flex>
    </HeaderContainer>
  );
};
