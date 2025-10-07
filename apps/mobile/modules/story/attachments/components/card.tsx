import React from "react";
import { Pressable, Linking, Image } from "react-native";
import { Row, Text, Col } from "@/components/ui";
import type { StoryAttachment } from "@/types/attachment";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { useColorScheme } from "nativewind";

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
};

const getMimeType = (mimeType: string) => {
  if (mimeType.includes("image")) return "image";
  if (mimeType.includes("video")) return "video";
  if (mimeType.includes("pdf")) return "pdf";
  return "doc";
};

const AttachmentPreview = ({ attachment }: { attachment: StoryAttachment }) => {
  const { colorScheme } = useColorScheme();
  const mimeType = getMimeType(attachment.mimeType);
  if (mimeType === "image") {
    return (
      <Image
        source={{ uri: attachment.url }}
        className="size-12 rounded-lg"
        resizeMode="cover"
      />
    );
  }
  if (mimeType === "pdf") {
    return (
      <Row
        align="center"
        justify="center"
        className="size-12 rounded-lg bg-gray-100 dark:bg-dark-200"
      >
        <SymbolView
          size={28}
          name="text.document.fill"
          tintColor={
            colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
          }
        />
      </Row>
    );
  }
  return (
    <Row
      align="center"
      justify="center"
      className="size-12 rounded-lg bg-gray-100 dark:bg-dark-200"
    >
      <SymbolView
        size={28}
        name="document.fill"
        tintColor={
          colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
        }
      />
    </Row>
  );
};

export const AttachmentCard = ({
  attachment,
}: {
  attachment: StoryAttachment;
}) => {
  const handlePress = async () => {
    try {
      await Linking.openURL(attachment.url);
    } catch (error) {
      console.error("Failed to open attachment:", error);
    }
  };

  return (
    <Pressable
      className="active:bg-gray-50 dark:active:bg-dark-300 px-4 py-3"
      onPress={handlePress}
    >
      <Row align="center" gap={3}>
        <AttachmentPreview attachment={attachment} />
        <Col className="flex-1">
          <Text fontWeight="medium" numberOfLines={1}>
            {attachment.filename}
          </Text>
          <Text fontSize="sm" color="muted">
            {formatFileSize(attachment.size)}
          </Text>
        </Col>
      </Row>
    </Pressable>
  );
};
