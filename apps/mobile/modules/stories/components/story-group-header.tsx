import React from "react";
import { View } from "react-native";
import { Text, Row } from "@/components/ui";
import { Dot } from "@/components/icons";

type StoryGroupHeaderProps = {
  title: string;
  color?: string;
};

export const StoryGroupHeader = ({ title, color }: StoryGroupHeaderProps) => {
  return (
    <View
      style={{
        paddingBottom: 8,
        paddingTop: 12,
        paddingHorizontal: 16,
      }}
    >
      <Row align="center" gap={2}>
        {color && <Dot color={color} size={12} />}
        <Text fontWeight="semibold" color="muted">
          {title}
        </Text>
      </Row>
    </View>
  );
};
