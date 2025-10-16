import React from "react";
import { Row, Text } from "@/components/ui";
import { DetailedStory } from "@/modules/stories/types";
import RenderHtml from "react-native-render-html";
import { useWindowDimensions } from "react-native";
import { useTheme } from "@/hooks";
import { colors } from "@/constants";

export const Description = ({ story }: { story: DetailedStory }) => {
  const { width } = useWindowDimensions();
  const { resolvedTheme } = useTheme();
  if (!story.description) {
    return (
      <Row asContainer className="mb-6 mt-2">
        <Text color="muted" className="pl-0.5" fontWeight="medium">
          Enter description
        </Text>
      </Row>
    );
  }

  const source = {
    html: story.descriptionHTML || story.description,
  };
  return (
    <Row asContainer className="mb-5">
      <RenderHtml
        baseStyle={{
          fontSize: 16.5,
          fontWeight: 500,
          lineHeight: 24,
          opacity: 0.7,
          color: resolvedTheme === "dark" ? colors.white : colors.black,
        }}
        contentWidth={width - 24}
        source={source}
      />
    </Row>
  );
};
