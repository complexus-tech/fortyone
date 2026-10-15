use super::{
    model::{next_operation, validate_operation, Operation},
    store::Store,
};
use serde_json::json;
use std::path::PathBuf;

const ACCOUNT: &str = "account-a";
const WORKSPACE: &str = "21111111-1111-4111-8111-111111111111";
const STORY: &str = "31111111-1111-4111-8111-111111111111";

struct Database {
    path: PathBuf,
    store: Option<Store>,
}
impl Database {
    fn new() -> Self {
        let path =
            std::env::temp_dir().join(format!("fortyone-offline-{}.sqlite", uuid::Uuid::new_v4()));
        let store = Store::open(&path).unwrap();
        store
            .save_account(
                &json!({"id":ACCOUNT,"fullName":"Test"}),
                &json!([{"id":WORKSPACE,"slug":"test","name":"Test"}]),
            )
            .unwrap();
        Self {
            path,
            store: Some(store),
        }
    }
    fn store(&self) -> &Store {
        self.store.as_ref().unwrap()
    }
}
impl Drop for Database {
    fn drop(&mut self) {
        drop(self.store.take());
        for path in [
            self.path.clone(),
            PathBuf::from(format!("{}-wal", self.path.display())),
            PathBuf::from(format!("{}-shm", self.path.display())),
        ] {
            let _ = std::fs::remove_file(path);
        }
    }
}

fn operation(kind: &str) -> Operation {
    Operation {
        id: uuid::Uuid::new_v4().to_string(),
        kind: kind.into(),
        resource_id: STORY.into(),
        payload: json!({"title":"Offline title"}),
        base_version: None,
        status: "pending".into(),
        error: None,
        server_version: None,
        server_value: None,
        attempted: false,
    }
}
fn snapshot() -> serde_json::Value {
    json!({"stories":[{"id":STORY,"title":"Original"}],"documents":[],"comments":[],"teams":[],"statuses":[],"complete":{"stories":true,"documents":true},"syncedAt":"2026-10-14T12:00:00Z"})
}

#[test]
fn local_changes_survive_reopening_without_mutating_server_snapshot() {
    let mut db = Database::new();
    db.store()
        .snapshot(ACCOUNT, WORKSPACE, &snapshot())
        .unwrap();
    let mut op = operation("story.update");
    op.base_version = Some("2026-10-14T12:00:00Z".into());
    db.store().enqueue(ACCOUNT, WORKSPACE, &op).unwrap();
    drop(db.store.take());
    db.store = Some(Store::open(&db.path).unwrap());
    let view = db.store().workspace(ACCOUNT, WORKSPACE).unwrap();
    assert_eq!(view["stories"][0]["title"], "Offline title");
    assert_eq!(view["pending"][0]["id"], op.id);
    let encoded: String = db
        .store()
        .0
        .lock()
        .unwrap()
        .query_row("SELECT data FROM snapshots", [], |r| r.get(0))
        .unwrap();
    assert_eq!(
        serde_json::from_str::<serde_json::Value>(&encoded).unwrap()["stories"][0]["title"],
        "Original"
    );
}

#[test]
fn incomplete_snapshots_retain_downloads_and_apply_explicit_removals() {
    let db = Database::new();
    db.store()
        .snapshot(ACCOUNT, WORKSPACE, &snapshot())
        .unwrap();
    let mut data = snapshot();
    data["stories"] = json!([]);
    data["complete"]["stories"] = json!(false);
    db.store().snapshot(ACCOUNT, WORKSPACE, &data).unwrap();
    assert_eq!(
        db.store().workspace(ACCOUNT, WORKSPACE).unwrap()["stories"]
            .as_array()
            .unwrap()
            .len(),
        1
    );
    data["removed"] = json!({"stories":[STORY]});
    db.store().snapshot(ACCOUNT, WORKSPACE, &data).unwrap();
    assert!(db.store().workspace(ACCOUNT, WORKSPACE).unwrap()["stories"]
        .as_array()
        .unwrap()
        .is_empty());
}

#[test]
fn acknowledgement_maps_dependents_and_removes_only_acknowledged_change() {
    let db = Database::new();
    let op = operation("story.create");
    db.store().enqueue(ACCOUNT, WORKSPACE, &op).unwrap();
    let mut comment = operation("comment.create");
    comment.resource_id = uuid::Uuid::new_v4().to_string();
    comment.payload = json!({"storyId":STORY,"comment":STORY});
    db.store().enqueue(ACCOUNT, WORKSPACE, &comment).unwrap();
    let server = uuid::Uuid::new_v4().to_string();
    db.store().acknowledge(ACCOUNT,WORKSPACE,&op,&json!({"resourceId":server,"resource":{"id":server,"title":"Offline title","updatedAt":"v2"}})).unwrap();
    let pending = db.store().operations(ACCOUNT, WORKSPACE).unwrap();
    assert_eq!(pending.len(), 1);
    assert_eq!(pending[0].payload["storyId"], server);
    assert_eq!(pending[0].payload["comment"], STORY);
    assert_eq!(
        db.store().workspace(ACCOUNT, WORKSPACE).unwrap()["stories"][0]["id"],
        server
    );
}

