import React from "react";
import { FlatList, View } from "react-native";
import { Text } from "@/components/ui";
import type { StoryAttachment } from "@/types/attachment";
import { AttachmentCard } from "./card";
import { EmptyState } from "./empty-state";

export const List = ({ attachments }: { attachments: StoryAttachment[] }) => {
  return (
    <View className="flex-1">
      {attachments.length > 0 && (
        <View className="px-4 py-3">
          <Text color="muted" fontSize="sm">
            {attachments.length} attachment{attachments.length !== 1 ? "s" : ""}
          </Text>
        </View>
      )}
      <FlatList
        data={attachments}
        renderItem={({ item }) => <AttachmentCard attachment={item} />}
        ListEmptyComponent={<EmptyState />}
        showsVerticalScrollIndicator={false}
        keyExtractor={(item) => item.id}
      />
    </View>
  );
};
