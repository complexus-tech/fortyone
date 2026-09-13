import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import { useAuthStore } from "@/store/auth";
import { getDraftKey } from "./draft-storage";
import { mobileDraftRepository as repository } from "./draft-store";

export const useDraft = <T>(
  documentId: string,
  initialValue: T,
  isValid: (value: unknown) => value is T,
) => {
  const userId = useAuthStore((state) => state.userId);
  const workspace = useAuthStore((state) => state.workspace);
  const sessionEpoch = useAuthStore((state) => state.sessionEpoch);
  const key =
    userId && workspace ? getDraftKey(userId, workspace, documentId) : null;
  const initial = useRef(initialValue);
  const validator = useRef(isValid);
  const [value, setValue] = useState(initialValue);
  const [loadedKey, setLoadedKey] = useState<string | null>(null);
  const ready = key !== null && loadedKey === key;
  const [error, setError] = useState<string | null>(null);
  const valueRef = useRef(value);
  useLayoutEffect(() => {
    valueRef.current = value;
  }, [value]);

  useEffect(() => {
    let active = true;
    if (!key) return;
    void repository
      .read(key, validator.current)
      .then((stored) => {
        if (!active) return;
        setValue(stored ?? initial.current);
        setError(null);
        setLoadedKey(key);
      })
      .catch((cause: unknown) => {
        if (!active) return;
        setError(
          cause instanceof Error
            ? cause.message
            : "Could not restore your draft.",
        );
        // Do not overwrite an unreadable draft with an empty form.
      });
    return () => {
      active = false;
    };
  }, [key]);

  const persist = useCallback(
    async (next: T) => {
      const session = useAuthStore.getState();
      if (
        !key ||
        !session.isAuthenticated ||
        session.isLoading ||
        session.userId !== userId ||
        session.workspace !== workspace ||
        session.sessionEpoch !== sessionEpoch
      ) {
        throw new Error("Your session changed. Reopen the editor to continue.");
      }
      valueRef.current = next;
      setValue(next);
      try {
        await repository.write(key, next);
        setError(null);
      } catch (cause) {
        setError(
          "Your changes are on screen, but could not be saved on this device. Keep this screen open and try saving again.",
        );
        throw cause;
      }
    },
    [key, userId, workspace, sessionEpoch],
  );

  const update = useCallback(
    (next: T) => {
      void persist(next).catch(() => undefined);
    },
    [persist],
  );

  const clear = useCallback(async () => {
    if (key) await repository.remove(key);
  }, [key]);

  const reset = useCallback(async () => {
    if (key) await repository.remove(key);
    valueRef.current = initial.current;
    setValue(initial.current);
    setError(null);
    setLoadedKey(key);
  }, [key]);

  return { value, valueRef, update, persist, ready, error, clear, reset };
};
