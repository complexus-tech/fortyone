import { Comments } from "./comments";
import { useAccountNative } from "../hooks/use-account-native";
import { lazy, Suspense, useState } from "react";
import { AttachmentIcon } from "icons/src/attachment";
import type {
  JsonRecord,
  OperationInput,
  Story,
  WorkspaceData,
} from "../types";
import { useDraft } from "../hooks/use-draft";
import { flushDrafts } from "../drafts";
import { errorMessage } from "../native";
import { storyReference, resourceBaseVersion } from "../model";
import { AttachmentList } from "./attachment-list";
import { Dialog, ErrorNotice, Spinner } from "./primitives";

const RichText = lazy(() =>
  import("./rich-text").then((module) => ({ default: module.RichText })),
);
type StoryDraft = JsonRecord & {
  resourceId: string;
  title: string;
  teamId: string;
  statusId: string;
  priority: string;
  description: string;
  descriptionHTML: string;
  baseVersion: string | null;
};

export function StoryEditor({
  story,
  data,
  save,
  acceptData,
  onClose,
}: {
  story: Story | null;
  data: WorkspaceData;
  save: (input: OperationInput) => Promise<void>;
  acceptData: (data: WorkspaceData) => void;
  onClose: () => void;
}) {
  const client = useAccountNative();
  const draftKey = story?.id ?? "new-story";
  const firstTeam = data.teams[0]?.id ?? "";
  const draft = useDraft<StoryDraft>(data.workspaceId, draftKey, "story", {
    resourceId: story?.id ?? crypto.randomUUID(),
    title: story?.title ?? "",
    teamId: story?.teamId ?? firstTeam,
    statusId:
      story?.statusId ??
      data.statuses.find((status) => status.teamId === firstTeam)?.id ??
      "",
    priority: story?.priority ?? "No Priority",
    description: story?.description ?? "",
    descriptionHTML: story?.descriptionHTML ?? story?.description ?? "",
    baseVersion: story
      ? resourceBaseVersion("story", story.id, story.updatedAt, data.pending)
      : null,
  });
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const fields = draft.value;
  const close = async () => {
    try {
      await flushDrafts();
      onClose();
    } catch (reason) {
      setError(errorMessage(reason));
    }
  };
  const submit = async () => {
    if (!fields.title.trim() || !fields.teamId || busy) return;
    setBusy(true);
    setError(null);
    try {
      const { resourceId, baseVersion, ...payload } = fields;
      if (story) delete (payload as JsonRecord).teamId;
      if (!payload.statusId) delete (payload as JsonRecord).statusId;
      await save({
        workspaceId: data.workspaceId,
        resourceId,
        baseVersion,
        kind: story ? "story.update" : "story.create",
        consumedDraft: { resourceId: draftKey, kind: "story" },
        payload,
      });
      onClose();
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
    }
  };
  return (
    <Dialog
      title={story ? storyReference(story, data) : "New story"}
      onClose={() => {
        void close();
      }}
      wide
    >
      <div className="editor-body">
        <ErrorNotice message={error ?? draft.error} />
        {!draft.ready ? (
          <p role="status">
            <Spinner />
            Recovering your saved work…
          </p>
        ) : (
          <>
            <label className="sr-only" htmlFor="story-title">
              Story title
            </label>
            <input
              id="story-title"
              className="title-input"
              placeholder="What needs to be done?"
              value={fields.title}
              readOnly={busy}
              onChange={(event) =>
                draft.update({ ...fields, title: event.target.value })
              }
              autoFocus
            />
            <div className="property-fields">
              <label>
                Team
                <select
                  value={fields.teamId}
                  disabled={busy || Boolean(story)}
                  onChange={(event) =>
                    draft.update({
                      ...fields,
                      teamId: event.target.value,
                      statusId:
                        data.statuses.find(
                          (status) => status.teamId === event.target.value,
                        )?.id ?? "",
                    })
                  }
                >
                  {data.teams.map((team) => (
                    <option value={team.id} key={team.id}>
                      {team.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Status
                <select
                  value={fields.statusId}
                  disabled={busy}
                  onChange={(event) =>
                    draft.update({ ...fields, statusId: event.target.value })
                  }
                >
                  <option value="">Default status</option>
                  {data.statuses
                    .filter((status) => status.teamId === fields.teamId)
                    .map((status) => (
                      <option value={status.id} key={status.id}>
                        {status.name}
                      </option>
                    ))}
                </select>
              </label>
              <label>
                Priority
                <select
                  value={fields.priority}
                  disabled={busy}
                  onChange={(event) =>
                    draft.update({ ...fields, priority: event.target.value })
                  }
                >
                  {["No Priority", "Low", "Medium", "High", "Urgent"].map(
                    (priority) => (
                      <option key={priority}>{priority}</option>
                    ),
                  )}
                </select>
              </label>
            </div>
            <Suspense fallback={<p>Loading editor…</p>}>
              <RichText
                mediaScope={{
                  workspaceId: data.workspaceId,
                  resourceId: fields.resourceId,
                  kind: "story",
                }}
                html={fields.descriptionHTML}
                readOnly={busy}
                onChange={(descriptionHTML, description) =>
                  draft.update({ ...fields, descriptionHTML, description })
                }
              />
            </Suspense>
            {story ? (
              <>
                <button
                  className="text-button attachment-button"
                  disabled={busy}
                  onClick={() => {
                    setBusy(true);
                    void client
                      .stageAttachment(data.workspaceId, story.id)
                      .then(acceptData)
                      .catch((reason: unknown) =>
                        setError(errorMessage(reason)),
                      )
                      .finally(() => setBusy(false));
                  }}
                >
                  <AttachmentIcon />
                  Add attachment
                </button>
                <AttachmentList
                  data={data}
                  resourceId={story.id}
                  kind="story"
                />
                <Comments storyId={story.id} data={data} save={save} />
              </>
            ) : null}
          </>
        )}
      </div>
      <footer className="dialog-footer">
        <span
          className={draft.error ? "save-status failed" : "save-status"}
          role="status"
        >
          {draft.status}
        </span>
        {draft.error ? (
          <button
            onClick={() =>
              draft.ready ? draft.update(fields) : draft.retryRead()
            }
          >
            {draft.ready ? "Retry local save" : "Retry draft recovery"}
          </button>
        ) : null}
        <button
          className="primary-button"
          disabled={
            !draft.ready || busy || !fields.title.trim() || !fields.teamId
          }
          onClick={() => {
            void submit();
          }}
        >
          {busy ? <Spinner /> : null}
          {story ? "Save changes" : "Create story"}
        </button>
      </footer>
    </Dialog>
  );
}
