import * as SecureStore from "expo-secure-store";

const storage = {
  async setItem(key: string, value: string) {
    if (process.env.EXPO_OS === "web") {
      window.localStorage.setItem(key, value);
      return;
    }

    await SecureStore.setItemAsync(key, value);
  },
  async getItem(key: string) {
    if (process.env.EXPO_OS === "web") {
      return window.localStorage.getItem(key);
    }

    return SecureStore.getItemAsync(key);
  },
  async removeItem(key: string) {
    if (process.env.EXPO_OS === "web") {
      window.localStorage.removeItem(key);
      return;
    }

    await SecureStore.deleteItemAsync(key);
  },
};

export async function saveWorkspace(workspace: string) {
  await storage.setItem("workspace", workspace);
}

export async function getWorkspace() {
  return storage.getItem("workspace");
}

export async function clearWorkspace() {
  await storage.removeItem("workspace");
}

export async function saveSessionFlag() {
  await storage.setItem("hasSession", "true");
}

export async function getSessionFlag() {
  return storage.getItem("hasSession");
}

export async function clearSessionFlag() {
  await storage.removeItem("hasSession");
}
