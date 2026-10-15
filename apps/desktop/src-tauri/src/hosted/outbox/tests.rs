use super::*;

fn database() -> Connection {
    let db = Connection::open_in_memory().unwrap();
    super::super::initialize(&db).unwrap();
    db
}
fn request(kind: &str, resource: &str, payload: Value, base: Option<&str>) -> EnqueueRequest {
    let scope = super::super::tests::scope();
    EnqueueRequest {
        account_id: scope.account,
        workspace_id: scope.workspace,
        id: uuid::Uuid::new_v4().to_string(),
        kind: kind.into(),
        resource_id: resource.into(),
        payload,
        base_version: base.map(str::to_owned),
        consumed_draft_key: None,
    }
}
fn enqueue(db: &mut Connection, scope: &BoundScope, request: &EnqueueRequest) -> Submission {
    let tx = db.transaction().unwrap();
    let result = store::enqueue(&tx, scope, request).unwrap();
    tx.commit().unwrap();
    result
}
fn ack(db: &mut Connection, scope: &BoundScope, operation: &Operation, result: Value) {
    let tx = db.transaction().unwrap();
    store::acknowledge(&tx, scope, operation, &result).unwrap();
    tx.commit().unwrap();
}

#[test]
fn repeated_enqueue_and_create_ack_never_create_a_second_server_command() {
    let mut db = database();
    let scope = super::super::tests::scope();
    let provisional = uuid::Uuid::new_v4().to_string();
    let mut create = request(
        "story.create",
        &provisional,
        json!({"title":"Saved locally","teamId":uuid::Uuid::new_v4().to_string()}),
        None,
    );
    let first = enqueue(&mut db, &scope, &create);
    assert_eq!(
        enqueue(&mut db, &scope, &create).operation_id,
        first.operation_id
    );
    create.id = uuid::Uuid::new_v4().to_string();
    assert_eq!(
        enqueue(&mut db, &scope, &create).operation_id,
        first.operation_id
    );
    let operation = store::read(&db, &scope).unwrap().operations.remove(0);
    let authoritative = uuid::Uuid::new_v4().to_string();
    ack(
        &mut db,
        &scope,
        &operation,
        json!({"resourceId":authoritative,"operationVersion":"v1"}),
    );
    let resumed = enqueue(&mut db, &scope, &create);
    assert!(!resumed.pending);
    assert_eq!(resumed.resource_id, authoritative);
    create.id = uuid::Uuid::new_v4().to_string();
    assert!(!enqueue(&mut db, &scope, &create).pending);
    assert!(store::read(&db, &scope).unwrap().operations.is_empty());
    create.payload["title"] = json!("different");
    assert!(store::enqueue(&db, &scope, &create).is_err());
}

#[test]
fn immutable_receipt_version_rebases_unattempted_dependents_only() {
    let mut db = database();
    let scope = super::super::tests::scope();
    let resource = uuid::Uuid::new_v4().to_string();
    let first = request("story.update", &resource, json!({"title":"A"}), Some("v1"));
    let second = request(
        "story.update",
        &resource,
        json!({"priority":"high"}),
        Some("v1"),
    );
    enqueue(&mut db, &scope, &first);
    enqueue(&mut db, &scope, &second);
    let operation = store::read(&db, &scope).unwrap().operations.remove(0);
    ack(
        &mut db,
        &scope,
        &operation,
        json!({"resourceId":resource,"operationVersion":"v2","resource":{"id":resource,"updatedAt":"v3","title":"Someone else"}}),
    );
    let pending = store::read(&db, &scope).unwrap().operations;
    assert_eq!(pending[0].base_version.as_deref(), Some("v2"));
    assert_eq!(
        store::read(&db, &scope).unwrap().completed[0]["operationVersion"],
        "v2"
    );
}

#[test]
fn provisional_creation_rewrites_comment_dependencies_and_pending_base_markers() {
    let mut db = database();
    let scope = super::super::tests::scope();
    let local = uuid::Uuid::new_v4().to_string();
    let remote = uuid::Uuid::new_v4().to_string();
    enqueue(
        &mut db,
        &scope,
        &request("story.create", &local, json!({"title":"Local"}), None),
    );
    let update = request(
        "story.update",
        &local,
        json!({"title":"Later"}),
        Some(&format!("pending-create:{local}")),
    );
    enqueue(&mut db, &scope, &update);
    enqueue(
        &mut db,
        &scope,
        &request(
            "comment.create",
            &uuid::Uuid::new_v4().to_string(),
            json!({"storyId":local,"comment":"Comment"}),
            None,
        ),
    );
    let operation = store::read(&db, &scope).unwrap().operations.remove(0);
    ack(
        &mut db,
        &scope,
        &operation,
        json!({"resourceId":remote,"operationVersion":"v1"}),
    );
    let pending = store::read(&db, &scope).unwrap().operations;
    assert_eq!(pending[0].resource_id, remote);
    assert_eq!(pending[0].base_version.as_deref(), Some("v1"));
    assert_eq!(pending[1].payload["storyId"], remote);
    assert_eq!(enqueue(&mut db, &scope, &update).resource_id, remote);
    let late = request(
        "story.update",
        &local,
        json!({"priority":"high"}),
        Some(&format!("pending-create:{local}")),
    );
    assert_eq!(enqueue(&mut db, &scope, &late).resource_id, remote);
}

