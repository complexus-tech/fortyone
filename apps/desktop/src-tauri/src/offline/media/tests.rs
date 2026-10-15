use super::*;
const ACCOUNT: &str = "11111111-1111-4111-8111-111111111111";
const WORKSPACE: &str = "21111111-1111-4111-8111-111111111111";
const RESOURCE: &str = "31111111-1111-4111-8111-111111111111";
const ATTACHMENT: &str = "41111111-1111-4111-8111-111111111111";

#[test]
fn storage_addresses_cannot_forward_credentials_or_reach_local_services() {
    assert!(blob_url("https://storage.example.com/container/file?signature=opaque").is_ok());
    for url in [
        "http://storage.example.com/file",
        "file:///tmp/file",
        "https://user:secret@example.com/file",
        "https://localhost/file",
        "https://127.0.0.1/file",
        "https://10.0.0.1/file",
        "https://[::1]/file",
        "https://example.com:8443/file",
        "https://example.com/file#fragment",
    ] {
        assert!(blob_url(url).is_err(), "{url}");
    }
}
#[test]
fn inline_media_requires_allowed_mime_and_matching_bytes() {
    for (mime, bytes) in [
        ("image/png", b"\x89PNG\r\n\x1a\n".as_slice()),
        ("image/jpeg", &[0xff, 0xd8, 0xff]),
        ("image/gif", b"GIF89a"),
        ("image/webp", b"RIFF1234WEBP"),
        ("video/mp4", b"0000ftypisom"),
    ] {
        assert!(matches_content(mime, bytes));
        assert!(!matches_content(mime, b"<html><script>alert(1)</script>"));
    }
    assert!(!matches_content("image/svg+xml", b"<svg></svg>"));
    assert!(!matches_content("video/mp4", b"0000ftypqt  "));
}
#[test]
fn cache_survives_restart_and_is_bound_to_account_workspace_and_resource() {
    let path = std::env::temp_dir().join(format!("fortyone-media-{}.sqlite", uuid::Uuid::new_v4()));
    let store = Store::open(&path).unwrap();
    initialize(&store).unwrap();
    store
        .save_account(
            &json!({"id":ACCOUNT}),
            &json!([{"id":WORKSPACE,"slug":"acme"}]),
        )
        .unwrap();
    store.snapshot(ACCOUNT, WORKSPACE, &json!({"stories":[{"id":RESOURCE}],"documents":[],"comments":[],"teams":[],"statuses":[]})).unwrap();
    let mut scope = Scope {
        account: ACCOUNT.into(),
        workspace: WORKSPACE.into(),
        resource: RESOURCE.into(),
        kind: "story".into(),
        attachment: ATTACHMENT.into(),
        slug: "acme".into(),
    };
    save(&store, &scope, "image/png", b"\x89PNG\r\n\x1a\n").unwrap();
    drop(store);
    let store = Store::open(&path).unwrap();
    initialize(&store).unwrap();
    assert_eq!(
        cached(&store, &scope).unwrap().unwrap().mime_type,
        "image/png"
    );
    scope.resource = ATTACHMENT.into();
    assert!(cached(&store, &scope).is_err());
    scope.resource = RESOURCE.into();
    scope.kind = "document".into();
    assert!(cached(&store, &scope).is_err());
    scope.kind = "story".into();
    store
        .set_setting("account", &json!("another-account"))
        .unwrap();
    assert!(cached(&store, &scope).is_err());
    store.set_setting("account", &json!(ACCOUNT)).unwrap();
    store.deny_workspace(ACCOUNT, WORKSPACE).unwrap();
    assert!(cached(&store, &scope).is_err());
    assert!(save(&store, &scope, "image/png", b"\x89PNG\r\n\x1a\n").is_err());
    drop(store);
    for suffix in ["", "-wal", "-shm"] {
        let _ = std::fs::remove_file(format!("{}{suffix}", path.display()));
    }
}

