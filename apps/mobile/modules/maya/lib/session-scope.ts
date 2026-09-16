export type MayaSessionScope = {
  userId: string;
  workspace: string;
  sessionEpoch: number;
};

export type MayaAuthState = {
  userId: string | null;
  workspace: string | null;
  sessionEpoch: number;
  isAuthenticated: boolean;
  isLoading: boolean;
};

export const matchesMayaSession = (
  scope: MayaSessionScope,
  state: MayaAuthState,
) =>
  state.isAuthenticated &&
  !state.isLoading &&
  state.userId === scope.userId &&
  state.workspace === scope.workspace &&
  state.sessionEpoch === scope.sessionEpoch;

/** A disposed controller can never become active again, even on the same account. */
export const createMayaSessionGuard = (
  scope: MayaSessionScope,
  getState: () => MayaAuthState,
  providerIsActive: () => boolean = () => true,
) => {
  let active = true;
  let attachmentVersion = 0;
  const isCurrent = () =>
    active && providerIsActive() && matchesMayaSession(scope, getState());
  return {
    isCurrent,
    dispose: () => {
      active = false;
    },
    // React Strict Mode can detach and immediately reattach the same controller.
    retain: () => {
      const version = ++attachmentVersion;
      return () =>
        queueMicrotask(() => {
          if (version === attachmentVersion) active = false;
        });
    },
    assertCurrent: () => {
      if (!isCurrent()) {
        throw new Error("Your session changed. Open Maya again to continue.");
      }
    },
  };
};
