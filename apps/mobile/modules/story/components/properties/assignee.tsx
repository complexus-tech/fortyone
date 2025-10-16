import React, { useState, useMemo } from "react";
import { Badge, Text, Avatar, Row } from "@/components/ui";
import { Story } from "@/modules/stories/types";
import { Pressable, TextInput } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useColorScheme } from "nativewind";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { PropertyBottomSheet } from "./property-bottom-sheet";
import { useMembers } from "@/modules/members/hooks/use-members";

const Item = ({
  member,
  onPress,
  isSelected,
}: {
  member: {
    id: string;
    username: string;
    fullName?: string;
    avatarUrl?: string;
  };
  onPress: () => void;
  isSelected: boolean;
}) => {
  const { colorScheme } = useColorScheme();
  return (
    <Pressable
      key={member.id}
      onPress={onPress}
      className="flex-row items-center px-4 py-3.5 gap-2"
    >
      <Avatar
        size="sm"
        color="primary"
        name={member.fullName || member.username}
        src={member.avatarUrl}
      />
      <Text className="flex-1">{member.fullName || member.username}</Text>
      {isSelected && (
        <SymbolView
          name="checkmark.circle.fill"
          size={20}
          tintColor={colorScheme === "light" ? colors.black : colors.white}
          fallback={
            <Ionicons
              name="checkmark-circle"
              size={20}
              color={colorScheme === "light" ? colors.black : colors.white}
            />
          }
        />
      )}
    </Pressable>
  );
};

export const AssigneeBadge = ({
  story,
  onAssigneeChange,
}: {
  story: Story;
  onAssigneeChange: (assigneeId: string | null) => void;
}) => {
  const { colorScheme } = useColorScheme();
  const { data: members = [] } = useMembers();
  const eleigibleMembers = members.filter((m) => m.role !== "system");
  const [searchQuery, setSearchQuery] = useState("");
  const currentAssignee = members.find((m) => m.id === story.assigneeId);

  const filteredMembers = useMemo(() => {
    if (!searchQuery.trim()) return eleigibleMembers;
    return eleigibleMembers.filter((member) =>
      (member.fullName || member.username)
        .toLowerCase()
        .includes(searchQuery.toLowerCase())
    );
  }, [eleigibleMembers, searchQuery]);

  return (
    <PropertyBottomSheet
      trigger={
        <Badge color="tertiary" className="pl-1.5">
          <Avatar
            size="xs"
            name={currentAssignee?.fullName || currentAssignee?.username}
            src={currentAssignee?.avatarUrl}
          />
          <Text>{currentAssignee?.username || "No Assignee"}</Text>
        </Badge>
      }
      snapPoints={["90%"]}
    >
      <Text className="font-semibold mb-3 text-center">Assignee</Text>
      <Row
        className="bg-gray-100/60 dark:bg-dark-50 rounded-xl pl-3 pr-2.5 mx-3.5"
        align="center"
        gap={2}
      >
        <SymbolView
          name="magnifyingglass"
          size={20}
          tintColor={
            colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[200]
          }
        />
        <TextInput
          className="flex-1 h-11 font-medium text-[16px] dark:text-white"
          placeholder="Search members..."
          placeholderTextColor={
            colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[200]
          }
          value={searchQuery}
          onChangeText={setSearchQuery}
          autoFocus
        />
        {searchQuery.length > 0 && (
          <Pressable
            onPress={() => {
              setSearchQuery("");
            }}
            className="p-1"
          >
            <SymbolView
              name="xmark.circle.fill"
              size={20}
              tintColor={
                colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[200]
              }
            />
          </Pressable>
        )}
      </Row>

      {filteredMembers.slice(0, 8).map((member) => (
        <Item
          key={member.id}
          member={member}
          onPress={() => onAssigneeChange(member.id)}
          isSelected={story.assigneeId === member.id}
        />
      ))}
    </PropertyBottomSheet>
  );
};