#[test]
fn scope_checks_prevent_cross_account_workspace_reads_and_writes() {
    let db = Database::new();
    assert!(db.store().workspace("other", WORKSPACE).is_err());
    assert!(db
        .store()
        .enqueue(ACCOUNT, "other", &operation("story.create"))
        .is_err());
}

#[test]
fn offline_updates_require_the_original_server_version() {
    assert!(validate_operation(&operation("story.update")).is_err());
    let mut op = operation("story.update");
    op.base_version = Some("version".into());
    assert!(validate_operation(&op).is_ok());
}

#[test]
fn newer_schema_is_never_silently_downgraded() {
    let mut db = Database::new();
    db.store()
        .0
        .lock()
        .unwrap()
        .pragma_update(None, "user_version", 2)
        .unwrap();
    drop(db.store.take());
    assert!(Store::open(&db.path).is_err());
}

#[test]
fn replay_acknowledgement_uses_committed_version_instead_of_newer_resource() {
    let db = Database::new();
    let mut first = operation("story.update");
    first.base_version = Some("v1".into());
    let mut dependent = operation("story.update");
    dependent.base_version = Some("v1".into());
    dependent.payload = json!({"title":"Second local edit"});
    let mut independently_based = operation("story.update");
    independently_based.base_version = Some("v0".into());
    for change in [&first, &dependent, &independently_based] {
        db.store().enqueue(ACCOUNT, WORKSPACE, change).unwrap();
    }

    // The first write committed v2, its response was lost, and another actor
    // subsequently wrote v3. Replaying its receipt must not approve overwriting v3.
    db.store()
        .acknowledge(
            ACCOUNT,
            WORKSPACE,
            &first,
            &json!({"resourceId":STORY,"operationVersion":"v2","replayed":true,
                "resource":{"id":STORY,"title":"Other actor","updatedAt":"v3"}}),
        )
        .unwrap();
    let pending = db.store().operations(ACCOUNT, WORKSPACE).unwrap();
    assert_eq!(pending.len(), 2);
    assert_eq!(pending[0].id, dependent.id);
    assert_eq!(pending[0].base_version.as_deref(), Some("v2"));
    assert_eq!(pending[1].base_version.as_deref(), Some("v0"));
    assert_eq!(
        db.store().server_snapshot(ACCOUNT, WORKSPACE).unwrap()["stories"][0]["updatedAt"],
        "v3"
    );
}

#[test]
fn missing_receipt_version_does_not_rebase_a_dependent_edit() {
    let db = Database::new();
    let mut first = operation("document.update");
    first.base_version = Some("1".into());
    let dependent = Operation {
        id: uuid::Uuid::new_v4().to_string(),
        ..first.clone()
    };
    db.store().enqueue(ACCOUNT, WORKSPACE, &first).unwrap();
    db.store().enqueue(ACCOUNT, WORKSPACE, &dependent).unwrap();
    db.store()
        .acknowledge(
            ACCOUNT,
            WORKSPACE,
            &first,
            &json!({"resourceId":STORY,"operationVersion":null,
                "resource":{"id":STORY,"revision":9}}),
        )
        .unwrap();
    assert_eq!(
        db.store().operations(ACCOUNT, WORKSPACE).unwrap()[0]
            .base_version
            .as_deref(),
        Some("1")
    );
}

fn insert_collaboration(db: &Database, resource: &str, dirty: bool) {
    db.store()
        .0
        .lock()
        .unwrap()
        .execute(
            "INSERT INTO drafts VALUES(?,?,?,?,?)",
            rusqlite::params![
                ACCOUNT,
                WORKSPACE,
                resource,
                "collaboration",
                json!({"dirty":dirty,"epoch":1,"state":"saved-state"}).to_string()
            ],
        )
        .unwrap();
}

