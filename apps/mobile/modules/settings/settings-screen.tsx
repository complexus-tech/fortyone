import { ScrollView, View } from "react-native";
import { SafeContainer } from "@/components/ui";
import { Form } from "./components/form";
import { Header } from "./components/header";

export const Settings = () => (
  <SafeContainer isFull edges={["bottom"]}>
    <View style={{ paddingTop: 20 }}>
      <Header />
    </View>
    <ScrollView
      showsVerticalScrollIndicator={false}
      contentInsetAdjustmentBehavior="automatic"
      contentContainerStyle={{ paddingBottom: 40 }}
    >
      <Form />
    </ScrollView>
  </SafeContainer>
);
