import { useState } from "react";
import { colors } from "../../constants/colors";
import {
  ActivityIndicator,
  Alert,
  KeyboardAvoidingView,
  Modal,
  Platform,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { useTheme } from "@/hooks";
import { useMembers } from "@/modules/members/hooks/use-members";
import type { RichTextValue } from "./content";
import RichTextEditorDOM from "./editor.dom";

type Props = {
  initialHtml: string;
  title?: string;
  saveLabel?: string;
  placeholder?: string;
  onDraft: (value: RichTextValue) => Promise<void>;
  onSave: (value: RichTextValue) => Promise<void>;
  onClose: () => void;
  onDiscard: () => Promise<void>;
};

export const RichTextEditor = ({
  initialHtml,
  title = "Description",
  saveLabel = "Save",
  placeholder = "Add details, acceptance criteria, or notes…",
  onDraft,
  onSave,
  onClose,
  onDiscard,
}: Props) => {
  const { resolvedTheme } = useTheme();
  const { data: members = [] } = useMembers();
  const [ready, setReady] = useState(false);
  const [closeRequest, setCloseRequest] = useState(0);
  const dark = resolvedTheme === "dark";

  return (
    <Modal
      visible
      animationType="slide"
      presentationStyle="fullScreen"
      onRequestClose={() => setCloseRequest((value) => value + 1)}
    >
      <SafeAreaView
        style={{
          flex: 1,
          backgroundColor: dark ? colors.dark.DEFAULT : colors.white,
        }}
      >
        <KeyboardAvoidingView
          behavior={Platform.OS === "ios" ? "padding" : undefined}
          style={{ flex: 1 }}
        >
          {!ready && (
            <View
              style={{
                position: "absolute",
                inset: 0,
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              <ActivityIndicator accessibilityLabel="Loading editor" />
            </View>
          )}
          <RichTextEditorDOM
            initialHtml={initialHtml}
            dark={dark}
            title={title}
            saveLabel={saveLabel}
            placeholder={placeholder}
            mentions={members
              .filter((member) => member.isActive)
              .map((member) => ({
                id: member.id,
                label: member.fullName || member.username,
              }))}
            closeRequest={closeRequest}
            onDraft={onDraft}
            onReady={async () => setReady(true)}
            onSave={async (value) => {
              try {
                await onSave(value);
                return {};
              } catch (cause) {
                return {
                  error:
                    cause instanceof Error
                      ? cause.message
                      : "Your changes were not saved. Please try again.",
                };
              }
            }}
            onClose={async () => onClose()}
            onDiscard={() =>
              new Promise<void>((resolve, reject) => {
                Alert.alert(
                  "Discard this draft?",
                  "Only the unsaved draft on this device will be removed.",
                  [
                    {
                      text: "Keep editing",
                      style: "cancel",
                      onPress: () => resolve(),
                    },
                    {
                      text: "Discard draft",
                      style: "destructive",
                      onPress: () => {
                        void onDiscard().then(resolve, reject);
                      },
                    },
                  ],
                );
              })
            }
            dom={{
              scrollEnabled: false,
              style: { flex: 1 },
              keyboardDisplayRequiresUserAction: false,
            }}
          />
        </KeyboardAvoidingView>
      </SafeAreaView>
    </Modal>
  );
};
