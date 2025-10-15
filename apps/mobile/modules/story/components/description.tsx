import React from "react";
import { Row, Text } from "@/components/ui";
import { DetailedStory } from "@/modules/stories/types";
import RenderHtml from "react-native-render-html";
import { useWindowDimensions } from "react-native";
import { useColorScheme } from "nativewind";
import { colors } from "@/constants";

export const Description = ({ story }: { story: DetailedStory }) => {
  const { width } = useWindowDimensions();
  const { colorScheme } = useColorScheme();
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
          fontSize: 16,
          fontWeight: 500,
          lineHeight: 23,
          opacity: 0.8,
          color: colorScheme === "dark" ? colors.white : colors.black,
        }}
        contentWidth={width}
        source={source}
      />
    </Row>
  );
};
