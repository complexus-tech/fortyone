use super::*;
use crate::offline::{model::empty_workspace, store::Store};

fn with_store(test: impl FnOnce(&Store)) {
    let path =
        std::env::temp_dir().join(format!("fortyone-download-{}.sqlite", uuid::Uuid::new_v4()));
    let store = Store::open(&path).unwrap();
    store
        .save_account(&json!({"id":"account"}), &json!([{"id":"workspace"}]))
        .unwrap();
    store.set_setting("workspace", &json!("workspace")).unwrap();
    test(&store);
    drop(store);
    for suffix in ["", "-wal", "-shm"] {
        let _ = std::fs::remove_file(format!("{}{suffix}", path.display()));
    }
}

#[test]
fn cursor_requires_progress_and_does_not_claim_missing_pages() {
    assert_eq!(
        next_offset(&json!({"nextStoryOffset":50}), 0).unwrap(),
        Some(50)
    );
    assert_eq!(
        next_offset(&json!({"nextStoryOffset":null}), 50).unwrap(),
        None
    );
    for value in [
        json!({}),
        json!({"nextStoryOffset":50}),
        json!({"nextStoryOffset":"100"}),
    ] {
        assert!(next_offset(&value, 50).is_err());
    }
}

#[test]
fn workspace_or_account_switch_stops_previous_download_context() {
    with_store(|store| {
        assert!(context_matches(store, "account", "workspace"));
        store
            .set_setting("workspace", &json!("new-workspace"))
            .unwrap();
        assert!(!context_matches(store, "account", "workspace"));
        store.set_setting("workspace", &json!("workspace")).unwrap();
        store.set_setting("account", &json!("new-account")).unwrap();
        assert!(!context_matches(store, "account", "workspace"));
    });
}

#[test]
fn thread_completeness_is_explicit_and_partial_pages_do_not_replace_history() {
    with_store(|store| {
        let mut snapshot = empty_workspace("workspace");
        snapshot["stories"] = json!([{"id":"story"}]);
        snapshot["comments"] = json!([{"id":"old","storyId":"story"}]);
        snapshot["commentsCompleteFor"] = json!(["story"]);
        store.snapshot("account", "workspace", &snapshot).unwrap();
        snapshot["comments"] = json!([{"id":"first-page","storyId":"story"}]);
        snapshot["commentsCompleteFor"] = json!([]);
        store.snapshot("account", "workspace", &snapshot).unwrap();
        let saved = store.server_snapshot("account", "workspace").unwrap();
        assert_eq!(saved["commentsCompleteFor"], json!([]));
        assert_eq!(saved["comments"].as_array().unwrap().len(), 2);
        snapshot["stories"] = json!([]);
        snapshot["comments"] =
            json!([{"id":"first-page","storyId":"story"},{"id":"last-page","storyId":"story"}]);
        snapshot["commentsCompleteFor"] = json!(["story"]);
        store.snapshot("account", "workspace", &snapshot).unwrap();
        let saved = store.server_snapshot("account", "workspace").unwrap();
        assert_eq!(saved["commentsCompleteFor"], json!(["story"]));
        assert_eq!(saved["comments"], snapshot["comments"]);
        snapshot["comments"] = json!([]);
        snapshot["commentsCompleteFor"] = json!([]);
        snapshot["removed"] = json!({"stories":["story"]});
        store.snapshot("account", "workspace", &snapshot).unwrap();
        let saved = store.server_snapshot("account", "workspace").unwrap();
        assert_eq!(saved["commentsCompleteFor"], json!([]));
        assert_eq!(saved["comments"], json!([]));
    });
}
