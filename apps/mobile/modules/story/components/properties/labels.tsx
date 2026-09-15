import type { Story } from "@/modules/stories/types";
import { useRef, useLayoutEffect } from "react";
import { useColorScheme } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { Text } from "@/components/ui";
import { themeColors } from "@/constants/colors";
import { Dot } from "@/components/icons";
import { useLabels } from "@/modules/labels/hooks/use-labels";
import { useCreateLabelMutation } from "@/modules/labels/hooks/use-create-label-mutation";
import { generateRandomColor } from "@/lib/utils/colors";
import { PropertyChip } from "./property-chip";
import { PropertyBottomSheet } from "./property-bottom-sheet";
import { togglePropertyId } from "./picker-utils";

export const LabelsBadge = ({
  story,
  disabled,
  onLabelsChange,
}: {
  story: Story;
  disabled?: boolean;
  onLabelsChange: (ids: string[]) => Promise<void>;
}) => {
  const dark = useColorScheme() === "dark";
  const { data: labels = [], isPending, error, refetch } = useLabels();
  const { mutateAsync: createLabel } = useCreateLabelMutation();
  const current = useRef(story.labels || []);
  useLayoutEffect(() => {
    current.current = story.labels || [];
  }, [story.labels]);
  const eligible = labels.filter(
    (label) =>
      (!label.teamId || label.teamId === story.teamId) &&
      !label.id.startsWith("temp-"),
  );
  return (
    <PropertyBottomSheet
      title="Labels"
      multiple
      disabled={disabled}
      loading={isPending}
      error={error}
      onRetry={() => {
        void refetch();
      }}
      trigger={
        <PropertyChip>
          <Ionicons
            name="pricetag-outline"
            size={16}
            color={themeColors[dark ? "dark" : "light"].textMuted}
          />
          <Text fontSize="sm" numberOfLines={1} style={{ flexShrink: 1 }}>
            {story.labels?.length
              ? `${story.labels.length} labels`
              : "Add labels"}
          </Text>
        </PropertyChip>
      }
      options={eligible.map((label) => ({
        id: label.id,
        label: label.name,
        icon: <Dot color={label.color} size={12} />,
      }))}
      selectedIds={story.labels || []}
      onSelect={(id) => onLabelsChange(togglePropertyId(current.current, id))}
      onCreate={async (name) => {
        const response = await createLabel({
          name,
          color: generateRandomColor({
            exclude: labels.map((label) => label.color),
          }),
          teamId: story.teamId,
        });
        if (response.error || !response.data?.id)
          throw new Error(
            response.error?.message ||
              "The new label was not returned. Try again.",
          );
        await onLabelsChange([
          ...new Set([...current.current, response.data.id]),
        ]);
      }}
    />
  );
};
