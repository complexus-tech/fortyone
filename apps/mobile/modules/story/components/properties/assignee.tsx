import React, { useState, useMemo } from "react";
import { Badge, Text, Avatar } from "@/components/ui";
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
      className="flex-row items-center p-4 gap-2"
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
      snapPoints={["80%"]}
    >
      <Text className="font-semibold mb-2 text-center">Assignee</Text>
      {/* Search Input */}
      <Pressable className="mx-4 mb-4 p-3 bg-gray-50 dark:bg-dark-100 rounded-lg">
        <TextInput
          value={searchQuery}
          onChangeText={setSearchQuery}
          placeholder="Search members..."
          placeholderTextColor={colors.gray[300]}
          className="text-base"
          style={{
            color: colors.black,
          }}
        />
      </Pressable>

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
