//! Private Next HTML stays behind the same native account/workspace boundary as
//! query data. The service worker retains public assets and ephemeral permits only.
use rusqlite::{params, Connection, OptionalExtension};
use serde::{Deserialize, Serialize};
use tauri::{Url, WebviewWindow};

use super::{authorize_scope, authorize_window, ensure_scope_current, now, BoundScope};

const MAX_PAGE_BYTES: usize = 4 * 1024 * 1024;
const MAX_ACCOUNT_BYTES: u64 = 64 * 1024 * 1024;
const MAX_PAGES: u32 = 24;
const MAX_AGE: u64 = 7 * 24 * 60 * 60 * 1000;

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct PageRead {
    account_id: String,
    workspace_id: String,
    url: String,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct PageWrite {
    account_id: String,
    workspace_id: String,
    url: String,
    html: String,
    content_security_policy: Option<String>,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct SavedPage {
    html: String,
    saved_at: u64,
    #[serde(skip_serializing_if = "Option::is_none")]
    content_security_policy: Option<String>,
}

pub fn initialize(db: &Connection) -> Result<(), String> {
    db.execute_batch("CREATE TABLE IF NOT EXISTS hosted_pages(origin TEXT NOT NULL,account TEXT NOT NULL,workspace TEXT NOT NULL,url TEXT NOT NULL,html TEXT NOT NULL,saved INTEGER NOT NULL,content_security_policy TEXT,PRIMARY KEY(origin,account,workspace,url));")
        .map_err(|error| error.to_string())?;
    let has_policy: bool = db.query_row("SELECT EXISTS(SELECT 1 FROM pragma_table_info('hosted_pages') WHERE name='content_security_policy')",[],|row|row.get(0)).map_err(|error|error.to_string())?;
    if !has_policy {
        db.execute_batch("ALTER TABLE hosted_pages ADD COLUMN content_security_policy TEXT")
            .map_err(|error| error.to_string())?;
    }
    Ok(())
}

fn canonical_url(value: &str, scope: &BoundScope, production: bool) -> Result<String, String> {
    if value.len() > 4096 {
        return Err("This page address is too long".into());
    }
    let mut url = Url::parse(value).map_err(|_| "Invalid offline page address")?;
    if url.origin().ascii_serialization() != scope.origin
        || !url.username().is_empty()
        || url.password().is_some()
        || url.fragment().is_some()
    {
        return Err("The offline page must belong to this workspace".into());
    }
    let mut parts = url.path().split('/').skip(1);
    if !production && parts.next() != Some(scope.slug.as_str()) {
        return Err("The offline page must belong to this workspace".into());
    }
    let section = parts.next();
    if !matches!(
        section,
        Some(
            "maya"
                | "my-work"
                | "summary"
                | "work"
                | "story"
                | "docs"
                | "teams"
                | "sprints"
                | "notifications"
                | "settings"
                | "profile"
                | "views"
                | "search"
                | "calendar"
                | "strategy"
                | "roadmap"
                | "analytics"
                | "key-results"
        )
    ) {
        return Err("This page is not available for offline startup".into());
    }
    if section == Some("settings") {
        let remainder = parts.collect::<Vec<_>>().join("/");
        if !matches!(
            remainder.trim_end_matches('/'),
            "" | "account" | "account/preferences" | "account/notifications" | "workspace/members"
        ) {
            return Err("This settings page cannot be saved for offline startup".into());
        }
    }
    let mut retained = Vec::new();
    for (key, value) in url.query_pairs() {
        if matches!(key.as_ref(), "_rsc" | "__fortyone_offline") {
            continue;
        }
        let normalized = key.to_ascii_lowercase();
        if [
            "token",
            "secret",
            "password",
            "code",
            "callback",
            "authorization",
        ]
        .iter()
        .any(|part| normalized.contains(part))
        {
            return Err("Pages containing authorization parameters cannot be saved".into());
        }
        retained.push((key.into_owned(), value.into_owned()));
    }
    url.set_query(None);
    if !retained.is_empty() {
        url.query_pairs_mut().extend_pairs(retained);
    }
    Ok(url.into())
}

fn save(
    db: &mut Connection,
    scope: &BoundScope,
    url: &str,
    html: &str,
    time: u64,
    content_security_policy: Option<&str>,
) -> Result<(), String> {
    if html.len() > MAX_PAGE_BYTES
        || !html.contains("/_next/")
        || html.contains("NEXT_REDIRECT")
        || html.contains("NEXT_HTTP_ERROR_FALLBACK")
    {
        return Err("This page is not ready for offline startup".into());
    }
    if content_security_policy
        .is_some_and(|policy| policy.len() > 8192 || policy.contains(['\r', '\n']))
    {
        return Err("The page security policy cannot be saved".into());
    }
    let tx = db.transaction().map_err(|error| error.to_string())?;
    tx.execute(
        "DELETE FROM hosted_pages WHERE origin=? AND account=? AND saved<?",
        params![scope.origin, scope.account, time.saturating_sub(MAX_AGE)],
    )
    .map_err(|error| error.to_string())?;
    tx.execute("INSERT INTO hosted_pages(origin,account,workspace,url,html,saved,content_security_policy) VALUES(?,?,?,?,?,?,?) ON CONFLICT(origin,account,workspace,url) DO UPDATE SET html=excluded.html,saved=excluded.saved,content_security_policy=excluded.content_security_policy",params![scope.origin,scope.account,scope.workspace,url,html,time,content_security_policy]).map_err(|error| error.to_string())?;
    tx.execute("DELETE FROM hosted_pages WHERE rowid IN (SELECT rowid FROM hosted_pages WHERE origin=? AND account=? AND workspace=? ORDER BY saved DESC,rowid DESC LIMIT -1 OFFSET ?)",params![scope.origin,scope.account,scope.workspace,MAX_PAGES]).map_err(|error| error.to_string())?;
    loop {
        let bytes: u64 = tx.query_row("SELECT COALESCE(SUM(LENGTH(CAST(html AS BLOB))+COALESCE(LENGTH(CAST(content_security_policy AS BLOB)),0)),0) FROM hosted_pages WHERE origin=? AND account=?",params![scope.origin,scope.account],|row|row.get(0)).map_err(|error|error.to_string())?;
        if bytes <= MAX_ACCOUNT_BYTES {
            break;
        }
        tx.execute("DELETE FROM hosted_pages WHERE rowid=(SELECT rowid FROM hosted_pages WHERE origin=? AND account=? ORDER BY saved,rowid LIMIT 1)",params![scope.origin,scope.account]).map_err(|error|error.to_string())?;
    }
    tx.commit().map_err(|error| error.to_string())
}

fn read(
    db: &Connection,
    scope: &BoundScope,
    url: &str,
    time: u64,
) -> Result<Option<SavedPage>, String> {
    db.query_row("SELECT html,saved,content_security_policy FROM hosted_pages WHERE origin=? AND account=? AND workspace=? AND url=? AND saved>=? AND saved<=?",params![scope.origin,scope.account,scope.workspace,url,time.saturating_sub(MAX_AGE),time.saturating_add(60_000)],|row|Ok(SavedPage {html:row.get(0)?,saved_at:row.get(1)?,content_security_policy:row.get(2)?})).optional().map_err(|error|error.to_string())
}

#[tauri::command]
pub async fn hosted_page_write(window: WebviewWindow, request: PageWrite) -> Result<(), String> {
    let state = authorize_window(&window)?;
    let scope = authorize_scope(&window, &request.account_id, &request.workspace_id, None).await?;
    let url = canonical_url(&request.url, &scope, cfg!(feature = "production"))?;
    let mut db = state.db.lock().map_err(|error| error.to_string())?;
    ensure_scope_current(&db, &scope)?;
    save(
        &mut db,
        &scope,
        &url,
        &request.html,
        now(),
        request.content_security_policy.as_deref(),
    )
}

#[tauri::command]
pub async fn hosted_page_read(
    window: WebviewWindow,
    request: PageRead,
) -> Result<Option<SavedPage>, String> {
    let state = authorize_window(&window)?;
    let scope = authorize_scope(&window, &request.account_id, &request.workspace_id, None).await?;
    let url = canonical_url(&request.url, &scope, cfg!(feature = "production"))?;
    let db = state.db.lock().map_err(|error| error.to_string())?;
    ensure_scope_current(&db, &scope)?;
    read(&db, &scope, &url, now())
}

#[cfg(test)]
mod tests {
    use super::*;
    fn bound_scope() -> BoundScope {
        BoundScope {
            origin: "http://localhost:3000".into(),
            account: "account".into(),
            workspace: "workspace".into(),
            slug: "engineering".into(),
            cookie_hash: "hash".into(),
            token: String::new(),
        }
    }
    fn page() -> &'static str {
        "<html><head><script src=\"/_next/static/app.js\"></script></head><body>Workspace</body></html>"
    }
    #[test]
    fn page_addresses_stay_inside_the_bound_workspace() {
        let scope = bound_scope();
        for url in [
            "https://outside.example/engineering/my-work",
            "http://localhost:3000/product/my-work",
            "http://localhost:3000/engineering/auth/desktop",
            "http://localhost:3000/engineering/settings?token=secret",
            "http://user@localhost:3000/engineering/my-work",
            "http://localhost:3000/engineering/my-work#fragment",
        ] {
            assert!(canonical_url(url, &scope, false).is_err(), "{url}");
        }
        assert_eq!(canonical_url("http://localhost:3000/engineering/my-work?tab=all&_rsc=nonce&__fortyone_offline=nonce",&scope,false).unwrap(),"http://localhost:3000/engineering/my-work?tab=all");
    }
    #[test]
    fn production_pages_use_the_workspace_origin_without_a_slug_prefix() {
        let mut scope = bound_scope();
        scope.origin = "https://engineering.fortyone.app".into();
        assert!(canonical_url(
            "https://engineering.fortyone.app/settings/account/preferences",
            &scope,
            true
        )
        .is_ok());
        assert!(canonical_url("https://product.fortyone.app/my-work", &scope, true).is_err());
    }
    #[test]
    fn only_explicit_non_secret_settings_pages_can_be_retained() {
        let scope = bound_scope();
        for page in [
            "settings",
            "settings/account",
            "settings/account/preferences",
            "settings/account/notifications",
            "settings/workspace/members",
        ] {
            assert!(
                canonical_url(
                    &format!("http://localhost:3000/engineering/{page}"),
                    &scope,
                    false
                )
                .is_ok(),
                "{page}"
            );
        }
        for page in [
            "settings/security",
            "settings/account/security",
            "settings/workspace/developer",
            "settings/workspace/api-keys",
            "settings/workspace/billing",
            "settings/workspace/integrations",
            "settings/account/unknown",
            "settings/account/preferences/secret",
            "settings/account/%73ecurity",
        ] {
            assert!(
                canonical_url(
                    &format!("http://localhost:3000/engineering/{page}"),
                    &scope,
                    false
                )
                .is_err(),
                "{page}"
            );
        }
    }
    #[test]
    fn saved_pages_are_account_workspace_scoped_and_expire() {
        let mut db = Connection::open_in_memory().unwrap();
        initialize(&db).unwrap();
        let scope = bound_scope();
        let url = "http://localhost:3000/engineering/my-work";
        save(&mut db, &scope, url, page(), MAX_AGE, None).unwrap();
        assert_eq!(
            read(&db, &scope, url, MAX_AGE).unwrap().unwrap().html,
            page()
        );
        assert!(read(&db, &scope, url, MAX_AGE * 2 + 1).unwrap().is_none());
        let mut other = bound_scope();
        other.account = "other".into();
        assert!(read(&db, &other, url, MAX_AGE).unwrap().is_none());
        other.account = "account".into();
        other.workspace = "other-workspace".into();
        assert!(read(&db, &other, url, MAX_AGE).unwrap().is_none());
    }
    #[test]
    fn page_count_is_bounded_and_redirect_html_is_not_saved() {
        let mut db = Connection::open_in_memory().unwrap();
        initialize(&db).unwrap();
        let scope = bound_scope();
        for index in 0..30 {
            save(
                &mut db,
                &scope,
                &format!("page-{index}"),
                page(),
                MAX_AGE + index,
                None,
            )
            .unwrap();
        }
        assert_eq!(
            db.query_row("SELECT COUNT(*) FROM hosted_pages", [], |row| row
                .get::<_, u32>(0))
                .unwrap(),
            24
        );
        assert!(read(&db, &scope, "page-0", MAX_AGE + 30).unwrap().is_none());
        assert!(save(
            &mut db,
            &scope,
            "redirect",
            "/_next/ NEXT_REDIRECT",
            MAX_AGE,
            None
        )
        .is_err());
        assert!(save(
            &mut db,
            &scope,
            "large",
            &format!("/_next/{}", "x".repeat(MAX_PAGE_BYTES)),
            MAX_AGE,
            None
        )
        .is_err());
    }
    #[test]
    fn preserves_bounded_security_policy_without_header_injection() {
        let mut db = Connection::open_in_memory().unwrap();
        initialize(&db).unwrap();
        let scope = bound_scope();
        save(
            &mut db,
            &scope,
            "page",
            page(),
            MAX_AGE,
            Some("default-src 'self'; script-src 'self' 'nonce-value'"),
        )
        .unwrap();
        assert_eq!(
            read(&db, &scope, "page", MAX_AGE)
                .unwrap()
                .unwrap()
                .content_security_policy
                .as_deref(),
            Some("default-src 'self'; script-src 'self' 'nonce-value'")
        );
        assert!(save(
            &mut db,
            &scope,
            "other",
            page(),
            MAX_AGE,
            Some("default-src 'self'\r\nX-Header: injected")
        )
        .is_err());
        assert!(save(
            &mut db,
            &scope,
            "other",
            page(),
            MAX_AGE,
            Some(&"x".repeat(8193))
        )
        .is_err());
    }
}
