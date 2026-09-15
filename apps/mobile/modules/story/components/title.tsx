import React, { useState } from "react";
import { themeColors } from "@/constants/colors";
import {
  ActivityIndicator,
  KeyboardAvoidingView,
  Modal,
  Platform,
  Pressable,
  TextInput,
  View,
} from "react-native";
import { SafeAreaProvider, SafeAreaView } from "react-native-safe-area-context";
import { Text, Badge, Col, Row, IconButton } from "@/components/ui";
import { Ionicons } from "@expo/vector-icons";
import { DetailedStory } from "@/modules/stories/types";
import { differenceInCalendarDays, addDays } from "date-fns";
import { useDraft } from "@/components/rich-text/use-draft";
import { DiscardDraftButton } from "@/components/rich-text/draft-recovery";
import { useTheme } from "@/hooks";
import { useUpdateStoryMutation } from "../hooks/use-update-story-mutation";
import { getStory } from "@/modules/stories/queries/get-story";

type TitleDraft = { title: string; baseline: string };
const isTitleDraft = (value: unknown): value is TitleDraft =>
  Boolean(
    value &&
      typeof value === "object" &&
      typeof (value as TitleDraft).title === "string" &&
      typeof (value as TitleDraft).baseline === "string",
  );

const EditTitle = ({
  story,
  onClose,
}: {
  story: DetailedStory;
  onClose: () => void;
}) => {
  const { resolvedTheme } = useTheme();
  const draft = useDraft(
    `story:${story.id}:title`,
    { title: story.title, baseline: story.title },
    isTitleDraft,
  );
  const mutation = useUpdateStoryMutation();
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const save = async () => {
    if (saving || !draft.value.title.trim()) return;
    setSaving(true);
    setError(null);
    try {
      const latest = await getStory(story.id);
      const nextTitle = draft.value.title.trim();
      if (latest.title !== draft.value.baseline && latest.title !== nextTitle) {
        setError(
          "This title changed elsewhere. Your draft is preserved; review the latest title before replacing it.",
        );
      } else {
        if (latest.title !== nextTitle) {
          await mutation.mutateAsync({
            storyId: story.id,
            payload: { title: nextTitle },
          });
        }
        await draft.clear();
        onClose();
      }
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "Your title was not saved. Try again.",
      );
    }
    setSaving(false);
  };
  return (
    <Modal
      visible
      presentationStyle="pageSheet"
      animationType="slide"
      onRequestClose={() => {
        if (!saving) onClose();
      }}
    >
      <SafeAreaProvider>
        <SafeAreaView
          style={{
            flex: 1,
            backgroundColor: themeColors[resolvedTheme].background,
          }}
        >
          <KeyboardAvoidingView
            behavior={Platform.OS === "ios" ? "padding" : undefined}
            style={{ flex: 1, padding: 20 }}
          >
            <Row
              justify="between"
              align="center"
              className="mb-[16px] gap-[12px]"
            >
              <IconButton
                icon="close"
                label="Close title editor"
                disabled={saving}
                onPress={onClose}
                style={{
                  borderRadius: 22,
                  backgroundColor: themeColors[resolvedTheme].surfaceMuted,
                }}
              />
              <Text fontWeight="semibold" className="flex-1 text-center">
                Edit title
              </Text>
              <Pressable
                accessibilityRole="button"
                accessibilityLabel={saving ? "Saving title" : "Save title"}
                accessibilityState={{
                  disabled: !draft.ready || !draft.value.title.trim() || saving,
                  busy: saving,
                }}
                disabled={!draft.ready || !draft.value.title.trim() || saving}
                onPress={() => void save()}
                className="size-[44px] items-center justify-center rounded-full bg-gray-50 dark:bg-dark-100 active:opacity-60"
              >
                {saving ? (
                  <ActivityIndicator accessibilityLabel="Saving title" />
                ) : (
                  <Ionicons
                    name="arrow-up"
                    size={24}
                    color={
                      !draft.ready || !draft.value.title.trim()
                        ? themeColors[resolvedTheme].textDisabled
                        : themeColors[resolvedTheme].foreground
                    }
                  />
                )}
              </Pressable>
            </Row>
            {!draft.ready && (
              <ActivityIndicator accessibilityLabel="Restoring title draft" />
            )}
            {draft.ready && (
              <TextInput
                accessibilityLabel="Task title"
                autoFocus
                multiline
                value={draft.value.title}
                editable={!saving}
                onChangeText={(title) =>
                  draft.update({ ...draft.valueRef.current, title })
                }
                style={{
                  fontSize: 24,
                  lineHeight: 30,
                  fontWeight: "600",
                  minHeight: 120,
                  color: themeColors[resolvedTheme].foreground,
                }}
              />
            )}
            {(error || draft.error) && (
              <Text color="danger" accessibilityRole="alert">
                {error || draft.error}
              </Text>
            )}
            <DiscardDraftButton
              onDiscard={async () => {
                await draft.clear();
                onClose();
              }}
            />
            <View style={{ marginTop: 16 }}>
              <Text color="muted" fontSize="sm">
                Changes are kept as a draft on this device.
              </Text>
            </View>
          </KeyboardAvoidingView>
        </SafeAreaView>
      </SafeAreaProvider>
    </Modal>
  );
};

export const Title = ({ story }: { story: DetailedStory }) => {
  const [editing, setEditing] = useState(false);
  const isDeleted = story.deletedAt !== null;
  const isArchived = story.archivedAt !== null;

  const getDaysLeft = () => {
    if (!story.deletedAt) return 0;
    const daysLeft = differenceInCalendarDays(
      addDays(new Date(story.deletedAt!), 30),
      new Date(),
    );
    return Math.max(0, daysLeft);
  };

  return (
    <Col asContainer className="pt-[4px]" align="stretch">
      {isDeleted && (
        <Badge color="tertiary" className="mb-3">
          <Text>{getDaysLeft()} days left in bin</Text>
        </Badge>
      )}
      {isArchived && !isDeleted && (
        <Badge color="tertiary" className="mb-3">
          <Text>Archived</Text>
        </Badge>
      )}
      <Pressable
        accessibilityRole="button"
        accessibilityLabel={`Edit title: ${story.title}`}
        disabled={isDeleted}
        onPress={() => setEditing(true)}
      >
        <Text
          fontSize="2xl"
          fontWeight="semibold"
          style={{ fontSize: 24, lineHeight: 31, letterSpacing: -0.5 }}
        >
          {story.title}
        </Text>
      </Pressable>
      {editing && <EditTitle story={story} onClose={() => setEditing(false)} />}
    </Col>
  );
};
