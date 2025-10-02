import { Text, Row, Back } from "@/components/ui";
import { Pressable } from "react-native";
import { colors } from "@/constants";
import { SymbolView } from "expo-symbols";

export const Header = () => {
  return (
    <Row className="pb-3" justify="between" align="center">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        PRO-259
      </Text>
      <Pressable
        className="p-2 rounded-md"
        style={({ pressed }) => [
          pressed && { backgroundColor: colors.gray[50] },
        ]}
      >
        <SymbolView name="ellipsis" tintColor={colors.dark[50]} />
      </Pressable>
    </Row>
  );
};
