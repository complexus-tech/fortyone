import { useDeferredValue, useMemo, useState } from "react";
import { SearchIcon } from "icons/src/search";
import { KanbanIcon } from "icons/src/kanban";
import { ListIcon } from "icons/src/list";
import { PlusIcon } from "icons/src/plus-icon";
import type { Story, WorkspaceData } from "../types";
import { findStories, storyReference } from "../model";
import { Empty } from "./primitives";

export function StoryList({
  data,
  userId,
  mine,
  onOpen,
  onCreate,
}: {
  data: WorkspaceData;
  userId: string;
  mine: boolean;
  onOpen: (story: Story) => void;
  onCreate: () => void;
}) {
  const [query, setQuery] = useState("");
  const [board, setBoard] = useState(false);
  const deferredQuery = useDeferredValue(query);
  const stories = useMemo(
    () => findStories(data, userId, deferredQuery, mine),
    [data, userId, deferredQuery, mine],
  );
  const groups = data.statuses
    .map((status) => ({
      ...status,
      stories: stories.filter((story) => story.statusId === status.id),
    }))
    .filter((group) => group.stories.length > 0);
  const ungrouped = stories.filter(
    (story) => !data.statuses.some((status) => status.id === story.statusId),
  );
  if (ungrouped.length)
    groups.push({
      id: "unassigned",
      name: "Unassigned",
      teamId: "",
      stories: ungrouped,
    });
  return (
    <section className="collection">
      <div className="page-heading">
        <div>
          <span className="eyebrow">Your workspace, wherever you are</span>
          <h1>{mine ? "My work" : "All stories"}</h1>
          <p>
            {mine
              ? "A clear view of what needs your attention."
              : "Everything your team is moving forward."}
          </p>
        </div>
        <button
          className="primary-button"
          onClick={onCreate}
          disabled={!data.teams.length}
        >
          <PlusIcon />
          New story
        </button>
      </div>
      <div className="collection-toolbar">
        <label className="search-field">
          <SearchIcon />
          <input
            id="desktop-search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search stories…"
            aria-label="Search stories"
          />
          <kbd>⌘ K</kbd>
        </label>
        <span className="item-count">{stories.length} stories</span>
        <div className="view-toggle">
          <button
            aria-label="List view"
            aria-pressed={!board}
            onClick={() => setBoard(false)}
          >
            <ListIcon />
          </button>
          <button
            aria-label="Board view"
            aria-pressed={board}
            onClick={() => setBoard(true)}
          >
            <KanbanIcon />
          </button>
        </div>
      </div>
      {!stories.length ? (
        <Empty
          title={
            query
              ? "No matching stories"
              : mine
                ? "A little room to focus"
                : "Your next idea starts here"
          }
        >
          {query
            ? "Try a different title or story reference."
            : mine
              ? "Stories assigned to you will appear here. Browse all stories to see your team’s work."
              : "Create a story. It will be saved on this device, even while you’re offline."}
        </Empty>
      ) : board ? (
        <div className="story-board">
          {groups.map((group) => (
            <section className="board-column" key={group.id}>
              <h2>
                <span
                  className="status-dot"
                  style={{ background: group.color }}
                />
                {group.name}
                <span>{group.stories.length}</span>
              </h2>
              {group.stories.map((story) => (
                <button
                  className="story-card"
                  key={story.id}
                  onClick={() => onOpen(story)}
                >
                  <span className="story-reference">
                    {storyReference(story, data)}
                  </span>
                  <strong>{story.title}</strong>
                  <small>{story.priority}</small>
                </button>
              ))}
            </section>
          ))}
        </div>
      ) : (
        <div className="story-list">
          {groups.map((group) => (
            <section key={group.id}>
              <h2>
                <span
                  className="status-dot"
                  style={{ background: group.color }}
                />
                {group.name}
                <span>{group.stories.length}</span>
              </h2>
              {group.stories.map((story) => (
                <button
                  className="story-row"
                  key={story.id}
                  onClick={() => onOpen(story)}
                >
                  <span className="story-reference">
                    {storyReference(story, data)}
                  </span>
                  <span className="story-title">{story.title}</span>
                  {data.pending.some(
                    (operation) => operation.resourceId === story.id,
                  ) ? (
                    <span className="pending-dot" title="Waiting to sync" />
                  ) : null}
                  <span className="priority-label">{story.priority}</span>
                </button>
              ))}
            </section>
          ))}
        </div>
      )}
    </section>
  );
}
