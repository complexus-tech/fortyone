import { View } from "react-native";
import { Text } from "@/components/ui";

export const GroupHeader = ({
  title,
  count,
}: {
  title: string;
  count: number;
}) => (
  <View
    style={{
      flexDirection: "row",
      alignItems: "center",
      gap: 7,
      paddingHorizontal: 20,
      paddingTop: 18,
      paddingBottom: 7,
    }}
  >
    <Text
      accessibilityRole="header"
      color="muted"
      numberOfLines={1}
      style={{ flexShrink: 1, fontSize: 13, lineHeight: 18, fontWeight: "500" }}
    >
      {title}
    </Text>
    <Text
      color="muted"
      style={{ fontSize: 12, lineHeight: 18, fontWeight: "400" }}
    >
      {count}
    </Text>
  </View>
);
