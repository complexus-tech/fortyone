import { ScrollView, StyleSheet } from "react-native";
import { SafeContainer } from "@/components/ui";
import { Form } from "./components/form";

export const Settings = () => (
  <SafeContainer isFull edges={["bottom"]}>
    <ScrollView
      showsVerticalScrollIndicator={false}
      contentInsetAdjustmentBehavior="automatic"
      contentContainerStyle={styles.content}
    >
      <Form />
    </ScrollView>
  </SafeContainer>
);

const styles = StyleSheet.create({
  content: { paddingHorizontal: 16, paddingTop: 20, paddingBottom: 40 },
});
