import type { Ref } from "react";
import { useEffect, useImperativeHandle, useRef, useState } from "react";
import {
  ActivityIndicator,
  Alert,
  KeyboardAvoidingView,
  Modal,
  Platform,
  View,
} from "react-native";
import { SafeAreaProvider, SafeAreaView } from "react-native-safe-area-context";
import { Button, Text } from "@/components/ui";
import { themeColors } from "../../constants/colors";
import { useTheme } from "@/hooks";
import { useMembers } from "@/modules/members/hooks/use-members";
import type { RichTextValue } from "./content";
import { createEditorFlushController } from "./flush-protocol";
import RichTextEditorDOM, { type EditorDOMHandle } from "./editor.dom";
import type { EditorMetadata } from "./metadata-icon";

type Props = {
  initialHtml: string;
  title?: string;
  saveLabel?: string;
  placeholder?: string;
  readOnly?: boolean;
  onDraft: (value: RichTextValue) => Promise<void>;
  onSave: (value: RichTextValue) => Promise<void>;
  onClose: () => void | Promise<void>;
  onDiscard: () => Promise<void>;
};

export type RichTextEditorHandle = {
  /** Freezes editing and resolves only after this exact snapshot is persisted. */
  flush: () => Promise<RichTextValue>;
  resume: () => void;
  requestClose: () => void;
};

type SurfaceProps = Omit<Props, "onSave" | "onClose" | "onDiscard"> & {
  ref?: Ref<RichTextEditorHandle>;
  onSave?: Props["onSave"];
  onClose?: Props["onClose"];
  onDiscard?: Props["onDiscard"];
  inline?: boolean;
  disabled?: boolean;
  metadata?: EditorMetadata[];
  onMetadataPress?: (key: string) => void;
  onReadyChange?: (ready: boolean) => void;
  onReadOnlyChange?: (readOnly: boolean) => void;
};

export const RichTextEditorSurface = ({
  ref,
  initialHtml,
  title = "Description",
  saveLabel = "Save",
  placeholder = "Add details, acceptance criteria, or notes…",
  onDraft,
  onSave,
  onClose,
  onDiscard,
  inline = false,
  disabled = false,
  readOnly = false,
  metadata = [],
  onMetadataPress,
  onReadyChange,
  onReadOnlyChange,
}: SurfaceProps) => {
  const { resolvedTheme } = useTheme();
  const { data: members = [] } = useMembers();
  const domRef = useRef<EditorDOMHandle>(null);
  const readyRef = useRef(false);
  const mounted = useRef(true);
  const [ready, setReady] = useState(false);
  const [domError, setDOMError] = useState<string | null>(null);
  const [generation, setGeneration] = useState(0);
  const [closeRequest, setCloseRequest] = useState(0);
  const [flushController] = useState(() => createEditorFlushController());

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
      flushController.fail(
        new Error("The editor closed before saving finished."),
      );
    };
  }, [flushController]);

  useImperativeHandle(
    ref,
    () => ({
      flush: () =>
        flushController.request((requestId) => {
          if (!mounted.current || !readyRef.current || !domRef.current?.flush)
            throw new Error(
              "The editor is still loading. Please try again in a moment.",
            );
          domRef.current.flush(requestId);
        }),
      resume: () => domRef.current?.resume(),
      requestClose: () => setCloseRequest((value) => value + 1),
    }),
    [flushController],
  );

  const fail = (message: string) => {
    readyRef.current = false;
    setReady(false);
    setDOMError(message);
    onReadyChange?.(false);
    flushController.fail(new Error(message));
  };

  return (
    <View style={{ flex: 1, minWidth: 0, width: "100%" }}>
      {!ready && !domError && (
        <View
          pointerEvents="none"
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
      {domError && (
        <View style={{ gap: 12, padding: 20 }}>
          <Text color="danger" accessibilityRole="alert">
            {domError}
          </Text>
          <Button
            color="tertiary"
            onPress={() => {
              setDOMError(null);
              readyRef.current = false;
              setReady(false);
              setGeneration((value) => value + 1);
            }}
          >
            Reload editor
          </Button>
        </View>
      )}
      <RichTextEditorDOM
        key={generation}
        ref={domRef}
        initialHtml={initialHtml}
        dark={resolvedTheme === "dark"}
        title={title}
        saveLabel={saveLabel}
        placeholder={placeholder}
        inline={inline}
        disabled={disabled}
        readOnly={readOnly}
        metadata={metadata}
        mentions={members
          .filter((member) => member.isActive)
          .map((member) => ({
            id: member.id,
            label: member.fullName || member.username,
          }))}
        closeRequest={closeRequest}
        onDraft={onDraft}
        onReady={async (readOnly) => {
          readyRef.current = true;
          setReady(true);
          setDOMError(null);
          onReadyChange?.(true);
          onReadOnlyChange?.(readOnly);
        }}
        onRuntimeError={async () =>
          fail(
            "The editor stopped responding. Your saved draft is still on this device. Reload it to continue.",
          )
        }
        onFlushComplete={async (requestId, value, error) =>
          flushController.acknowledge(requestId, value, error)
        }
        onMetadataPress={async (key) => onMetadataPress?.(key)}
        onSave={async (value) => {
          try {
            await onSave?.(value);
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
        onClose={async () => onClose?.()}
        onDiscard={() =>
          new Promise<void>((resolve, reject) => {
            if (!onDiscard) {
              resolve();
              return;
            }
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
          containerStyle: { flex: 1, minWidth: 0, width: "100%" },
          style: { flex: 1, width: "100%", backgroundColor: "transparent" },
          contentInsetAdjustmentBehavior: "never",
          automaticallyAdjustContentInsets: false,
          automaticallyAdjustsScrollIndicatorInsets: false,
          keyboardDisplayRequiresUserAction: false,
          hideKeyboardAccessoryView: true,
          onError: () =>
            fail("Could not load the editor. Reload it to continue."),
          onContentProcessDidTerminate: () =>
            fail(
              "The editor was interrupted. Reload your saved draft to continue.",
            ),
          onRenderProcessGone: () =>
            fail(
              "The editor was interrupted. Reload your saved draft to continue.",
            ),
        }}
      />
    </View>
  );
};

export const RichTextEditor = (props: Props) => {
  const { resolvedTheme } = useTheme();
  const surfaceRef = useRef<RichTextEditorHandle>(null);
  return (
    <Modal
      visible
      animationType="slide"
      presentationStyle="fullScreen"
      onRequestClose={() => surfaceRef.current?.requestClose()}
    >
      <SafeAreaProvider>
        <SafeAreaView
          edges={["top", "right", "bottom", "left"]}
          style={{
            flex: 1,
            backgroundColor: themeColors[resolvedTheme].background,
          }}
        >
          <KeyboardAvoidingView
            behavior={Platform.OS === "ios" ? "padding" : undefined}
            style={{ flex: 1, minWidth: 0, width: "100%" }}
          >
            <RichTextEditorSurface ref={surfaceRef} {...props} />
          </KeyboardAvoidingView>
        </SafeAreaView>
      </SafeAreaProvider>
    </Modal>
  );
};
