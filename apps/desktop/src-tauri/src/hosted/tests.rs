use super::*;
use serde_json::json;

pub(super) fn scope() -> BoundScope {
    BoundScope {
        origin: "https://acme.fortyone.app".into(),
        account: "11111111-1111-4111-8111-111111111111".into(),
        workspace: "22222222-2222-4222-8222-222222222222".into(),
        slug: "acme".into(),
        cookie_hash: fingerprint("test-only-opaque-session"),
        token: "test-only-opaque-session".into(),
    }
}

fn database() -> Connection {
    let db = Connection::open_in_memory().unwrap();
    initialize(&db).unwrap();
    db
}

fn browser_cookie(attributes: &str) -> tauri::webview::Cookie<'static> {
    tauri::webview::Cookie::parse(format!("fortyone_session=test-session; {attributes}"))
        .unwrap()
        .into_owned()
}

#[test]
fn session_cookie_accepts_production_parent_domain_and_exact_api_host() {
    for domain in [
        ".fortyone.app",
        "fortyone.app",
        "api.fortyone.app",
        ".API.FORTYONE.APP",
    ] {
        let cookie = browser_cookie(&format!("Domain={domain}; Path=/; Secure; HttpOnly"));
        assert_eq!(
            select_session_cookie(vec![cookie], "https://api.fortyone.app", 1_800_000_000_000),
            Ok("test-session".into())
        );
    }
}

#[test]
fn session_cookie_rejects_sibling_suffix_confusion_and_development_domains() {
    for domain in [
        "art-circles.fortyone.app",
        ".cloud.fortyone.app",
        "evilfortyone.app",
        "fortyone.app.evil.test",
        "api.fortyone.app.evil.test",
        "app",
        "..fortyone.app",
        "localhost",
        "127.0.0.1",
    ] {
        let cookie = browser_cookie(&format!("Domain={domain}; Path=/; Secure; HttpOnly"));
        assert!(select_session_cookie(vec![cookie], "https://api.fortyone.app", 1000).is_err());
    }
    assert!(select_session_cookie(
        vec![browser_cookie("Path=/; Secure; HttpOnly")],
        "https://api.fortyone.app",
        1000
    )
    .is_err());
}

#[test]
fn session_cookie_requires_root_path_http_only_and_production_secure() {
    for attributes in [
        "Domain=.fortyone.app; Path=/auth; Secure; HttpOnly",
        "Domain=.fortyone.app; Secure; HttpOnly",
        "Domain=.fortyone.app; Path=/; Secure",
        "Domain=.fortyone.app; Path=/; HttpOnly",
    ] {
        assert!(select_session_cookie(
            vec![browser_cookie(attributes)],
            "https://api.fortyone.app",
            1000
        )
        .is_err());
    }
    let mut wrong_name = browser_cookie("Domain=.fortyone.app; Path=/; Secure; HttpOnly");
    wrong_name.set_name("other_session");
    assert!(select_session_cookie(vec![wrong_name], "https://api.fortyone.app", 1000).is_err());
}

#[test]
fn session_cookie_rejects_expiry_at_boundary_and_deleted_cookies() {
    let expired = browser_cookie(
        "Domain=.fortyone.app; Path=/; Secure; HttpOnly; Expires=Thu, 01 Jan 1970 00:00:01 GMT",
    );
    assert_eq!(
        select_session_cookie(vec![expired.clone()], "https://api.fortyone.app", 999),
        Ok("test-session".into())
    );
    for time in [1000, 1001] {
        assert!(
            select_session_cookie(vec![expired.clone()], "https://api.fortyone.app", time).is_err()
        );
    }
    for age in [0, -1] {
        let deleted = browser_cookie(&format!(
            "Domain=.fortyone.app; Path=/; Secure; HttpOnly; Max-Age={age}"
        ));
        assert!(select_session_cookie(vec![deleted], "https://api.fortyone.app", 1000).is_err());
    }
}

