import { Text, ScrollView } from "react-native";
import { SafeContainer } from "@/components/ui";
import { colors } from "@/constants";
import { useTheme } from "@/hooks";

export const NewStory = () => {
  const { resolvedTheme } = useTheme();
  return (
    <SafeContainer isFull>
      <ScrollView
        style={{
          paddingTop: 44,
          flex: 1,
          backgroundColor:
            resolvedTheme === "light" ? colors.white : colors.dark[200],
        }}
        contentContainerStyle={{ paddingBottom: 120 }}
      >
        <Text>New Story</Text>
      </ScrollView>
    </SafeContainer>
  );
};
