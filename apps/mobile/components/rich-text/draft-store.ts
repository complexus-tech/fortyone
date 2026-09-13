import AsyncStorage from "@react-native-async-storage/async-storage";
import { createDraftRepository, DRAFT_STORAGE_PREFIX } from "./draft-storage";

export const mobileDraftRepository = createDraftRepository(AsyncStorage);

export const clearMobileDrafts = async () => {
  await mobileDraftRepository.flush();
  const keys = (await AsyncStorage.getAllKeys()).filter((key) =>
    key.startsWith(DRAFT_STORAGE_PREFIX),
  );
  if (keys.length) await AsyncStorage.multiRemove(keys);
};