#[test]
fn session_cookie_rejects_ambiguous_sessions_and_invalid_values() {
    let parent = browser_cookie("Domain=.fortyone.app; Path=/; Secure; HttpOnly");
    let host = browser_cookie("Domain=api.fortyone.app; Path=/; Secure; HttpOnly");
    assert_eq!(
        select_session_cookie(vec![parent.clone(), host], "https://api.fortyone.app", 1000),
        Err("The current browser session is invalid".into())
    );
    for value in ["", "bad\rvalue", "bad\nvalue", "bad;value"] {
        let mut invalid = parent.clone();
        invalid.set_value(value);
        assert_eq!(
            select_session_cookie(vec![invalid], "https://api.fortyone.app", 1000),
            Err("The current browser session is invalid".into())
        );
    }
    let unrelated = browser_cookie("Domain=unrelated.test; Path=/; Secure; HttpOnly");
    assert_eq!(
        select_session_cookie(vec![unrelated, parent], "https://api.fortyone.app", 1000),
        Ok("test-session".into())
    );
}

#[test]
fn session_cookie_keeps_loopback_development_separate_from_production() {
    let local = browser_cookie("Domain=localhost; Path=/; HttpOnly");
    assert_eq!(
        select_session_cookie(vec![local.clone()], "http://localhost:8000", 1000),
        Ok("test-session".into())
    );
    assert!(select_session_cookie(vec![local], "https://api.fortyone.app", 1000).is_err());
    let production = browser_cookie("Domain=.fortyone.app; Path=/; Secure; HttpOnly");
    assert!(
        select_session_cookie(vec![production.clone()], "http://localhost:8000", 1000).is_err()
    );
    for api in [
        "https://evil.test",
        "http://api.fortyone.app",
        "http://localhost:9000",
    ] {
        assert!(select_session_cookie(vec![production.clone()], api, 1000).is_err());
    }
}

fn seed(db: &Connection, value: &BoundScope, verified: u64) {
    db.execute(
        "INSERT INTO hosted_active VALUES(?,?,?)",
        params![value.origin, value.account, value.cookie_hash],
    )
    .unwrap();
    db.execute(
        "INSERT INTO hosted_bindings VALUES(?,?,?,?,?,?)",
        params![
            value.origin,
            value.account,
            value.workspace,
            value.slug,
            value.cookie_hash,
            verified
        ],
    )
    .unwrap();
}

#[test]
fn binding_requires_active_account_workspace_origin_cookie_and_live_lease() {
    let db = database();
    let value = scope();
    let time = LEASE_MS * 2;
    seed(&db, &value, time);
    assert!(binding(
        &db,
        &value.origin,
        &value.account,
        &value.workspace,
        &value.cookie_hash,
        time
    )
    .is_ok());
    for (origin, account, workspace, hash) in [
        (
            "https://other.fortyone.app",
            value.account.as_str(),
            value.workspace.as_str(),
            value.cookie_hash.as_str(),
        ),
        (
            value.origin.as_str(),
            "other",
            value.workspace.as_str(),
            value.cookie_hash.as_str(),
        ),
        (
            value.origin.as_str(),
            value.account.as_str(),
            "other",
            value.cookie_hash.as_str(),
        ),
        (
            value.origin.as_str(),
            value.account.as_str(),
            value.workspace.as_str(),
            "changed-cookie",
        ),
    ] {
        assert!(binding(&db, origin, account, workspace, hash, time).is_err());
    }
    assert!(binding(
        &db,
        &value.origin,
        &value.account,
        &value.workspace,
        &value.cookie_hash,
        time + LEASE_MS + 1
    )
    .is_err());
    assert!(binding(
        &db,
        &value.origin,
        &value.account,
        &value.workspace,
        &value.cookie_hash,
        time - 60_001
    )
    .is_err());
    db.execute("UPDATE hosted_active SET account='another-account'", [])
        .unwrap();
    assert!(binding(
        &db,
        &value.origin,
        &value.account,
        &value.workspace,
        &value.cookie_hash,
        time
    )
    .is_err());
}

