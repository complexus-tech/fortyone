use super::*;
use std::path::PathBuf;

const ACCOUNT: &str = "identity-account";
const WORKSPACE: &str = "21111111-1111-4111-8111-111111111111";
const LOCAL: &str = "31111111-1111-4111-8111-111111111111";
const SERVER: &str = "41111111-1111-4111-8111-111111111111";

struct Fixture {
    path: PathBuf,
    store: Option<Store>,
}

impl Fixture {
    fn new() -> Self {
        let path = std::env::temp_dir().join(format!(
            "fortyone-identities-{}.sqlite",
            uuid::Uuid::new_v4()
        ));
        let store = Store::open(&path).unwrap();
        store
            .save_account(&json!({"id":ACCOUNT}), &json!([{"id":WORKSPACE}]))
            .unwrap();
        Self {
            path,
            store: Some(store),
        }
    }
    fn store(&self) -> &Store {
        self.store.as_ref().unwrap()
    }
    fn draft(&self, resource: &str, payload: Value) {
        self.store()
            .0
            .lock()
            .unwrap()
            .execute(
                "INSERT INTO drafts VALUES(?,?,?,?,?)",
                params![ACCOUNT, WORKSPACE, resource, "story", payload.to_string()],
            )
            .unwrap();
    }
    fn acknowledge(&self, operation: &Operation) -> Result<(), String> {
        self.store().acknowledge(ACCOUNT, WORKSPACE, operation, &json!({"resourceId":SERVER,"operationVersion":"created-v1","resource":{"id":SERVER,"title":"Saved","updatedAt":"current-v3"}}))
    }
}

impl Drop for Fixture {
    fn drop(&mut self) {
        drop(self.store.take());
        for suffix in ["", "-wal", "-shm"] {
            let _ = std::fs::remove_file(format!("{}{suffix}", self.path.display()));
        }
    }
}

fn creation() -> Operation {
    Operation {
        id: uuid::Uuid::new_v4().to_string(),
        kind: "story.create".into(),
        resource_id: LOCAL.into(),
        payload: json!({"title":"Saved"}),
        base_version: None,
        status: "pending".into(),
        error: None,
        server_version: None,
        server_value: None,
        attempted: false,
    }
}

#[test]
fn duplicate_creation_is_idempotent_before_and_after_restarting_after_ack() {
    let mut fixture = Fixture::new();
    let original = creation();
    fixture
        .store()
        .enqueue(ACCOUNT, WORKSPACE, &original)
        .unwrap();
    fixture
        .store()
        .enqueue(ACCOUNT, WORKSPACE, &creation())
        .unwrap();
    assert_eq!(
        fixture
            .store()
            .operations(ACCOUNT, WORKSPACE)
            .unwrap()
            .len(),
        1
    );
    fixture.acknowledge(&original).unwrap();
    drop(fixture.store.take());
    fixture.store = Some(Store::open(&fixture.path).unwrap());
    fixture
        .store()
        .enqueue(ACCOUNT, WORKSPACE, &creation())
        .unwrap();
    let mut remapped_creation = creation();
    remapped_creation.resource_id = SERVER.into();
    fixture
        .store()
        .enqueue(ACCOUNT, WORKSPACE, &remapped_creation)
        .unwrap();
    assert!(fixture
        .store()
        .operations(ACCOUNT, WORKSPACE)
        .unwrap()
        .is_empty());
    let mut different = creation();
    different.payload["title"] = json!("New draft edit");
    assert!(fixture
        .store()
        .enqueue(ACCOUNT, WORKSPACE, &different)
        .is_err());
    let db = fixture.store().0.lock().unwrap();
    assert_eq!(
        resolve_resource(&db, ACCOUNT, WORKSPACE, LOCAL).unwrap(),
        SERVER
    );
    assert_eq!(
        resolve_resource(&db, "another-account", WORKSPACE, LOCAL).unwrap(),
        LOCAL
    );
    assert_eq!(
        resolve_resource(&db, ACCOUNT, "another-workspace", LOCAL).unwrap(),
        LOCAL
    );
}

