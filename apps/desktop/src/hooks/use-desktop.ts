import { useCallback, useEffect, useRef, useState } from "react";
import type {
  Bootstrap,
  DownloadState,
  OperationInput,
  WorkspaceData,
} from "../types";
import { native, bindNativeAccount, errorMessage } from "../native";
import { subscribeNative } from "../events";
import { flushDrafts } from "../drafts";

export function useDesktop() {
  const [session, setSession] = useState<Bootstrap | null>(null);
  const [data, setData] = useState<WorkspaceData | null>(null);
  const [starting, setStarting] = useState(true);
  const [switchingId, setSwitchingId] = useState<string | null>(null);
  const [syncing, setSyncing] = useState(false);
  const [online, setOnline] = useState(navigator.onLine);
  const [error, setError] = useState<string | null>(null);
  const accountId = session?.user?.id ?? null;
  const activeAccount = useRef<string | null>(null);
  const activeWorkspace = useRef<string | null>(null);
  const request = useRef(0);
  const switching = useRef(false);
  const syncInFlight = useRef(false);

  const acceptData = useCallback(
    (next: WorkspaceData) => {
      if (
        accountId &&
        activeAccount.current === accountId &&
        next.workspaceId === activeWorkspace.current
      )
        setData(next);
    },
    [accountId],
  );

  const sync = useCallback(
    async (retryCredentials = false) => {
      const workspaceId = activeWorkspace.current;
      const account = activeAccount.current;
      if (!account || !workspaceId || syncInFlight.current || switching.current)
        return;
      syncInFlight.current = true;
      setSyncing(true);
      try {
        const next = await native
          .forAccount(account)
          .sync(workspaceId, retryCredentials);
        acceptData(next);
        if (
          activeAccount.current === account &&
          activeWorkspace.current === workspaceId
        )
          setError(null);
      } catch (reason) {
        if (
          activeAccount.current === account &&
          activeWorkspace.current === workspaceId
        )
          setError(errorMessage(reason));
      } finally {
        syncInFlight.current = false;
        setSyncing(false);
      }
    },
    [acceptData],
  );

  const reload = useCallback(async () => {
    const version = ++request.current;
    try {
      const next = await native.bootstrap();
      if (version !== request.current) return;
      if (activeAccount.current !== (next.user?.id ?? null)) setData(null);
      activeAccount.current = next.user?.id ?? null;
      bindNativeAccount(activeAccount.current);
      setSession(next);
      activeWorkspace.current = next.activeWorkspaceId;
      if (!next.user) setData(null);
      else if (next.activeWorkspaceId) {
        const cached = await native
          .forAccount(next.user.id)
          .readWorkspace(next.activeWorkspaceId);
        if (version === request.current) setData(cached);
      }
      if (version === request.current) setError(null);
    } catch (reason) {
      if (version === request.current) {
        setData(null);
        setError(errorMessage(reason));
      }
    } finally {
      if (version === request.current) setStarting(false);
    }
  }, []);

  const selectWorkspace = useCallback(async (workspaceId: string) => {
    if (switching.current) return false;
    if (workspaceId === activeWorkspace.current) return true;
    const account = activeAccount.current;
    switching.current = true;
    setSwitchingId(workspaceId);
    setError(null);
    ++request.current;
    try {
      await flushDrafts();
      if (!account || activeAccount.current !== account) return false;
      const next = await native
        .forAccount(account)
        .selectWorkspace(workspaceId);
      if (activeAccount.current !== account) return false;
      ++request.current;
      activeWorkspace.current = workspaceId;
      setData(next);
      setSession((current) =>
        current ? { ...current, activeWorkspaceId: workspaceId } : current,
      );
      return true;
    } catch (reason) {
      if (activeAccount.current === account) setError(errorMessage(reason));
      return false;
    } finally {
      switching.current = false;
      setSwitchingId(null);
    }
  }, []);

  const saveOperation = useCallback(
    async (input: OperationInput) => {
      await flushDrafts();
      if (!accountId || activeAccount.current !== accountId)
        throw new Error(
          "Your account changed. Reopen this item before saving.",
        );
      acceptData(await native.forAccount(accountId).saveOperation(input));
      if (navigator.onLine) void sync();
    },
    [accountId, acceptData, sync],
  );

  const logout = useCallback(
    async (discardLocalWork = false) => {
      await flushDrafts();
      if (activeAccount.current !== accountId)
        throw new Error(
          "Your account changed. Review the active account before signing out.",
        );
      if (!accountId) throw new Error("No account is signed in.");
      await native.forAccount(accountId).logout(discardLocalWork);
      ++request.current;
      activeAccount.current = null;
      bindNativeAccount(null);
      activeWorkspace.current = null;
      setData(null);
      setSession(null);
      await reload();
    },
    [accountId, reload],
  );

  useEffect(() => {
    void reload();
  }, [reload]);
  useEffect(() => {
    const update = () => {
      setOnline(navigator.onLine);
      if (navigator.onLine) void sync();
    };
    window.addEventListener("online", update);
    window.addEventListener("offline", update);
    return () => {
      window.removeEventListener("online", update);
      window.removeEventListener("offline", update);
    };
  }, [sync]);
  useEffect(
    () =>
      subscribeNative(
        "desktop-state-changed",
        () => {
          void reload();
        },
        setError,
      ),
    [reload],
  );
  useEffect(
    () =>
      subscribeNative<string>(
        "desktop-error",
        (event) => setError(event.payload),
        setError,
      ),
    [],
  );
  useEffect(
    () =>
      subscribeNative<{ online: boolean }>(
        "desktop-connectivity",
        (event) => setOnline(navigator.onLine && event.payload.online),
        setError,
      ),
    [],
  );
  useEffect(
    () =>
      subscribeNative<DownloadState>(
        "desktop-download-progress",
        (event) => {
          if (event.payload.workspaceId === activeWorkspace.current)
            setData((current) =>
              current ? { ...current, downloadState: event.payload } : current,
            );
        },
        setError,
      ),
    [],
  );

  return {
    session,
    data,
    starting,
    switchingId,
    syncing,
    online,
    error,
    setError,
    selectWorkspace,
    saveOperation,
    acceptData,
    sync,
    reload,
    logout,
  };
}
