use super::{download, model::now, Runtime};
use serde_json::{json, Value};
use std::{
    collections::{HashMap, HashSet},
    sync::{Arc, Mutex},
    time::Duration,
};
use tauri::{AppHandle, Emitter, Listener};

#[derive(Default)]
struct Schedule {
    last_sync: u64,
    last_warm: u64,
    retry_at: u64,
    failures: u32,
}
#[derive(Debug, PartialEq)]
struct Work {
    synchronize: bool,
    warm: bool,
}
impl Schedule {
    fn failed(&mut self, time: u64) {
        self.failures = self.failures.saturating_add(1);
        self.retry_at =
            time.saturating_add((5u64.saturating_mul(1u64 << self.failures.min(6))).min(300));
    }
    fn plan(&self, time: u64, visible: bool, pending: bool, requested: bool) -> Option<Work> {
        if !requested && time < self.retry_at {
            return None;
        }
        let warm = visible
            && (requested || self.last_warm == 0 || time.saturating_sub(self.last_warm) >= 3600);
        let synchronize = pending
            || (visible && (self.last_sync == 0 || time.saturating_sub(self.last_sync) >= 300));
        (synchronize || warm).then_some(Work { synchronize, warm })
    }
    fn succeeded(&mut self, time: u64, warmed: bool) {
        if warmed {
            self.last_warm = time;
        }
        self.failures = 0;
        self.retry_at = 0;
    }
}

fn pending_workspaces(runtime: &Runtime, account: &str) -> Result<HashSet<String>, String> {
    let db = runtime.store.0.lock().map_err(|error| error.to_string())?;
    let mut query = db.prepare("SELECT DISTINCT workspace FROM operations WHERE account=? AND json_extract(operation,'$.status')='pending'").map_err(|error|error.to_string())?;
    let rows = query
        .query_map([account], |row| row.get::<_, String>(0))
        .map_err(|error| error.to_string())?;
    rows.collect::<Result<HashSet<_>, _>>()
        .map_err(|error| error.to_string())
}

fn reset_account_schedule(schedules: &mut HashMap<String, Schedule>, account: &str) {
    let prefix = format!("download:{account}:");
    schedules.retain(|key, _| !key.starts_with(&prefix));
}

