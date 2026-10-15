import type { Event } from "@tauri-apps/api/event";
import { native, errorMessage } from "./native";

// Tauri registers listeners asynchronously. Cleanup must also handle unmount before registration finishes.
export function subscribeNative<T>(
  name: string,
  listener: (event: Event<T>) => void,
  onError: (message: string) => void,
) {
  let disposed = false;
  let stop: (() => void) | undefined;
  void native
    .listen<T>(name, (event) => {
      if (!disposed) listener(event);
    })
    .then(
      (unlisten) => {
        if (disposed) unlisten();
        else stop = unlisten;
      },
      (error: unknown) => {
        if (!disposed) onError(errorMessage(error));
      },
    );
  return () => {
    disposed = true;
    stop?.();
  };
}