#[test]
fn origin_scope_rejects_other_workspace_and_infrastructure() {
    let policy = crate::navigation::NavigationPolicy::production();
    for accepted in [
        "https://acme.fortyone.app/my-work",
        "https://acme.fortyone.app/settings/preferences",
    ] {
        assert!(origin_scope(
            &policy,
            &Url::parse(accepted).unwrap(),
            "acme"
        ));
    }
    for rejected in [
        "https://other.fortyone.app/my-work",
        "https://cloud.fortyone.app/acme/my-work",
        "https://api.fortyone.app/",
        "http://acme.fortyone.app/",
        "https://acme.fortyone.app:444/",
        "https://acme.fortyone.app.evil.test/",
        "https://user:pass@acme.fortyone.app/",
    ] {
        assert!(
            !origin_scope(&policy, &Url::parse(rejected).unwrap(), "acme"),
            "{rejected}"
        );
    }
    assert!(!origin_scope(
        &policy,
        &Url::parse("https://api.fortyone.app/").unwrap(),
        "api"
    ));
}

#[test]
fn development_scope_requires_exact_origin_and_workspace_route() {
    let policy =
        crate::navigation::NavigationPolicy::new(Url::parse("http://localhost:3000/").unwrap())
            .unwrap();
    assert!(origin_scope(
        &policy,
        &Url::parse("http://localhost:3000/acme/my-work").unwrap(),
        "acme"
    ));
    for rejected in [
        "http://localhost:3000/other/my-work",
        "http://localhost:3001/acme/my-work",
        "http://127.0.0.1:3000/acme/my-work",
        "http://localhost:3000/acme-other",
    ] {
        assert!(!origin_scope(
            &policy,
            &Url::parse(rejected).unwrap(),
            "acme"
        ));
    }
}

#[test]
fn forget_revokes_cached_reads_but_retains_unsent_draft_text() {
    let db = database();
    let value = scope();
    seed(&db, &value, now());
    db.execute(
        "INSERT INTO hosted_cache VALUES(?,?,?,1,'{}',?)",
        params![value.origin, value.account, value.workspace, now()],
    )
    .unwrap();
    db.execute(
        "INSERT INTO hosted_drafts VALUES(?,?,?,'story:local','{}',?)",
        params![value.origin, value.account, value.workspace, now()],
    )
    .unwrap();
    assert!(ensure_scope_current(&db, &value).is_ok());
    forget_origin(&db, &value.origin).unwrap();
    assert!(ensure_scope_current(&db, &value).is_err());
    assert_eq!(
        db.query_row("SELECT COUNT(*) FROM hosted_cache", [], |row| row
            .get::<_, u64>(0))
            .unwrap(),
        0
    );
    assert_eq!(
        db.query_row("SELECT COUNT(*) FROM hosted_drafts", [], |row| row
            .get::<_, u64>(0))
            .unwrap(),
        1
    );
}

#[test]
fn cache_payload_identity_and_shape_are_validated_before_storage() {
    let value = scope();
    let payload = json!({"version":1,"accountId":value.account,"workspaceId":value.workspace,"workspaceSlug":value.slug,"savedAt":now(),"queries":[{"queryKey":["stories"],"data":[],"updatedAt":now()}]});
    assert!(cache::validate_for_test(&payload, &value).is_ok());
    for field in ["accountId", "workspaceId", "workspaceSlug"] {
        let mut other = payload.clone();
        other[field] = json!("other");
        assert!(cache::validate_for_test(&other, &value).is_err());
    }
    let mut invalid = payload.clone();
    invalid["queries"][0]["queryKey"] = json!("not-an-array");
    assert!(cache::validate_for_test(&invalid, &value).is_err());
    let mut oversized = payload;
    oversized["queries"] = json!(vec![json!({"queryKey":[],"data":null,"updatedAt":0}); 2001]);
    assert!(cache::validate_for_test(&oversized, &value).is_err());
}

#[test]
fn sqlite_contains_hash_only_and_binding_does_not_renew_when_read() {
    let db = database();
    let value = scope();
    let verified = now();
    seed(&db, &value, verified);
    ensure_scope_current(&db, &value).unwrap();
    let (stored, after): (String, u64) = db
        .query_row(
            "SELECT cookie_hash,verified FROM hosted_bindings",
            [],
            |row| Ok((row.get(0)?, row.get(1)?)),
        )
        .unwrap();
    assert_ne!(stored, value.token);
    assert_eq!(stored.len(), 64);
    assert_eq!(after, verified);
}
