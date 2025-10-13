import { Text, Row, Back, ContextMenuButton } from "@/components/ui";
import { useGlobalSearchParams } from "expo-router";
import { useStory } from "@/modules/stories/hooks";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { Alert } from "react-native";
import * as Clipboard from "expo-clipboard";
import {
  useArchiveStoryMutation,
  useUnarchiveStoryMutation,
  useDeleteStoryMutation,
  useRestoreStoryMutation,
  useDuplicateStoryMutation,
} from "@/modules/stories/hooks/story-mutations";
import { useCurrentWorkspace } from "@/lib/hooks";

export const Header = () => {
  const { storyId } = useGlobalSearchParams<{ storyId: string }>();
  const { workspace } = useCurrentWorkspace();
  const { data: story } = useStory(storyId);
  const { data: teams = [] } = useTeams();
  const team = teams.find((team) => team.id === story?.teamId);

  const archiveMutation = useArchiveStoryMutation();
  const unarchiveMutation = useUnarchiveStoryMutation();
  const deleteMutation = useDeleteStoryMutation();
  const restoreMutation = useRestoreStoryMutation();
  const duplicateMutation = useDuplicateStoryMutation();

  const isArchived = story?.archivedAt;
  const isDeleted = story?.deletedAt;

  const handleArchive = () => {
    Alert.alert(
      "Archive Story",
      "This story will be moved to the archive and can be unarchived later. It won't appear in your active story lists.",
      [
        { text: "Cancel", style: "cancel" },
        {
          text: "Archive",
          style: "destructive",
          onPress: () => archiveMutation.mutate(storyId),
        },
      ]
    );
  };

  const handleUnarchive = () => {
    Alert.alert(
      "Unarchive Story",
      "This story will be moved back to your active stories.",
      [
        { text: "Cancel", style: "cancel" },
        {
          text: "Unarchive",
          onPress: () => unarchiveMutation.mutate(storyId),
        },
      ]
    );
  };

  const handleDelete = () => {
    const isHardDelete = Boolean(isArchived || isDeleted);
    Alert.alert(
      isHardDelete ? "Delete Forever" : "Delete Story",
      isHardDelete
        ? "This is an irreversible action. The story will be permanently deleted. You can't restore it."
        : "This story will be moved to the recycle bin and will be permanently deleted after 30 days. You can restore it at any time before that.",
      [
        { text: "Cancel", style: "cancel" },
        {
          text: isHardDelete ? "Delete Forever" : "Delete",
          style: "destructive",
          onPress: () =>
            deleteMutation.mutate({ storyId, hardDelete: isHardDelete }),
        },
      ]
    );
  };

  const handleRestore = () => {
    Alert.alert(
      "Restore Story",
      "This story will be moved back to your active stories.",
      [
        { text: "Cancel", style: "cancel" },
        {
          text: "Restore",
          onPress: () => restoreMutation.mutate(storyId),
        },
      ]
    );
  };

  const handleDuplicate = () => {
    Alert.alert("Duplicate Story", "This will create a copy of this story.", [
      { text: "Cancel", style: "cancel" },
      {
        text: "Duplicate",
        onPress: () => duplicateMutation.mutate(storyId),
      },
    ]);
  };

  const handleCopyLink = async () => {
    try {
      const storyUrl = `https://${workspace?.slug}.fortyone.app/story/${storyId}`;
      await Clipboard.setStringAsync(storyUrl);
      Alert.alert("Success", "Story link copied to clipboard");
    } catch {
      Alert.alert("Error", "Failed to copy link");
    }
  };

  const getActions = () => {
    const baseActions = [
      {
        systemImage: "pencil" as const,
        label: "Edit",
        onPress: () => {
          // TODO: Navigate to edit screen
          Alert.alert(
            "Coming Soon",
            "Edit functionality will be available soon"
          );
        },
      },
      {
        systemImage: "link" as const,
        label: "Copy link",
        onPress: handleCopyLink,
      },
    ];

    if (isDeleted) {
      return [
        ...baseActions,
        {
          systemImage: "arrow.uturn.backward" as const,
          label: "Restore",
          onPress: handleRestore,
        },
        {
          systemImage: "trash.fill" as const,
          label: "Delete forever",
          onPress: handleDelete,
        },
      ];
    }

    if (isArchived) {
      return [
        ...baseActions,
        {
          systemImage: "archivebox" as const,
          label: "Unarchive",
          onPress: handleUnarchive,
        },
        {
          systemImage: "trash.fill" as const,
          label: "Delete forever",
          onPress: handleDelete,
        },
      ];
    }

    return [
      ...baseActions,
      {
        systemImage: "doc.on.doc" as const,
        label: "Duplicate",
        onPress: handleDuplicate,
      },
      {
        systemImage: "archivebox.fill" as const,
        label: "Archive",
        onPress: handleArchive,
      },
      {
        systemImage: "trash.fill" as const,
        label: "Delete",
        onPress: handleDelete,
      },
    ];
  };

  return (
    <Row className="mb-3" justify="between" align="center" asContainer>
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        {team?.code}-
        <Text fontSize="2xl" fontWeight="semibold" color="muted">
          {story?.sequenceId}
        </Text>
      </Text>
      <ContextMenuButton actions={getActions()} />
    </Row>
  );
};
