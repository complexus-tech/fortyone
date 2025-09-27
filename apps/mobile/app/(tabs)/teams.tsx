import React from "react";
import { View, StyleSheet } from "react-native";
import { Header } from "../../components/shared/Header";

export default function Teams() {
  const handleMenuPress = () => {
    console.log("Menu pressed");
    // Show menu options
  };

  return (
    <View style={styles.container}>
      <Header title="Teams" onSettingsPress={handleMenuPress} />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#FFFFFF",
  },
});
