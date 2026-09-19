import AsyncStorage from "@react-native-async-storage/async-storage";

const PUSH_ENABLED_KEY = "@fortyone/push-notifications-enabled";

export const getPushEnabledPreference = async () => {
  const value = await AsyncStorage.getItem(PUSH_ENABLED_KEY);
  return value !== "false";
};

export const setPushEnabledPreference = (enabled: boolean) =>
  AsyncStorage.setItem(PUSH_ENABLED_KEY, String(enabled));
