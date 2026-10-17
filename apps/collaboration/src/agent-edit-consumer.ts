import { randomUUID } from "node:crypto";
import * as Y from "yjs";
import { AgentEditStore } from "./agent-edit-store";

/** Persist first, then use Hocuspocus's normal update path for sockets/Redis. */
export class AgentEditConsumer {
  private readonly instanceId = randomUUID();
  private readonly synchronized = new WeakMap<Y.Doc, string>();
  private timer: ReturnType<typeof setTimeout> | undefined;
  private active: Promise<void> | undefined;
  private stopped = true;
  private lastHeartbeat = 0;
  private roomOffset = 0;

  constructor(
    private readonly store: AgentEditStore,
    private readonly rooms: () => ReadonlyMap<string, Y.Doc>,
    private readonly reportError: () => void = () =>
      console.error(
        "Document agent edit processing is temporarily unavailable",
      ),
  ) {}

  start() {
    if (!this.stopped) return;
    this.stopped = false;
    this.schedule(0);
  }

  private schedule(delay: number) {
    this.timer = setTimeout(() => {
      this.active = this.tick()
        .catch(this.reportError)
        .finally(() => {
          this.active = undefined;
          if (!this.stopped) this.schedule(1_000);
        });
    }, delay);
    this.timer.unref();
  }

  async tick() {
    for (const edit of await this.store.pending()) {
      const committed = await this.store.apply(edit.id, edit.invocation_id);
      if (committed)
        this.broadcast(committed.name, committed.state, committed.revision);
    }
    // Every instance catches up its loaded rooms, including a process that did
    // not claim the command. This also recovers a crash after commit/before fanout.
    const rooms = [...this.rooms().keys()];
    if (this.roomOffset >= rooms.length) this.roomOffset = 0;
    const batch = rooms.slice(this.roomOffset, this.roomOffset + 10);
    this.roomOffset += batch.length;
    for (const current of await this.store.revisions(batch)) {
      const room = this.rooms().get(current.name);
      if (!room || this.synchronized.get(room) === current.revision) continue;
      const persisted = await this.store.committedState(current.name);
      if (persisted)
        this.broadcast(
          current.name,
          persisted.collaboration_state,
          persisted.revision,
        );
    }
    if (Date.now() - this.lastHeartbeat > 15_000) {
      await this.store.heartbeat(this.instanceId);
      this.lastHeartbeat = Date.now();
    }
  }

  private broadcast(name: string, state: Uint8Array, revision: string) {
    const room = this.rooms().get(name);
    if (!room) return;
    Y.applyUpdate(room, state, {
      source: "local",
      context: { agentEdit: true },
    });
    this.synchronized.set(room, revision);
  }

  async stop() {
    this.stopped = true;
    clearTimeout(this.timer);
    await this.active;
    await this.store.removeHeartbeat(this.instanceId);
  }
}
