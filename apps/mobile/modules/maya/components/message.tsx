import type { UIMessage } from "ai";
import { memo } from "react";
import { Linking, StyleSheet, View } from "react-native";
import Markdown from "react-native-markdown-display";
import { Ionicons } from "@expo/vector-icons";
import { toast } from "sonner-native";
import { Text } from "@/components/ui";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks";
import { ApprovalCard } from "./approval-card";
import {
  approvalDetails,
  humanize,
  messageLinkURL,
  record,
  toolResultError,
} from "./message-model";

export const MayaMessage = memo(function MayaMessage({
  message,
  names,
  allowApproval,
  busy,
  onApprove,
}: {
  message: UIMessage;
  names: ReadonlyMap<string, string>;
  allowApproval: boolean;
  busy: boolean;
  onApprove: (id: string, approved: boolean) => void;
}) {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];
  const isUser = message.role === "user";
  const markdownStyles = {
    body: {
      color: theme.foreground,
      fontSize: 17,
      lineHeight: 26,
      fontWeight: "500" as const,
    },
    paragraph: { marginTop: 0, marginBottom: 10 },
    heading1: {
      fontSize: 23,
      lineHeight: 29,
      fontWeight: "700" as const,
      marginBottom: 10,
    },
    heading2: {
      fontSize: 20,
      lineHeight: 26,
      fontWeight: "700" as const,
      marginBottom: 8,
    },
    heading3: { fontSize: 17, lineHeight: 24, fontWeight: "700" as const },
    strong: { fontWeight: "700" as const },
    code_inline: {
      backgroundColor: theme.accent,
      color: theme.foreground,
      borderColor: theme.border,
    },
    fence: {
      backgroundColor: theme.surfaceMuted,
      color: theme.foreground,
      borderColor: theme.border,
      borderRadius: 12,
      padding: 12,
      fontSize: 14,
    },
    blockquote: {
      backgroundColor: theme.surfaceMuted,
      borderColor: theme.borderStrong,
    },
    link: { color: theme.foreground, textDecorationLine: "underline" as const },
    table: { borderColor: theme.border, borderRadius: 8 },
    tr: { borderColor: theme.border },
    th: { padding: 6 },
    td: { padding: 6 },
  };
  return (
    <View style={[styles.message, isUser && styles.user]}>
      <View
        style={[
          styles.body,
          isUser && {
            backgroundColor: theme.surfaceMuted,
            borderRadius: 22,
            paddingHorizontal: 16,
            paddingVertical: 12,
          },
        ]}
      >
        {message.parts.map((part, index) => {
          if (isUser && part.type === "file") {
            return (
              <View key={index} style={styles.attachment}>
                <Ionicons
                  name={
                    part.mediaType.startsWith("image/")
                      ? "image-outline"
                      : "document-text-outline"
                  }
                  size={20}
                  color={theme.textMuted}
                />
                <Text numberOfLines={1} style={{ flexShrink: 1 }}>
                  {part.filename ||
                    (part.mediaType.startsWith("image/")
                      ? "Image"
                      : "Document")}
                </Text>
              </View>
            );
          }
          if (part.type === "text") {
            if (!part.text.trim()) return null;
            return isUser ? (
              <Text key={index} selectable>
                {part.text}
              </Text>
            ) : (
              <Markdown
                key={index}
                style={markdownStyles}
                onLinkPress={(url) => {
                  const target = messageLinkURL(url);
                  if (!target) {
                    toast.error("Could not open this link", {
                      description:
                        "This link is not supported in the mobile app.",
                    });
                    return false;
                  }
                  void Linking.openURL(target).catch(() => {
                    toast.error("Could not open this link", {
                      description: "Please check the link and try again.",
                    });
                  });
                  return false;
                }}
                rules={{ image: () => null }}
              >
                {part.text}
              </Markdown>
            );
          }
          if (!part.type.startsWith("tool-") && part.type !== "dynamic-tool")
            return null;
          const tool = record(part);
          if (!tool) return null;
          const name =
            typeof tool.toolName === "string"
              ? tool.toolName
              : part.type.replace(/^tool-/, "");
          const approval = record(tool.approval);
          if (
            tool.state === "approval-requested" &&
            typeof approval?.id === "string"
          ) {
            if (!allowApproval)
              return (
                <Text key={index} color="muted" fontSize="sm">
                  This earlier proposal is no longer active.
                </Text>
              );
            return (
              <ApprovalCard
                key={index}
                title={humanize(name)}
                description="Review this change before Maya applies it."
                details={approvalDetails(tool.input, names)}
                destructive={/delete|remove/i.test(name)}
                busy={busy}
                onConfirm={() => onApprove(approval.id as string, true)}
                onCancel={() => onApprove(approval.id as string, false)}
              />
            );
          }
          if (tool.state === "output-denied")
            return (
              <Text key={index} color="muted" fontSize="sm">
                Change cancelled
              </Text>
            );
          if (tool.state === "output-error")
            return (
              <Text key={index} color="danger">
                {typeof tool.errorText === "string"
                  ? tool.errorText
                  : "This action could not be completed."}
              </Text>
            );
          const output = record(tool.output);
          const error = output ? toolResultError(output) : null;
          if (error)
            return (
              <Text key={index} fontSize="sm" color="danger">
                {error}
              </Text>
            );
          return null;
        })}
      </View>
    </View>
  );
});

const styles = StyleSheet.create({
  message: { paddingHorizontal: 20, paddingVertical: 12 },
  user: { alignItems: "flex-end", paddingLeft: 48 },
  body: { gap: 10, maxWidth: "100%" },
  attachment: { flexDirection: "row", alignItems: "center", gap: 8 },
});
