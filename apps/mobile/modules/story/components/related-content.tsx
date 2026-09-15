import type { DetailedStory } from "@/modules/stories/types";
import { Pressable, View, useColorScheme } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useRouter } from "expo-router";
import { Text } from "@/components/ui/Text";
import { themeColors } from "@/constants/colors";
import { useTerminology } from "@/hooks/use-terminology";
import { useStory } from "@/modules/stories/hooks/use-story";
import { useStatuses } from "@/modules/statuses/hooks/use-statuses";
import { StatusIcon } from "@/components/icons";

export function ParentStory({ parentId }: { parentId: string }) {
  const router = useRouter();
  const { data: parent } = useStory(parentId);
  const { getTermDisplay } = useTerminology();

  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={
        parent ? `Open parent: ${parent.title}` : "Open parent task"
      }
      onPress={() => router.push(`/story/${parentId}`)}
      className="min-h-[44px] flex-row flex-wrap items-center gap-[6px] px-[20px] py-[8px] active:opacity-60"
    >
      <Text color="muted" fontSize="sm">
        Sub-{getTermDisplay("storyTerm")} of
      </Text>
      <Text fontSize="sm" numberOfLines={1} className="shrink">
        {parent?.title || "View parent"}
      </Text>
    </Pressable>
  );
}

export function RelatedContent({ story }: { story: DetailedStory }) {
  const router = useRouter();
  const { data: statuses = [] } = useStatuses();
  const { getTermDisplay } = useTerminology();
  const foreground =
    themeColors[useColorScheme() === "dark" ? "dark" : "light"].textMuted;
  const subStories = story.subStories ?? [];
  const subStoryLabel = `Sub-${getTermDisplay("storyTerm", { variant: "plural" })}`;

  return (
    <View className="gap-[10px] px-[20px] pb-[20px]">
      <View className="overflow-hidden rounded-[20px] bg-gray-50 dark:bg-dark-100">
        <Pressable
          accessibilityRole="button"
          accessibilityLabel={`${subStoryLabel}, ${subStories.length}. View all`}
          onPress={() => router.push(`/story/${story.id}/sub-stories`)}
          className="min-h-[48px] flex-row items-center gap-[8px] px-[16px] py-[12px] active:opacity-60"
        >
          <Text fontSize="sm" color="muted" className="flex-1">
            {subStoryLabel}
          </Text>
          <Text fontSize="sm" color="muted">
            {subStories.length}
          </Text>
          <Ionicons name="chevron-forward" size={16} color={foreground} />
        </Pressable>
        {subStories.slice(0, 3).map((child) => {
          const status = statuses.find((item) => item.id === child.statusId);
          return (
            <Pressable
              key={child.id}
              accessibilityRole="button"
              accessibilityLabel={child.title}
              onPress={() => router.push(`/story/${child.id}`)}
              className="min-h-[48px] flex-row items-center gap-[10px] px-[16px] py-[10px] active:opacity-60"
            >
              <StatusIcon
                category={status?.category}
                color={status?.color || foreground}
                size={18}
              />
              <Text
                fontWeight="medium"
                className="flex-1"
                numberOfLines={1}
                ellipsizeMode="tail"
              >
                {child.title}
              </Text>
            </Pressable>
          );
        })}
      </View>
      <Pressable
        accessibilityRole="button"
        accessibilityLabel="View linked resources"
        onPress={() => router.push(`/story/${story.id}/links`)}
        className="min-h-[48px] flex-row items-center gap-[10px] rounded-[20px] bg-gray-50 px-[16px] py-[12px] active:opacity-60 dark:bg-dark-100"
      >
        <Ionicons name="chevron-forward" size={14} color={foreground} />
        <Text fontSize="sm" color="muted">
          Resources
        </Text>
      </Pressable>
    </View>
  );
}