fn with_cached_workspace(test: impl FnOnce(&Store, &Scope)) {
    let path = std::env::temp_dir().join(format!(
        "fortyone-media-cleanup-{}.sqlite",
        uuid::Uuid::new_v4()
    ));
    let store = Store::open(&path).unwrap();
    initialize(&store).unwrap();
    store
        .save_account(
            &json!({"id":ACCOUNT}),
            &json!([{"id":WORKSPACE,"slug":"acme"}]),
        )
        .unwrap();
    store.snapshot(ACCOUNT, WORKSPACE, &json!({"stories":[{"id":RESOURCE}],"documents":[{"id":"document"}],"comments":[],"teams":[],"statuses":[]})).unwrap();
    let scope = Scope {
        account: ACCOUNT.into(),
        workspace: WORKSPACE.into(),
        resource: RESOURCE.into(),
        kind: "story".into(),
        attachment: ATTACHMENT.into(),
        slug: "acme".into(),
    };
    save(&store, &scope, "image/png", b"\x89PNG\r\n\x1a\n").unwrap();
    {
        let db = store.0.lock().unwrap();
        for id in ["retained", "pending"] {
            db.execute(
                "INSERT INTO attachments VALUES(?,?,?,?,?)",
                params![id, ACCOUNT, WORKSPACE, "file.pdf", b"pdf".as_slice()],
            )
            .unwrap();
        }
        db.execute("UPDATE snapshots SET data=json_set(data,'$.attachments',json(?)) WHERE account=? AND workspace=?",params![json!([{"id":"retained","resourceId":RESOURCE,"targetKind":"story"},{"id":"pending","resourceId":RESOURCE,"targetKind":"story"}]).to_string(),ACCOUNT,WORKSPACE]).unwrap();
        db.execute("INSERT INTO operations VALUES(?,?,?,?)",params!["upload",ACCOUNT,WORKSPACE,json!({"kind":"attachment.upload","resourceId":RESOURCE,"payload":{"targetKind":"story","attachmentId":"pending"}}).to_string()]).unwrap();
        for (resource, dirty) in [("clean", false), ("dirty", true)] {
            db.execute(
                "INSERT INTO drafts VALUES(?,?,?,?,?)",
                params![
                    ACCOUNT,
                    WORKSPACE,
                    resource,
                    "collaboration",
                    json!({"dirty":dirty}).to_string()
                ],
            )
            .unwrap();
        }
        db.execute_batch(
            "CREATE TABLE notification_receipts(account TEXT,workspace TEXT,id TEXT,version TEXT);",
        )
        .unwrap();
        db.execute(
            "INSERT INTO notification_receipts VALUES(?,?,?,?)",
            params![ACCOUNT, WORKSPACE, "notification", "version"],
        )
        .unwrap();
    }
    test(&store, &scope);
    drop(store);
    for suffix in ["", "-wal", "-shm"] {
        let _ = std::fs::remove_file(format!("{}{suffix}", path.display()));
    }
}

fn count(store: &Store, table: &str) -> usize {
    store
        .0
        .lock()
        .unwrap()
        .query_row(&format!("SELECT COUNT(*) FROM {table}"), [], |row| {
            row.get(0)
        })
        .unwrap()
}

#[test]
fn workspace_denial_or_revocation_removes_server_files_but_retains_staged_work() {
    for revoked in [false, true] {
        with_cached_workspace(|store, scope| {
            if revoked {
                store
                    .save_account(&json!({"id":ACCOUNT}), &json!([]))
                    .unwrap();
            } else {
                store.deny_workspace(ACCOUNT, WORKSPACE).unwrap();
            }
            assert_eq!(count(store, "cached_files"), 0);
            assert_eq!(count(store, "notification_receipts"), 0);
            assert_eq!(count(store, "snapshots"), 0);
            assert_eq!(count(store, "attachments"), 1);
            assert_eq!(count(store, "operations"), 1);
            assert_eq!(count(store, "drafts"), 1);
            assert!(save(store, scope, "image/png", b"\x89PNG\r\n\x1a\n").is_err());
            let retained: String = store
                .0
                .lock()
                .unwrap()
                .query_row("SELECT id FROM attachments", [], |row| row.get(0))
                .unwrap();
            assert_eq!(retained, "pending");
            store
                .save_account(
                    &json!({"id":ACCOUNT}),
                    &json!([{"id":WORKSPACE,"slug":"acme"}]),
                )
                .unwrap();
            store.grant_workspace(ACCOUNT, WORKSPACE).unwrap();
            store.snapshot(ACCOUNT,WORKSPACE,&json!({"stories":[{"id":RESOURCE}],"documents":[],"comments":[],"teams":[],"statuses":[]})).unwrap();
            assert!(cached(store, scope).unwrap().is_none());
        });
    }
}

#[test]
fn resource_tombstones_remove_files_and_reject_late_downloads() {
    with_cached_workspace(|store, scope| {
        let document_scope = Scope {
            account: ACCOUNT.into(),
            workspace: WORKSPACE.into(),
            resource: "document".into(),
            kind: "document".into(),
            attachment: "document-media".into(),
            slug: "acme".into(),
        };
        save(store, &document_scope, "image/png", b"\x89PNG\r\n\x1a\n").unwrap();
        store.snapshot(ACCOUNT,WORKSPACE,&json!({"stories":[],"documents":[],"comments":[],"teams":[],"statuses":[],"removed":{"stories":[RESOURCE]}})).unwrap();
        assert_eq!(count(store, "cached_files"), 1);
        assert_eq!(count(store, "attachments"), 1);
        assert!(
            store.server_snapshot(ACCOUNT, WORKSPACE).unwrap()["attachments"]
                .as_array()
                .unwrap()
                .is_empty()
        );
        assert!(cached(store, scope).is_err());
        assert!(save(store, scope, "image/png", b"\x89PNG\r\n\x1a\n").is_err());
        assert!(cached(store, &document_scope).unwrap().is_some());
        store.snapshot(ACCOUNT,WORKSPACE,&json!({"stories":[],"documents":[],"comments":[],"teams":[],"statuses":[],"removed":{"documents":["document"]}})).unwrap();
        assert_eq!(count(store, "cached_files"), 0);
        assert!(save(store, &document_scope, "image/png", b"\x89PNG\r\n\x1a\n").is_err());
    });
}