#[test]
fn tombstones_remove_cached_children_but_preserve_unsynced_recovery() {
    let db = Database::new();
    let retained_story = uuid::Uuid::new_v4().to_string();
    let clean_document = uuid::Uuid::new_v4().to_string();
    let dirty_document = uuid::Uuid::new_v4().to_string();
    let mut initial = snapshot();
    initial["stories"] = json!([{"id":STORY},{"id":retained_story}]);
    initial["documents"] = json!([{"id":clean_document},{"id":dirty_document}]);
    initial["comments"] = json!([
        {"id":"removed-thread","storyId":STORY,"replies":[{"id":"removed-reply"}]},
        {"id":"retained-thread","storyId":retained_story}
    ]);
    db.store().snapshot(ACCOUNT, WORKSPACE, &initial).unwrap();
    insert_collaboration(&db, &clean_document, false);
    insert_collaboration(&db, &dirty_document, true);
    let mut unsynced = operation("story.update");
    unsynced.base_version = Some("v1".into());
    db.store().enqueue(ACCOUNT, WORKSPACE, &unsynced).unwrap();

    let mut removal = snapshot();
    removal["stories"] = json!([]);
    removal["complete"] = json!({});
    removal["removed"] = json!({"stories":[STORY],"documents":[clean_document,dirty_document]});
    db.store().snapshot(ACCOUNT, WORKSPACE, &removal).unwrap();
    let cached = db.store().server_snapshot(ACCOUNT, WORKSPACE).unwrap();
    assert_eq!(cached["comments"].as_array().unwrap().len(), 1);
    assert_eq!(cached["comments"][0]["id"], "retained-thread");
    assert!(cached["documents"].as_array().unwrap().is_empty());
    assert_eq!(
        db.store().operations(ACCOUNT, WORKSPACE).unwrap()[0].id,
        unsynced.id
    );
    let retained_draft: String = db
        .store()
        .0
        .lock()
        .unwrap()
        .query_row("SELECT resource FROM drafts", [], |row| row.get(0))
        .unwrap();
    assert_eq!(retained_draft, dirty_document);
}

#[test]
fn workspace_denial_purges_server_cache_and_survives_global_account_refresh() {
    let db = Database::new();
    db.store().grant_workspace(ACCOUNT, WORKSPACE).unwrap();
    db.store()
        .snapshot(ACCOUNT, WORKSPACE, &snapshot())
        .unwrap();
    insert_collaboration(&db, "clean-document", false);
    insert_collaboration(&db, "dirty-document", true);
    let unsynced = operation("story.create");
    db.store().enqueue(ACCOUNT, WORKSPACE, &unsynced).unwrap();
    db.store()
        .0
        .lock()
        .unwrap()
        .execute(
            "INSERT INTO workspace_settings VALUES(?,?,?)",
            rusqlite::params![ACCOUNT, WORKSPACE, "{\"members\":[]}"],
        )
        .unwrap();

    db.store().deny_workspace(ACCOUNT, WORKSPACE).unwrap();
    db.store()
        .save_account(
            &json!({"id":ACCOUNT}),
            &json!([{"id":WORKSPACE,"slug":"test"}]),
        )
        .unwrap();
    assert!(db.store().workspace(ACCOUNT, WORKSPACE).is_err());
    assert!(db.store().server_snapshot(ACCOUNT, WORKSPACE).is_err());
    assert!(db
        .store()
        .enqueue(ACCOUNT, WORKSPACE, &operation("story.create"))
        .is_err());
    assert_eq!(
        db.store().operations(ACCOUNT, WORKSPACE).unwrap()[0].id,
        unsynced.id
    );
    let counts: (i64, i64, i64) = db
        .store()
        .0
        .lock()
        .unwrap()
        .query_row(
            "SELECT (SELECT count(*) FROM snapshots), (SELECT count(*) FROM workspace_settings), (SELECT count(*) FROM drafts)",
            [],
            |row| Ok((row.get(0)?, row.get(1)?, row.get(2)?)),
        )
        .unwrap();
    assert_eq!(counts, (0, 0, 1));
    db.store().grant_workspace(ACCOUNT, WORKSPACE).unwrap();
    assert!(db.store().workspace(ACCOUNT, WORKSPACE).is_ok());
}

#[test]
fn account_refresh_does_not_renew_an_expired_workspace_lease() {
    let db = Database::new();
    db.store()
        .0
        .lock()
        .unwrap()
        .execute(
            "INSERT INTO workspace_access VALUES(?,?,0,NULL)",
            rusqlite::params![ACCOUNT, WORKSPACE],
        )
        .unwrap();
    db.store()
        .save_account(
            &json!({"id":ACCOUNT}),
            &json!([{"id":WORKSPACE,"slug":"test"}]),
        )
        .unwrap();
    assert!(db.store().membership(ACCOUNT, WORKSPACE).is_ok());
    assert!(db.store().authorize(ACCOUNT, WORKSPACE).is_err());
    db.store().grant_workspace(ACCOUNT, WORKSPACE).unwrap();
    assert!(db.store().authorize(ACCOUNT, WORKSPACE).is_ok());
}

