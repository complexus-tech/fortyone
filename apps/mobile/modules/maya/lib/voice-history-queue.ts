export type VoiceHistoryMessage = {
  id: string;
  role: "user" | "assistant";
  text: string;
  order?: number;
};

type Options = {
  assertCurrent: () => void;
  persist: (
    messages: Pick<VoiceHistoryMessage, "id" | "role" | "text">[],
  ) => Promise<void>;
};

/** Retains exact IDs after uncertain saves; only an explicit retry can resend them. */
export const createVoiceHistoryQueue = (options: Options) => {
  let pending: VoiceHistoryMessage[] = [];
  let previousUser: VoiceHistoryMessage | undefined;
  let failure: unknown;
  let saving: Promise<void> | null = null;
  const known = new Map<string, VoiceHistoryMessage>();
  const sessionOrder = new Map<string, number>();
  const session = (message: VoiceHistoryMessage) =>
    message.id.match(/^voice-([0-9a-f-]{36})-/i)?.[1] ?? "default";
  const wire = ({ id, role, text }: VoiceHistoryMessage) => ({
    id,
    role,
    text,
  });
  const flush = (explicitRetry = true): Promise<void> => {
    options.assertCurrent();
    if (saving) return saving.then(() => flush(explicitRetry));
    if (failure && !explicitRetry) return Promise.reject(failure);
    failure = undefined;
    const operation = async () => {
      while (pending.length) {
        options.assertCurrent();
        // A provider greeting is not a user conversation yet.
        const anchor =
          previousUser && session(previousUser) === session(pending[0])
            ? previousUser
            : undefined;
        if (!anchor && !pending.some((message) => message.role === "user")) {
          // An explicit flush ends the call; a greeting needs no fabricated user.
          if (explicitRetry) pending = [];
          return;
        }
        const snapshot: VoiceHistoryMessage[] = [];
        if (pending[0].role === "assistant" && anchor) snapshot.push(anchor);
        for (const message of pending) {
          if (snapshot.length === 40) break;
          const next = [...snapshot, message];
          if (
            new TextEncoder().encode(JSON.stringify(next.map(wire)))
              .byteLength > 120_000
          )
            break;
          snapshot.push(message);
        }
        const anchored = snapshot[0] === anchor ? 1 : 0;
        const remaining = pending[snapshot.length - anchored];
        if (remaining?.role === "assistant") {
          // Persist a user's complete set of assistant followups in one write.
          const lastUser = snapshot.findLastIndex(
            (message) => message.role === "user",
          );
          snapshot.splice(Math.max(lastUser, 0));
        }
        if (!snapshot.some((message) => message.role === "user")) {
          throw new Error(
            "This voice turn exceeds the conversation history limit. Keep this conversation open; it has not been saved.",
          );
        }
        await options.persist(snapshot.map(wire));
        options.assertCurrent();
        const saved = new Set(snapshot.map((message) => message.id));
        pending = pending.filter((message) => !saved.has(message.id));
        previousUser =
          snapshot.findLast((message) => message.role === "user") ??
          previousUser;
      }
    };
    const result = Promise.resolve()
      .then(operation)
      .catch((error: unknown) => {
        failure = error;
        throw error;
      });
    saving = result;
    return result.finally(() => {
      if (saving === result) saving = null;
    });
  };
  return {
    add: (message: VoiceHistoryMessage) => {
      options.assertCurrent();
      if (
        !message.id ||
        message.id.length > 128 ||
        !message.text.trim() ||
        message.text.length > 16_000
      ) {
        throw new Error(
          "This voice message is too long to save. End the conversation before continuing.",
        );
      }
      const old = known.get(message.id);
      if (old) {
        if (old.role !== message.role || old.text !== message.text)
          throw new Error("A voice message changed after it was finalized.");
        return;
      }
      const key = session(message);
      if (!sessionOrder.has(key)) sessionOrder.set(key, sessionOrder.size);
      known.set(message.id, message);
      pending.push(message);
      pending.sort(
        (a, b) =>
          sessionOrder.get(session(a))! - sessionOrder.get(session(b))! ||
          (a.order ?? Number.MAX_SAFE_INTEGER) -
            (b.order ?? Number.MAX_SAFE_INTEGER),
      );
    },
    flush,
    get hasPending() {
      return (
        pending.some((message) => message.role === "user") ||
        (!!previousUser &&
          pending.some(
            (message) => session(message) === session(previousUser!),
          ))
      );
    },
  };
};