pub fn start(app: &AppHandle, runtime: Arc<Runtime>) {
    let signal = Arc::new(tokio::sync::Notify::new());
    let changed = signal.clone();
    app.listen("desktop-state-changed", move |_| changed.notify_one());
    let restored_accounts = Arc::new(Mutex::new(HashSet::new()));
    let (restored, restored_runtime, wake) =
        (restored_accounts.clone(), runtime.clone(), signal.clone());
    app.listen("desktop-session-restored", move |event| {
        let Ok(value) = serde_json::from_str::<Value>(event.payload()) else {
            return;
        };
        let Some(account) = value["accountId"].as_str() else {
            return;
        };
        if restored_runtime.store.account().ok().as_deref() != Some(account) {
            return;
        }
        if let Ok(mut restored) = restored.lock() {
            restored.insert(account.to_owned());
        }
        wake.notify_one();
    });
    let forced = Arc::new(Mutex::new(HashSet::new()));
    let (requests, wake) = (forced.clone(), signal.clone());
    app.listen("desktop-refresh-request", move |event| {
        if let Ok(value) = serde_json::from_str::<Value>(event.payload()) {
            if let Some(workspace) = value["workspaceId"]
                .as_str()
                .filter(|id| uuid::Uuid::parse_str(id).is_ok())
            {
                if let Ok(mut requests) = requests.lock() {
                    requests.insert(workspace.to_owned());
                }
                wake.notify_one();
            }
        }
    });
    let app = app.clone();
    tauri::async_runtime::spawn(async move {
        let mut schedules = HashMap::<String, Schedule>::new();
        loop {
            let restored: Vec<String> = restored_accounts
                .lock()
                .map(|mut accounts| accounts.drain().collect())
                .unwrap_or_default();
            for account in restored {
                if runtime.store.account().ok().as_deref() == Some(&account) {
                    // Browser reconnection installs a new native session. Reset
                    // prior failures so full access checks resume immediately.
                    reset_account_schedule(&mut schedules, &account);
                }
            }
            if let Ok(bootstrap) = runtime.store.bootstrap() {
                if let Some(account) = bootstrap["user"]["id"].as_str() {
                    let active = bootstrap["activeWorkspaceId"].as_str();
                    let pending = pending_workspaces(&runtime, account).unwrap_or_default();
                    // Visible content has priority. Warmup yields to any pending
                    // writes so other workspaces do not wait for a full download.
                    let mut jobs: Vec<String> = active.map(str::to_owned).into_iter().collect();
                    jobs.extend(
                        pending
                            .iter()
                            .filter(|id| Some(id.as_str()) != active)
                            .cloned(),
                    );
                    for workspace in jobs {
                        if runtime.store.account().ok().as_deref() != Some(account) {
                            break;
                        }
                        let schedule = schedules
                            .entry(download::state_key(account, &workspace))
                            .or_default();
                        let requested = forced
                            .lock()
                            .map(|mut values| values.remove(&workspace))
                            .unwrap_or(false);
                        let visible = download::is_current(&runtime, account, &workspace);
                        if requested {
                            // This event follows a successful explicit Sync Now.
                            // Warm its cache without replaying the whole sync again.
                            schedule.last_sync = now();
                        }
                        let Some(work) =
                            schedule.plan(now(), visible, pending.contains(&workspace), requested)
                        else {
                            continue;
                        };
                        let result = async {
                            if work.synchronize {
                                super::sync::synchronize(&runtime, &workspace).await?;
                                schedule.last_sync = now();
                            }
                            if work.warm {
                                download::warm(&app, &runtime, &workspace).await.map(Some)
                            } else {
                                Ok(None)
                            }
                        }
                        .await;
                        match result {
                            Ok(outcome) => {
                                let yielded = outcome == Some(download::WarmOutcome::Yielded);
                                schedule.succeeded(
                                    now(),
                                    outcome == Some(download::WarmOutcome::Complete),
                                );
                                if !yielded {
                                    let _ =
                                        app.emit("desktop-connectivity", json!({"online":true}));
                                    let _ = app.emit("desktop-state-changed", ());
                                }
                            }
                            Err(error) => {
                                schedule.failed(now());
                                if download::is_current(&runtime, account, &workspace) {
                                    let _ = download::progress(
                                        &app,
                                        &runtime,
                                        account,
                                        &workspace,
                                        "paused",
                                        0,
                                        Some(&error),
                                    );
                                }
                                let unreachable = [
                                    "Could not reach",
                                    "Could not download",
                                    "interrupted",
                                    "You are offline",
                                ]
                                .iter()
                                .any(|text| error.contains(text));
                                if unreachable {
                                    let _ =
                                        app.emit("desktop-connectivity", json!({"online":false}));
                                }
                                let _ = app.emit("desktop-state-changed", ());
                            }
                        }
                    }
                }
            }
            tokio::select! {
                _ = signal.notified() => {},
                _ = tokio::time::sleep(Duration::from_secs(30)) => {},
            }
        }
    });
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn retries_back_off_without_overflow_and_success_resets_failures() {
        let mut schedule = Schedule::default();
        for _ in 0..100 {
            schedule.failed(100);
            assert!((110..=400).contains(&schedule.retry_at));
        }
        assert!(schedule.plan(200, true, true, false).is_none());
        assert!(schedule.plan(200, true, true, true).is_some());
        schedule.succeeded(450, true);
        assert_eq!(schedule.retry_at, 0);
        assert_eq!(schedule.failures, 0);
        assert_eq!(schedule.last_warm, 450);
    }
    #[test]
    fn visible_dependencies_warm_once_and_hidden_workspaces_only_sync_pending_work() {
        let mut schedule = Schedule::default();
        assert_eq!(
            schedule.plan(100, true, false, false),
            Some(Work {
                synchronize: true,
                warm: true
            })
        );
        assert_eq!(schedule.plan(100, false, false, false), None);
        assert_eq!(
            schedule.plan(100, false, true, false),
            Some(Work {
                synchronize: true,
                warm: false
            })
        );
        schedule.last_sync = 100;
        schedule.succeeded(100, true);
        assert_eq!(schedule.plan(200, true, false, false), None);
        assert_eq!(
            schedule.plan(400, true, false, false),
            Some(Work {
                synchronize: true,
                warm: false
            })
        );
        assert_eq!(
            schedule.plan(200, true, false, true),
            Some(Work {
                synchronize: false,
                warm: true
            })
        );
    }
    #[test]
    fn yield_does_not_mark_incomplete_download_ready() {
        let mut schedule = Schedule::default();
        schedule.last_sync = 100;
        schedule.succeeded(100, false);
        assert_eq!(schedule.last_warm, 0);
        assert_eq!(
            schedule.plan(101, true, false, false),
            Some(Work {
                synchronize: false,
                warm: true
            })
        );
    }

    #[test]
    fn browser_reconnect_resets_only_its_accounts_backoff_and_warmup() {
        let mut schedules = HashMap::new();
        let account_key = download::state_key("account", "workspace");
        let other_key = download::state_key("other", "workspace");
        for key in [&account_key, &other_key] {
            let mut schedule = Schedule::default();
            schedule.last_sync = 100;
            schedule.last_warm = 100;
            schedule.failed(100);
            schedules.insert(key.clone(), schedule);
        }
        reset_account_schedule(&mut schedules, "account");
        assert!(!schedules.contains_key(&account_key));
        assert!(schedules[&other_key].plan(101, true, true, false).is_none());
        assert_eq!(
            schedules
                .entry(account_key)
                .or_default()
                .plan(101, true, false, false),
            Some(Work {
                synchronize: true,
                warm: true
            })
        );
    }
}
