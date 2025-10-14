import React, { useState } from "react";
import { ScrollView } from "react-native";
import { Properties } from "./components/properties";
import { useGlobalSearchParams } from "expo-router";
import { useStory } from "../stories/hooks/use-story";
import { Activity } from "./components/activity";
import { Description } from "./components/descrition";
import { Title } from "./components/title";
import { StorySkeleton } from "./components/story-skeleton";
import { BottomSheetModal } from "@/components/ui";
import { Text, DateTimePicker } from "@expo/ui/swift-ui";

export const Story = () => {
  const { storyId } = useGlobalSearchParams<{ storyId: string }>();
  const { data: story, isPending } = useStory(storyId);
  const [selectedDate, setSelectedDate] = useState(new Date());

  if (isPending) {
    return <StorySkeleton />;
  }

  return (
    <ScrollView
      className="flex-1"
      contentContainerStyle={{ paddingBottom: 100 }}
    >
      <Title story={story!} />
      <Properties story={story!} />
      <Description story={story!} />
      <Activity story={story!} />
      <BottomSheetModal isOpen spacing={10} onClose={() => {}}>
        <DateTimePicker
          onDateSelected={(date) => {
            setSelectedDate(date);
          }}
          displayedComponents="date"
          initialDate={selectedDate.toISOString()}
          variant="graphical"
          color="red"
        />
      </BottomSheetModal>
    </ScrollView>
  );
};
