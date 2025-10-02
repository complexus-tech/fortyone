import * as SecureStore from "expo-secure-store";

export async function saveAccessToken(token: string) {
  await SecureStore.setItemAsync("accessToken", token, {
    keychainService: "fortyone.auth",
  });
}

export async function getAccessToken() {
  return await SecureStore.getItemAsync("accessToken");
}

export async function clearAccessToken() {
  await SecureStore.deleteItemAsync("accessToken");
}

export async function saveWorkspace(workspace: string) {
  await SecureStore.setItemAsync("workspace", workspace);
}

export async function getWorkspace() {
  return await SecureStore.getItemAsync("workspace");
}

export async function clearWorkspace() {
  await SecureStore.deleteItemAsync("workspace");
}
