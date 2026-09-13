import { Alert } from "react-native";
import { Button } from "@/components/ui";
import { toast } from "sonner-native";

export const DiscardDraftButton = ({
  onDiscard,
}: {
  onDiscard: () => Promise<void>;
}) => (
  <Button
    color="tertiary"
    isDestructive
    onPress={() => {
      Alert.alert(
        "Discard this draft?",
        "The saved content on the server will stay unchanged.",
        [
          { text: "Keep draft", style: "cancel" },
          {
            text: "Discard draft",
            style: "destructive",
            onPress: () => {
              void onDiscard().catch(() =>
                toast.error("Could not discard this draft. Try again."),
              );
            },
          },
        ],
      );
    }}
  >
    Discard draft
  </Button>
);
