/* global afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { act, renderHook } from "@testing-library/react";
import { hydrateRoot } from "react-dom/client";
import { renderToString } from "react-dom/server";
import {
  DEFAULT_CHAT_ALERT_PREFERENCES,
  readChatAlertPreferences,
  saveChatAlertPreferences,
} from "./preferences";
import { useChatAlertPreferences } from "./use-alert-preferences";

afterEach(() => {
  localStorage.clear();
});

describe("shared chat alert preferences", () => {
  it("hydrates defaults before showing saved quiet hours", async () => {
    const Preferences = () => {
      const { preferences } = useChatAlertPreferences("me", "acme");
      return (
        <div>
          <span>
            {preferences.quietHours ? "Quiet hours on" : "Quiet hours off"}
          </span>
          {preferences.quietHours ? (
            <span>{preferences.quietStart}</span>
          ) : null}
        </div>
      );
    };
    const container = document.createElement("div");
    container.innerHTML = renderToString(<Preferences />);
    expect(container.textContent).toBe("Quiet hours off");
    saveChatAlertPreferences("me", "acme", {
      ...DEFAULT_CHAT_ALERT_PREFERENCES,
      quietHours: true,
      quietStart: "21:30",
    });
    const onRecoverableError = jest.fn();
    let root: ReturnType<typeof hydrateRoot>;
    await act(async () => {
      root = hydrateRoot(container, <Preferences />, { onRecoverableError });
    });
    expect(onRecoverableError).not.toHaveBeenCalled();
    expect(container.textContent).toBe("Quiet hours on21:30");
    act(() => {
      root.unmount();
    });
  });

  it("updates subscribers while preserving the latest unrelated preferences", () => {
    const first = renderHook(() => useChatAlertPreferences("me", "acme"));
    const second = renderHook(() => useChatAlertPreferences("me", "acme"));

    act(() => {
      first.result.current.updatePreferences({ sound: false });
      second.result.current.updatePreferences({ quietHours: true });
    });

    expect(first.result.current.preferences).toMatchObject({
      sound: false,
      quietHours: true,
    });
    expect(second.result.current.preferences).toEqual(
      first.result.current.preferences,
    );
  });

  it("keeps account and workspace preferences isolated when the scope changes", () => {
    saveChatAlertPreferences("me", "acme", {
      ...DEFAULT_CHAT_ALERT_PREFERENCES,
      sound: false,
    });
    const view = renderHook(
      ({ userId, workspace }) => useChatAlertPreferences(userId, workspace),
      { initialProps: { userId: "me", workspace: "acme" } },
    );
    expect(view.result.current.preferences.sound).toBe(false);
    view.rerender({ userId: "other", workspace: "acme" });
    expect(view.result.current.preferences.sound).toBe(true);
    view.rerender({ userId: "me", workspace: "different" });
    expect(view.result.current.preferences.sound).toBe(true);
    act(() => {
      view.result.current.updatePreferences({ desktop: false });
    });
    expect(readChatAlertPreferences("me", "acme").desktop).toBe(true);
    expect(readChatAlertPreferences("me", "different").desktop).toBe(false);
  });

  it("rejects writes before the authenticated scope is available", () => {
    const { result } = renderHook(() => useChatAlertPreferences("", "acme"));
    expect(() => {
      result.current.updatePreferences({ sound: false });
    }).toThrow("Sign in");
    expect(localStorage.length).toBe(0);
  });
});