#[test]
fn blocked_dependencies_do_not_prevent_independent_work_from_synchronizing() {
    let mut blocked = operation("story.create");
    blocked.status = "conflict".into();
    let mut comment = operation("comment.create");
    comment.resource_id = uuid::Uuid::new_v4().to_string();
    comment.payload = json!({"storyId":STORY,"comment":"Pending parent"});
    let mut reply = operation("comment.create");
    reply.resource_id = uuid::Uuid::new_v4().to_string();
    reply.payload = json!({"parentId":comment.resource_id,"comment":"Pending reply"});
    let mut independent = operation("document.create");
    independent.resource_id = uuid::Uuid::new_v4().to_string();
    assert_eq!(
        next_operation(vec![
            blocked.clone(),
            comment.clone(),
            reply.clone(),
            independent.clone()
        ])
        .unwrap()
        .id,
        independent.id
    );
    assert!(next_operation(vec![blocked, comment, reply]).is_none());
}

#[test]
fn attachment_acknowledgement_retains_bytes_and_original_target() {
    let db = Database::new();
    db.store()
        .snapshot(ACCOUNT, WORKSPACE, &snapshot())
        .unwrap();
    let mut attachment = operation("attachment.upload");
    attachment.payload =
        json!({"attachmentId":attachment.id,"targetKind":"story","fileName":"notes.txt"});
    db.store().enqueue(ACCOUNT, WORKSPACE, &attachment).unwrap();
    db.store()
        .0
        .lock()
        .unwrap()
        .execute(
            "INSERT INTO attachments VALUES(?,?,?,?,?)",
            rusqlite::params![
                attachment.id,
                ACCOUNT,
                WORKSPACE,
                "notes.txt",
                b"saved bytes".as_slice()
            ],
        )
        .unwrap();
    let mut dependent = operation("comment.create");
    dependent.resource_id = uuid::Uuid::new_v4().to_string();
    dependent.payload = json!({"storyId":STORY,"comment":"File attached"});
    db.store().enqueue(ACCOUNT, WORKSPACE, &dependent).unwrap();
    let uploaded = uuid::Uuid::new_v4().to_string();
    db.store()
        .acknowledge(
            ACCOUNT,
            WORKSPACE,
            &attachment,
            &json!({"resourceId":uploaded,"operationVersion":null,
                "resource":{"id":uploaded,"filename":"notes.txt","mimeType":"text/plain"}}),
        )
        .unwrap();
    let (stored_id, bytes): (String, Vec<u8>) = db
        .store()
        .0
        .lock()
        .unwrap()
        .query_row("SELECT id,bytes FROM attachments", [], |row| {
            Ok((row.get(0)?, row.get(1)?))
        })
        .unwrap();
    assert_eq!(stored_id, uploaded);
    assert_eq!(bytes, b"saved bytes");
    let cached = db.store().server_snapshot(ACCOUNT, WORKSPACE).unwrap();
    assert_eq!(cached["attachments"][0]["id"], uploaded);
    assert_eq!(cached["attachments"][0]["resourceId"], STORY);
    assert_eq!(cached["attachments"][0]["targetKind"], "story");
    assert_eq!(cached["stories"][0]["id"], STORY);
    assert_eq!(
        db.store().operations(ACCOUNT, WORKSPACE).unwrap()[0].payload["storyId"],
        STORY
    );
}

#[test]
fn pending_create_markers_require_matching_creation_and_rebase_after_acknowledgement() {
    let db = Database::new();
    let mut update = operation("document.update");
    update.base_version = Some(format!("pending-create:{STORY}"));
    assert!(db.store().enqueue(ACCOUNT, WORKSPACE, &update).is_err());
    let wrong_kind = operation("story.create");
    db.store().enqueue(ACCOUNT, WORKSPACE, &wrong_kind).unwrap();
    assert!(db.store().enqueue(ACCOUNT, WORKSPACE, &update).is_err());
    let create = operation("document.create");
    db.store().enqueue(ACCOUNT, WORKSPACE, &create).unwrap();
    let mut mismatched = update.clone();
    mismatched.base_version = Some(format!("pending-create:{}", uuid::Uuid::new_v4()));
    assert!(db.store().enqueue(ACCOUNT, WORKSPACE, &mismatched).is_err());
    db.store().enqueue(ACCOUNT, WORKSPACE, &update).unwrap();
    let server = uuid::Uuid::new_v4().to_string();
    db.store()
        .acknowledge(
            ACCOUNT,
            WORKSPACE,
            &create,
            &json!({"resourceId":server,"operationVersion":"1","resource":{"id":server,"revision":1}}),
        )
        .unwrap();
    let dependent = db
        .store()
        .operations(ACCOUNT, WORKSPACE)
        .unwrap()
        .into_iter()
        .find(|pending| pending.id == update.id)
        .unwrap();
    assert_eq!(dependent.resource_id, server);
    assert_eq!(dependent.base_version.as_deref(), Some("1"));
}
