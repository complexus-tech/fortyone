import { Text, View, StyleSheet, Pressable, Button } from "react-native";

export default function Index() {
  return (
    <View style={styles.container}>
      <Text style={styles.text}>Home</Text>
      <Pressable
        onPress={() => {
          alert("Hello");
        }}
        style={styles.button}
      >
        <Text style={styles.buttonText}>Press me</Text>
      </Pressable>

      <Button title="Press me" />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  text: {
    fontSize: 30,
    fontWeight: "bold",
  },
  button: {
    backgroundColor: "#ffffff",
    padding: 10,
    borderRadius: 5,
  },
  buttonText: {
    color: "#08090a",
  },
});
