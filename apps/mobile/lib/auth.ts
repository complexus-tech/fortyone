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
