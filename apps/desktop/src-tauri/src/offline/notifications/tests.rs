use super::super::store::Store;
use super::codec::MAX_EVENT_BYTES;

use super::*;
const ACCOUNT: &str = "11111111-1111-4111-8111-111111111111";
const WORKSPACE: &str = "21111111-1111-4111-8111-111111111111";
const STORY: &str = "31111111-1111-4111-8111-111111111111";
fn notification() -> Value {
    json!({"id":"41111111-1111-4111-8111-111111111111","recipient_id":ACCOUNT,"workspace_id":WORKSPACE,"entity_type":"story","entity_id":STORY,"type":"story_update","title":"A story changed","created_at":"2026-10-14T12:00:00Z","read_at":null})
}

#[test]
fn sse_decodes_split_utf8_multiline_crlf_and_ignores_named_events() {
    let mut parser = SseDecoder::default();
    let bytes = "event: connected\r\ndata: {}\r\n\r\n:keep-alive\n\ndata: {\"title\":\"café\",\r\ndata: \"id\":1}\r\n\r\n".as_bytes();
    let mut events = Vec::new();
    for byte in bytes {
        events.extend(parser.push(&[*byte]).unwrap());
    }
    assert_eq!(events, vec!["{\"title\":\"café\",\n\"id\":1}"]);
}
#[test]
fn sse_rejects_oversize_frames_and_invalid_utf8() {
    assert!(SseDecoder::default()
        .push(&vec![b'a'; MAX_EVENT_BYTES + 1])
        .is_err());
    assert!(SseDecoder::default().push(&[255, b'\n']).is_err());
}
#[test]
fn notification_scope_and_shape_are_verified() {
    let value = notification();
    assert!(matches!(
        decode_event(&value.to_string(), ACCOUNT, WORKSPACE),
        Ok(RealtimeEvent::Notification(_))
    ));
    assert!(decode_event(&value.to_string(), STORY, WORKSPACE).is_err());
    assert!(decode_event(&value.to_string(), ACCOUNT, STORY).is_err());
    let mut value = value;
    value["entity_id"] = json!("https://attacker.invalid");
    assert!(decode_event(&value.to_string(), ACCOUNT, WORKSPACE).is_err());
}
#[test]
fn story_updates_reject_arbitrary_fields_and_invalid_values() {
    for changes in [
        json!({"title":"Updated"}),
        json!({"assigneeId":null}),
        json!({"checklistChanged":true}),
    ] {
        assert!(decode_event(
            &json!({"type":"story.workspace_update","storyId":STORY,"changes":changes}).to_string(),
            ACCOUNT,
            WORKSPACE
        )
        .is_ok());
    }
    for changes in [
        json!({"id":WORKSPACE}),
        json!({"priority":"secret"}),
        json!({"title":false}),
        json!({}),
    ] {
        assert!(decode_event(
            &json!({"type":"story.workspace_update","storyId":STORY,"changes":changes}).to_string(),
            ACCOUNT,
            WORKSPACE
        )
        .is_err());
    }
}
#[test]
fn activation_accepts_only_current_account_and_known_resource_kinds() {
    let target = serde_json::from_value::<Notification>(notification())
        .unwrap()
        .activation();
    assert!(activation_target(&target, ACCOUNT).is_some());
    assert!(activation_target(&target, WORKSPACE).is_none());
    let mut invalid = target.clone();
    invalid["kind"] = json!("https://attacker.invalid");
    assert!(activation_target(&invalid, ACCOUNT).is_none());
    invalid = target;
    invalid["resourceId"] = json!("../other");
    assert!(activation_target(&invalid, ACCOUNT).is_none());
}
#[test]
fn receipts_are_durable_scoped_and_allow_new_coalesced_versions() {
    let path = std::env::temp_dir().join(format!(
        "fortyone-notifications-{}.sqlite",
        uuid::Uuid::new_v4()
    ));
    let store = Store::open(&path).unwrap();
    store
        .save_account(&json!({"id":ACCOUNT}), &json!([{"id":WORKSPACE}]))
        .unwrap();
    initialize(&store).unwrap();
    let mut value = notification();
    let first = decode_event(&value.to_string(), ACCOUNT, WORKSPACE).unwrap();
    assert!(update_cache(&store, ACCOUNT, WORKSPACE, &first).unwrap());
    drop(store);
    let store = Store::open(&path).unwrap();
    initialize(&store).unwrap();
    assert!(!update_cache(&store, ACCOUNT, WORKSPACE, &first).unwrap());
    value["created_at"] = json!("2026-10-14T12:01:00Z");
    assert!(update_cache(
        &store,
        ACCOUNT,
        WORKSPACE,
        &decode_event(&value.to_string(), ACCOUNT, WORKSPACE).unwrap()
    )
    .unwrap());
    let db = store.0.lock().unwrap();
    let snapshot: String = db
        .query_row("SELECT data FROM snapshots", [], |r| r.get(0))
        .unwrap();
    let snapshot: Value = serde_json::from_str(&snapshot).unwrap();
    assert_eq!(snapshot["notifications"].as_array().unwrap().len(), 1);
    assert_eq!(snapshot["unreadNotificationCount"], 1);
    assert_eq!(
        db.query_row("SELECT count(*) FROM notification_receipts", [], |r| r
            .get::<_, i64>(0))
            .unwrap(),
        2
    );
    drop(db);
    store
        .save_account(&json!({"id":ACCOUNT}), &json!([]))
        .unwrap();
    assert!(update_cache(&store, ACCOUNT, WORKSPACE, &first).is_err());
    prune_receipts(&store, ACCOUNT).unwrap();
    assert_eq!(
        store
            .0
            .lock()
            .unwrap()
            .query_row("SELECT count(*) FROM notification_receipts", [], |row| row
                .get::<_, i64>(
                0
            ))
            .unwrap(),
        0
    );
    drop(store);
    for suffix in ["", "-wal", "-shm"] {
        let _ = std::fs::remove_file(format!("{}{suffix}", path.display()));
    }
}
