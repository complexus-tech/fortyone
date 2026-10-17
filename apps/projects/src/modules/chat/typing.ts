"use client";

import { useCallback, useSyncExternalStore } from "react";

export type ChatTypingEvent = {
  workspaceId: string;
  conversationId: string;
  actorId: string;
  fullName: string;
  threadRootId: string | null;
  isTyping: boolean;
  expiresAt: string;
};
export type ChatTypingPerson = { id: string; fullName: string };
type Presence = {
  person: ChatTypingPerson;
  version: number;
  expires: number;
  retainUntil: number;
  active: boolean;
};
type Scope = {
  workspaceId: string;
  people: Map<string, Presence>;
  snapshot: ChatTypingPerson[];
};
const EMPTY: ChatTypingPerson[] = [];
const scopes = new Map<string, Scope>();
const listeners = new Map<string, Set<() => void>>();
let expiryTimer: ReturnType<typeof setTimeout> | undefined;
const keyFor = (
  workspaceId: string,
  conversationId: string,
  threadRootId?: string | null,
) => JSON.stringify([workspaceId, conversationId, threadRootId ?? null]);
const publish = (key: string, scope: Scope) => {
  const next = Array.from(scope.people.values())
    .filter((person) => person.active)
    .map((person) => person.person)
    .sort((a, b) => a.id.localeCompare(b.id));
  if (
    next.length === scope.snapshot.length &&
    next.every(
      (person, index) =>
        person.id === scope.snapshot[index].id &&
        person.fullName === scope.snapshot[index].fullName,
    )
  )
    return;
  scope.snapshot = next;
  listeners.get(key)?.forEach((listener) => {
    listener();
  });
};
const scheduleExpiry = () => {
  clearTimeout(expiryTimer);
  let next = Infinity;
  for (const scope of scopes.values())
    for (const person of scope.people.values())
      next = Math.min(
        next,
        person.active ? person.expires : person.retainUntil,
      );
  if (!Number.isFinite(next)) return;
  expiryTimer = setTimeout(
    () => {
      const now = Date.now();
      for (const [key, scope] of scopes) {
        for (const [id, person] of scope.people) {
          if (person.retainUntil <= now) scope.people.delete(id);
          else if (person.expires <= now) person.active = false;
        }
        publish(key, scope);
        if (!scope.people.size) scopes.delete(key);
      }
      scheduleExpiry();
    },
    Math.max(0, next - Date.now()),
  );
};

/** Server-derived names, recipient/workspace authorization checked by the SSE bridge. */
export const applyChatTypingEvent = (event: ChatTypingEvent) => {
  const now = Date.now();
  const version = Date.parse(event.expiresAt);
  if (!Number.isFinite(version) || version <= now) return;
  const key = keyFor(
    event.workspaceId,
    event.conversationId,
    event.threadRootId,
  );
  const scope = scopes.get(key) ?? {
    workspaceId: event.workspaceId,
    people: new Map<string, Presence>(),
    snapshot: EMPTY,
  };
  const previous = scope.people.get(event.actorId);
  if (
    previous &&
    (version < previous.version ||
      (version === previous.version && !previous.active))
  )
    return;
  scope.people.set(event.actorId, {
    person: { id: event.actorId, fullName: event.fullName },
    version,
    expires: Math.min(version, now + 6000),
    retainUntil: now + 12_000,
    active: event.isTyping,
  });
  scopes.set(key, scope);
  publish(key, scope);
  scheduleExpiry();
};

export const clearChatTyping = (workspaceId: string) => {
  for (const [key, scope] of scopes)
    if (scope.workspaceId === workspaceId) {
      scopes.delete(key);
      listeners.get(key)?.forEach((listener) => {
        listener();
      });
    }
  scheduleExpiry();
};

const serverSnapshot = () => EMPTY;
export const useChatTyping = (
  workspaceId: string,
  conversationId: string,
  threadRootId?: string | null,
) => {
  const key = keyFor(workspaceId, conversationId, threadRootId);
  const subscribe = useCallback(
    (listener: () => void) => {
      const current = listeners.get(key) ?? new Set<() => void>();
      current.add(listener);
      listeners.set(key, current);
      return () => {
        current.delete(listener);
        if (!current.size) listeners.delete(key);
      };
    },
    [key],
  );
  const snapshot = useCallback(() => scopes.get(key)?.snapshot ?? EMPTY, [key]);
  return useSyncExternalStore(subscribe, snapshot, serverSnapshot);
};