#[test]
fn draft_consumption_and_enqueue_commit_or_rollback_together() {
    let fixture = Fixture::new();
    let operation = creation();
    fixture.draft("new-story", json!({"resourceId":LOCAL,"title":"Saved"}));
    let invalid = ConsumedDraft {
        resource_id: "new-story".into(),
        kind: "unsupported".into(),
    };
    assert!(fixture
        .store()
        .enqueue_consuming(ACCOUNT, WORKSPACE, &operation, Some(&invalid))
        .is_err());
    assert!(fixture
        .store()
        .operations(ACCOUNT, WORKSPACE)
        .unwrap()
        .is_empty());
    let consumed = ConsumedDraft {
        resource_id: "new-story".into(),
        kind: "story".into(),
    };
    fixture
        .store()
        .enqueue_consuming(ACCOUNT, WORKSPACE, &operation, Some(&consumed))
        .unwrap();
    assert_eq!(
        fixture
            .store()
            .operations(ACCOUNT, WORKSPACE)
            .unwrap()
            .len(),
        1
    );
    let count: usize = fixture
        .store()
        .0
        .lock()
        .unwrap()
        .query_row("SELECT count(*) FROM drafts", [], |row| row.get(0))
        .unwrap();
    assert_eq!(count, 0);
}

#[test]
fn acknowledgement_remaps_saved_drafts_and_stale_editor_saves_without_rebasing_other_versions() {
    let fixture = Fixture::new();
    let operation = creation();
    fixture
        .store()
        .enqueue(ACCOUNT, WORKSPACE, &operation)
        .unwrap();
    fixture.draft(LOCAL, json!({"resourceId":LOCAL,"baseVersion":format!("pending-create:{LOCAL}"),"title":"Editor still open","description":LOCAL}));
    fixture.acknowledge(&operation).unwrap();
    let encoded: String = fixture
        .store()
        .0
        .lock()
        .unwrap()
        .query_row(
            "SELECT payload FROM drafts WHERE resource=?",
            [SERVER],
            |row| row.get(0),
        )
        .unwrap();
    let payload: Value = serde_json::from_str(&encoded).unwrap();
    assert_eq!(payload["resourceId"], SERVER);
    assert_eq!(payload["baseVersion"], "created-v1");
    assert_eq!(
        payload["description"], LOCAL,
        "user-authored text is not an identity field"
    );
    let mut stale = creation();
    stale.kind = "story.update".into();
    stale.base_version = Some(format!("pending-create:{LOCAL}"));
    stale.payload = json!({"title":"New local edit"});
    fixture.store().enqueue(ACCOUNT, WORKSPACE, &stale).unwrap();
    let mut original_base = stale.clone();
    original_base.id = uuid::Uuid::new_v4().to_string();
    original_base.base_version = Some("original-server-v0".into());
    fixture
        .store()
        .enqueue(ACCOUNT, WORKSPACE, &original_base)
        .unwrap();
    let pending = fixture.store().operations(ACCOUNT, WORKSPACE).unwrap();
    assert_eq!(pending[0].resource_id, SERVER);
    assert_eq!(pending[0].base_version.as_deref(), Some("created-v1"));
    assert_eq!(
        pending[1].base_version.as_deref(),
        Some("original-server-v0")
    );
}

#[test]
fn conflicting_draft_keys_do_not_partially_acknowledge_or_discard_local_work() {
    let fixture = Fixture::new();
    let operation = creation();
    fixture
        .store()
        .enqueue(ACCOUNT, WORKSPACE, &operation)
        .unwrap();
    fixture.draft(LOCAL, json!({"resourceId":LOCAL,"title":"First draft"}));
    fixture.draft(SERVER, json!({"resourceId":SERVER,"title":"Other draft"}));
    assert!(fixture.acknowledge(&operation).is_err());
    assert_eq!(
        fixture
            .store()
            .operations(ACCOUNT, WORKSPACE)
            .unwrap()
            .len(),
        1
    );
    let db = fixture.store().0.lock().unwrap();
    assert_eq!(
        resolve_resource(&db, ACCOUNT, WORKSPACE, LOCAL).unwrap(),
        LOCAL
    );
    let count: usize = db
        .query_row("SELECT count(*) FROM drafts", [], |row| row.get(0))
        .unwrap();
    assert_eq!(count, 2);
}

#[test]
fn account_changes_reject_stale_enqueue_without_consuming_the_draft() {
    let fixture = Fixture::new();
    fixture.draft("new-story", json!({"resourceId":LOCAL,"title":"Saved"}));
    fixture
        .store()
        .save_account(&json!({"id":"another-account"}), &json!([{"id":WORKSPACE}]))
        .unwrap();
    let consumed = ConsumedDraft {
        resource_id: "new-story".into(),
        kind: "story".into(),
    };
    assert!(fixture
        .store()
        .enqueue_consuming(ACCOUNT, WORKSPACE, &creation(), Some(&consumed))
        .is_err());
    assert!(fixture
        .store()
        .operations(ACCOUNT, WORKSPACE)
        .unwrap()
        .is_empty());
    let count: usize = fixture
        .store()
        .0
        .lock()
        .unwrap()
        .query_row("SELECT count(*) FROM drafts", [], |row| row.get(0))
        .unwrap();
    assert_eq!(count, 1);
}
