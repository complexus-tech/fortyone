import React from "react";
import { FlatList, View } from "react-native";
import { Story } from "@/modules/stories/types";
import { Row, Story as StoryComponent, Text } from "@/components/ui";
import { EmptyState } from "./empty-state";

export const List = ({ stories }: { stories: Story[] }) => {
  return (
    <View style={{ flex: 1 }}>
      {stories.length > 0 && (
        <Row asContainer className="mb-1">
          <Text color="muted">Sub Stories</Text>
        </Row>
      )}
      <FlatList
        data={stories}
        renderItem={({ item }) => <StoryComponent story={item} />}
        ListEmptyComponent={<EmptyState />}
        showsVerticalScrollIndicator={false}
        keyExtractor={(item) => item.id}
      />
    </View>
  );
};
