import React from "react";
import { Pressable, Linking } from "react-native";
import { Row, Text, Col } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants/colors";
import { useColorScheme } from "nativewind";
import type { StoryAttachment } from "@/types/attachment";

type AttachmentCardProps = {
  attachment: StoryAttachment;
};

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
};

const getFileIcon = (mimeType: string): string => {
  if (mimeType.includes("image")) return "ellipsis";
  if (mimeType.includes("video")) return "ellipsis";
  if (mimeType.includes("pdf")) return "ellipsis";
  return "ellipsis";
};

export const AttachmentCard = ({ attachment }: AttachmentCardProps) => {
  const { colorScheme } = useColorScheme();
  const iconName = getFileIcon(attachment.mimeType);

  const handlePress = async () => {
    try {
      await Linking.openURL(attachment.url);
    } catch (error) {
      console.error("Failed to open attachment:", error);
    }
  };

  return (
    <Pressable
      className="active:bg-gray-50 dark:active:bg-dark-300 rounded-lg p-4 mb-3"
      onPress={handlePress}
    >
      <Row align="center" gap={3}>
        <SymbolView name={iconName as any} size={24} />
        <Col className="flex-1">
          <Text fontSize="sm" fontWeight="medium" numberOfLines={2}>
            {attachment.filename}
          </Text>
          <Text fontSize="xs" color="muted">
            {formatFileSize(attachment.size)}
          </Text>
        </Col>
        <SymbolView
          name="arrow.up.right"
          size={16}
          tintColor={
            colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
          }
        />
      </Row>
    </Pressable>
  );
};
