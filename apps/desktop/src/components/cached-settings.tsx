import { useAccountNative } from "../hooks/use-account-native";
import { useDeferredValue, useEffect, useState } from "react";
import type { JsonRecord, WorkspaceSettings } from "../types";
import { errorMessage } from "../native";
import { formatTime } from "../model";
import { subscribeNative } from "../events";
import { ErrorNotice, Spinner } from "./primitives";

const tabs = [
  "Profile",
  "Workspace",
  "Teams",
  "Members",
  "Preferences",
] as const;
type Tab = (typeof tabs)[number];
const paths: Record<Tab, string> = {
  Profile: "/settings/account",
  Workspace: "/settings",
  Teams: "/settings/workspace/teams",
  Members: "/settings/workspace/members",
  Preferences: "/settings/account/preferences",
};
function displayValue(value: unknown) {
  return typeof value === "boolean"
    ? value
      ? "On"
      : "Off"
    : Array.isArray(value)
      ? value.join(", ")
      : value == null || value === ""
        ? "Not set"
        : String(value);
}
function Details({
  value,
  fields,
}: {
  value: JsonRecord;
  fields: Array<[string, string]>;
}) {
  return (
    <dl className="cached-details">
      {fields
        .filter(([key]) => key in value)
        .map(([key, label]) => (
          <div key={key}>
            <dt>{label}</dt>
            <dd>
              {key === "workingStartMinute" || key === "workingEndMinute"
                ? formatMinute(value[key])
                : displayValue(value[key])}
            </dd>
          </div>
        ))}
    </dl>
  );
}
export function CachedSettings({ workspaceId }: { workspaceId: string }) {
  const client = useAccountNative();
  const [data, setData] = useState<WorkspaceSettings | null>(null);
  const [tab, setTab] = useState<Tab>("Profile");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [search, setSearch] = useState("");
  const query = useDeferredValue(search).toLocaleLowerCase();
  useEffect(() => {
    let active = true;
    let readVersion = 0;
    const readCached = async () => {
      const version = ++readVersion;
      try {
        const cached = await client.readSettings(workspaceId);
        if (active && version === readVersion) {
          setData(cached);
          setError(null);
        }
      } catch (reason) {
        if (active && version === readVersion) {
          setData(null);
          setError(errorMessage(reason));
        }
      } finally {
        if (active && version === readVersion) setLoading(false);
      }
    };
    const stop = subscribeNative(
      "desktop-state-changed",
      () => {
        void readCached();
      },
      setError,
    );
    const load = async () => {
      await readCached();
      if (!navigator.onLine || !active) return;
      setRefreshing(true);
      try {
        await client.syncSettings(workspaceId);
        if (active) await readCached();
      } catch (reason) {
        // Re-read native authorization after an online denial; a warmup may
        // also have populated the cache while this foreground request waited.
        if (active) await readCached();
        if (active) setError(errorMessage(reason));
      } finally {
        if (active) setRefreshing(false);
      }
    };
    void load();
    return () => {
      active = false;
      stop();
    };
  }, [client, workspaceId]);
  const open = async () => {
    try {
      await client.openWorkspacePage(workspaceId, paths[tab]);
    } catch (reason) {
      setError(errorMessage(reason));
    }
  };
  const rows =
    tab === "Teams" ? data?.teams : tab === "Members" ? data?.members : null;
  return (
    <section className="settings-group cached-settings">
      <h2>Account & workspace</h2>
      <div className="cached-settings-toolbar">
        <div
          className="settings-tabs"
          role="tablist"
          aria-label="Cached settings"
        >
          {tabs.map((item) => (
            <button
              key={item}
              role="tab"
              aria-selected={tab === item}
              onClick={() => setTab(item)}
            >
              {item}
            </button>
          ))}
        </div>
        {refreshing ? <Spinner /> : null}
      </div>
      <ErrorNotice message={error} />
      {loading ? (
        <p role="status">
          <Spinner />
          Opening downloaded settings…
        </p>
      ) : !data?.downloadedAt ? (
        <p className="footnote">
          Connect once to download account and workspace settings. They will
          stay available on this device.
        </p>
      ) : (
        <div role="tabpanel" aria-label={tab}>
          {tab === "Profile" ? (
            <Details
              value={data.profile}
              fields={[
                ["fullName", "Name"],
                ["email", "Email"],
                ["username", "Username"],
                ["timezone", "Time zone"],
                ["workingDays", "Working days"],
                ["workingStartMinute", "Day starts"],
                ["workingEndMinute", "Day ends"],
              ]}
            />
          ) : tab === "Workspace" ? (
            <Details
              value={data.workspace}
              fields={[
                ["name", "Workspace"],
                ["slug", "Address"],
                ["userRole", "Your role"],
                ["storyTerm", "Stories"],
                ["sprintTerm", "Sprints"],
                ["objectiveTerm", "Objectives"],
                ["workingDays", "Working days"],
                ["workingStartMinute", "Day starts"],
                ["workingEndMinute", "Day ends"],
              ]}
            />
          ) : tab === "Preferences" ? (
            <Details
              value={data.preferences}
              fields={Object.keys(data.preferences)
                .filter((key) => typeof data.preferences[key] !== "object")
                .map((key) => [
                  key,
                  key
                    .replace(/([A-Z])/g, " $1")
                    .replace(/^./, (letter) => letter.toUpperCase()),
                ])}
            />
          ) : (
            <>
              <input
                className="settings-search"
                aria-label={`Search ${tab.toLowerCase()}`}
                placeholder={`Search ${tab.toLowerCase()}…`}
                value={search}
                onChange={(event) => setSearch(event.target.value)}
              />
              <div className="cached-people">
                {rows
                  ?.filter((row) =>
                    `${row.name ?? row.fullName ?? ""} ${row.email ?? ""} ${row.code ?? ""}`
                      .toLocaleLowerCase()
                      .includes(query),
                  )
                  .map((row) => (
                    <div key={String(row.id)}>
                      <strong>
                        {displayValue(row.name ?? row.fullName ?? row.username)}
                      </strong>
                      <span>
                        {displayValue(tab === "Teams" ? row.code : row.email)}
                      </span>
                      {tab === "Members" ? (
                        <small>{displayValue(row.role)}</small>
                      ) : null}
                    </div>
                  ))}
              </div>
            </>
          )}
        </div>
      )}
      <div className="cached-settings-footer">
        <p className="footnote">
          {data?.downloadedAt
            ? `Downloaded ${formatTime(data.downloadedAt)}. `
            : ""}
          Settings changes require a connection so permissions and shared
          preferences stay consistent.
        </p>
        <button
          disabled={!navigator.onLine}
          onClick={() => {
            void open();
          }}
        >
          Edit in browser
        </button>
      </div>
    </section>
  );
}

function formatMinute(value: unknown) {
  if (typeof value !== "number") return "Not set";
  return `${String(Math.floor(value / 60)).padStart(2, "0")}:${String(value % 60).padStart(2, "0")}`;
}