#[test]
fn draft_consumption_rolls_back_if_queue_save_fails_and_duplicate_save_keeps_newer_draft() {
    let mut db = database();
    let scope = super::super::tests::scope();
    let resource = uuid::Uuid::new_v4().to_string();
    let key = format!("story:{resource}");
    db.execute(
        "INSERT INTO hosted_drafts VALUES(?,?,?,?,?,?)",
        params![
            scope.origin,
            scope.account,
            scope.workspace,
            key,
            "{\"title\":\"old\"}",
            now()
        ],
    )
    .unwrap();
    let mut change = request(
        "story.update",
        &resource,
        json!({"title":"Saved"}),
        Some("v1"),
    );
    change.consumed_draft_key = Some(key.clone());
    db.execute_batch("CREATE TRIGGER reject_queue BEFORE INSERT ON hosted_operations BEGIN SELECT RAISE(ABORT,'disk full'); END;").unwrap();
    {
        let tx = db.transaction().unwrap();
        assert!(store::enqueue(&tx, &scope, &change).is_err());
    }
    assert_eq!(
        db.query_row("SELECT COUNT(*) FROM hosted_drafts", [], |row| row
            .get::<_, u64>(0))
            .unwrap(),
        1
    );
    assert_eq!(
        db.query_row("SELECT COUNT(*) FROM hosted_inputs", [], |row| row
            .get::<_, u64>(0))
            .unwrap(),
        0
    );
    db.execute_batch("DROP TRIGGER reject_queue;").unwrap();
    enqueue(&mut db, &scope, &change);
    db.execute(
        "INSERT INTO hosted_drafts VALUES(?,?,?,?,?,?)",
        params![
            scope.origin,
            scope.account,
            scope.workspace,
            key,
            "{\"title\":\"newer\"}",
            now()
        ],
    )
    .unwrap();
    enqueue(&mut db, &scope, &change);
    assert_eq!(
        db.query_row("SELECT COUNT(*) FROM hosted_drafts", [], |row| row
            .get::<_, u64>(0))
            .unwrap(),
        1
    );
}

#[test]
fn conflict_resolution_preserves_queue_order_and_unknown_outcomes_are_immutable() {
    let mut db = database();
    let scope = super::super::tests::scope();
    let resource = uuid::Uuid::new_v4().to_string();
    let first = request(
        "story.update",
        &resource,
        json!({"title":"First"}),
        Some("v1"),
    );
    let second = request(
        "story.update",
        &resource,
        json!({"title":"Second"}),
        Some("v1"),
    );
    enqueue(&mut db, &scope, &first);
    enqueue(&mut db, &scope, &second);
    let mut operation = store::read(&db, &scope).unwrap().operations.remove(0);
    operation.attempted = true;
    store::save(&db, &scope, &operation).unwrap();
    let mut resolution = ResolveRequest {
        account_id: scope.account.clone(),
        workspace_id: scope.workspace.clone(),
        operation_id: operation.id.clone(),
        resolution: "discard".into(),
        payload: None,
        base_version: None,
    };
    assert!(store::resolve(&db, &scope, &resolution).is_err());
    operation.status = "conflict".into();
    operation.server_version = Some("v2".into());
    store::save(&db, &scope, &operation).unwrap();
    resolution.resolution = "retry".into();
    assert!(store::resolve(&db, &scope, &resolution).is_err());
    resolution.payload = Some(json!({"title":"Reviewed"}));
    resolution.base_version = Some("v2".into());
    {
        let tx = db.transaction().unwrap();
        store::resolve(&tx, &scope, &resolution).unwrap();
        tx.commit().unwrap();
    }
    let pending = store::read(&db, &scope).unwrap().operations;
    assert_ne!(pending[0].id, first.id);
    assert_eq!(pending[0].payload["title"], "Reviewed");
    assert_eq!(pending[1].id, second.id);
    assert_eq!(
        next_operation(pending).unwrap().payload["title"],
        "Reviewed"
    );
    assert!(store::enqueue(&db, &scope, &first).is_err());
}

#[test]
fn workspace_account_and_origin_isolate_durable_queue_and_acknowledgements() {
    let mut db = database();
    let scope = super::super::tests::scope();
    let change = request(
        "story.create",
        &uuid::Uuid::new_v4().to_string(),
        json!({"title":"Private"}),
        None,
    );
    enqueue(&mut db, &scope, &change);
    for other in [
        BoundScope {
            account: "other".into(),
            ..super::super::tests::scope()
        },
        BoundScope {
            workspace: "other".into(),
            ..super::super::tests::scope()
        },
        BoundScope {
            origin: "https://other.fortyone.app".into(),
            ..super::super::tests::scope()
        },
    ] {
        assert!(store::read(&db, &other).unwrap().operations.is_empty());
    }
    let mut changed = change;
    changed.payload = json!({"title":"Reused ID"});
    assert!(store::enqueue(&db, &scope, &changed).is_err());
}
