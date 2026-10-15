//! Downloaded inline media is scoped to the account, workspace, and resource.
//! Browser credentials and signed object URLs never cross the renderer bridge.
mod export;
use super::{model::now, runtime, store::Store, Runtime};
use base64::{engine::general_purpose::STANDARD, Engine};
use reqwest::{header, Method};
use rusqlite::{params, OptionalExtension};
use serde::Serialize;
use serde_json::json;
use std::{net::IpAddr, time::Duration};
use tauri::{Url, WebviewWindow};

const MAX_MEDIA_BYTES: usize = 10 * 1024 * 1024;
const MAX_ACCOUNT_CACHE_BYTES: usize = 256 * 1024 * 1024;
static DOWNLOADS: tokio::sync::Semaphore = tokio::sync::Semaphore::const_new(4);

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct MediaData {
    mime_type: String,
    data: String,
}

struct Scope {
    account: String,
    workspace: String,
    resource: String,
    kind: String,
    attachment: String,
    slug: String,
}
impl Scope {
    fn new(
        runtime: &Runtime,
        expected_account_id: &str,
        workspace: String,
        resource: String,
        kind: String,
        attachment: String,
    ) -> Result<Self, String> {
        for id in [&workspace, &resource, &attachment] {
            uuid::Uuid::parse_str(id).map_err(|_| "Invalid media identifier")?;
        }
        if !matches!(kind.as_str(), "story" | "document") {
            return Err("Unsupported media resource".into());
        }
        let account = super::require_account(runtime, expected_account_id)?;
        let membership = runtime.store.authorize(&account, &workspace)?;
        let slug = membership["slug"]
            .as_str()
            .filter(|value| {
                !value.is_empty()
                    && value.len() <= 120
                    && value
                        .bytes()
                        .all(|byte| byte.is_ascii_alphanumeric() || byte == b'-')
            })
            .ok_or("Invalid workspace address")?
            .to_owned();
        Ok(Self {
            account,
            workspace,
            resource,
            kind,
            attachment,
            slug,
        })
    }
    fn path(&self) -> String {
        let collection = if self.kind == "story" {
            "stories"
        } else {
            "documents"
        };
        format!(
            "/workspaces/{}/{}/{}/media/{}",
            self.slug, collection, self.resource, self.attachment
        )
    }
}

pub fn initialize(store: &Store) -> Result<(), String> {
    store.0.lock().map_err(|error|error.to_string())?.execute_batch(
        "CREATE TABLE IF NOT EXISTS cached_files(account TEXT NOT NULL,workspace TEXT NOT NULL,resource TEXT NOT NULL,kind TEXT NOT NULL,attachment TEXT NOT NULL,mime TEXT NOT NULL,name TEXT NOT NULL,bytes BLOB NOT NULL,saved INTEGER NOT NULL,PRIMARY KEY(account,workspace,resource,kind,attachment));"
    ).map_err(|error|error.to_string())
}

fn verify_scope(db: &rusqlite::Connection, scope: &Scope) -> Result<(), String> {
    let earliest = now().saturating_sub(7 * 24 * 3600);
    let valid:bool=db.query_row(
        "SELECT EXISTS(SELECT 1 FROM accounts WHERE id=? AND verified>=? AND EXISTS(SELECT 1 FROM json_each(accounts.workspaces) WHERE json_extract(value,'$.id')=?)) AND EXISTS(SELECT 1 FROM settings WHERE key='account' AND value=?) AND NOT EXISTS(SELECT 1 FROM workspace_access WHERE account=? AND workspace=? AND (denied IS NOT NULL OR verified<?))",
        params![scope.account,earliest,scope.workspace,json!(scope.account).to_string(),scope.account,scope.workspace,earliest],|row|row.get(0)
    ).map_err(|error|error.to_string())?;
    if !valid {
        return Err("Account or workspace access changed".into());
    }
    // Check inside the same transaction as cache writes. A late download must
    // not recreate bytes after a resource tombstone removed its snapshot.
    let collection = if scope.kind == "story" {
        "$.stories"
    } else {
        "$.documents"
    };
    let accessible: bool = db.query_row(
        "SELECT EXISTS(SELECT 1 FROM snapshots,json_each(snapshots.data,?1) WHERE snapshots.account=?2 AND snapshots.workspace=?3 AND json_extract(value,'$.id')=?4) OR EXISTS(SELECT 1 FROM operations WHERE account=?2 AND workspace=?3 AND json_extract(operation,'$.kind')=?5 AND json_extract(operation,'$.resourceId')=?4)",
        params![collection,scope.account,scope.workspace,scope.resource,format!("{}.create",scope.kind)],|row|row.get(0)
    ).map_err(|error|error.to_string())?;
    if !accessible {
        return Err("This resource is not available in the downloaded workspace".into());
    }
    Ok(())
}

fn cached(store: &Store, scope: &Scope) -> Result<Option<MediaData>, String> {
    store.authorize(&scope.account, &scope.workspace)?;
    let db = store.0.lock().map_err(|error| error.to_string())?;
    verify_scope(&db, scope)?;
    let row:Option<(String,Vec<u8>)>=db.query_row(
        "SELECT mime,bytes FROM cached_files WHERE account=? AND workspace=? AND resource=? AND kind=? AND attachment=?",
        params![scope.account,scope.workspace,scope.resource,scope.kind,scope.attachment],|row|Ok((row.get(0)?,row.get(1)?))
    ).optional().map_err(|error|error.to_string())?;
    row.map(|(mime_type, bytes)| {
        if bytes.len() > MAX_MEDIA_BYTES || !matches_content(&mime_type, &bytes) {
            return Err("This saved media needs to be downloaded again".into());
        }
        Ok(MediaData {
            mime_type,
            data: STANDARD.encode(bytes),
        })
    })
    .transpose()
}

fn matches_content(mime: &str, bytes: &[u8]) -> bool {
    match mime {
        "image/png" => bytes.starts_with(b"\x89PNG\r\n\x1a\n"),
        "image/jpeg" => bytes.starts_with(&[0xff, 0xd8, 0xff]),
        "image/gif" => bytes.starts_with(b"GIF87a") || bytes.starts_with(b"GIF89a"),
        "image/webp" => bytes.len() >= 12 && &bytes[..4] == b"RIFF" && &bytes[8..12] == b"WEBP",
        "video/mp4" => bytes.len() >= 12 && &bytes[4..8] == b"ftyp" && &bytes[8..12] != b"qt  ",
        _ => false,
    }
}

fn blob_url(value: &str) -> Result<Url, String> {
    let url = Url::parse(value).map_err(|_| "Invalid media download address")?;
    let host = url.host_str().ok_or("Invalid media download address")?;
    if url.scheme() != "https"
        || !url.username().is_empty()
        || url.password().is_some()
        || url.fragment().is_some()
        || url.port().is_some_and(|port| port != 443)
        || host == "localhost"
        || host.ends_with(".localhost")
        || host.ends_with(".local")
    {
        return Err("Media downloads require a secure storage address".into());
    }
    if let Ok(ip) = host.trim_matches(['[', ']']).parse::<IpAddr>() {
        let blocked = match ip {
            IpAddr::V4(ip) => {
                ip.is_private()
                    || ip.is_loopback()
                    || ip.is_link_local()
                    || ip.is_unspecified()
                    || ip.is_broadcast()
                    || ip.is_multicast()
            }
            IpAddr::V6(ip) => {
                ip.is_loopback()
                    || ip.is_unspecified()
                    || ip.is_unique_local()
                    || ip.is_unicast_link_local()
                    || ip.is_multicast()
                    || ip.to_ipv4_mapped().is_some()
            }
        };
        if blocked {
            return Err("Media storage must use a public HTTPS address".into());
        }
    }
    Ok(url)
}

fn evict(store: &Store, scope: &Scope) -> Result<(), String> {
    store.0.lock().map_err(|error|error.to_string())?.execute(
        "DELETE FROM cached_files WHERE account=? AND workspace=? AND resource=? AND kind=? AND attachment=?",
        params![scope.account,scope.workspace,scope.resource,scope.kind,scope.attachment]
    ).map_err(|error|error.to_string())?;
    Ok(())
}

fn save(store: &Store, scope: &Scope, mime: &str, bytes: &[u8]) -> Result<(), String> {
    if bytes.len() > MAX_MEDIA_BYTES || !matches_content(mime, bytes) {
        return Err("Unsupported or oversized inline media".into());
    }
    store.authorize(&scope.account, &scope.workspace)?;
    let mut db = store.0.lock().map_err(|error| error.to_string())?;
    let tx = db.transaction().map_err(|error| error.to_string())?;
    verify_scope(&tx, scope)?;
    tx.execute("DELETE FROM cached_files WHERE account=? AND workspace=? AND resource=? AND kind=? AND attachment=?",params![scope.account,scope.workspace,scope.resource,scope.kind,scope.attachment]).map_err(|error|error.to_string())?;
    let total: usize = tx
        .query_row(
            "SELECT COALESCE(sum(length(bytes)),0) FROM cached_files WHERE account=?",
            [&scope.account],
            |row| row.get(0),
        )
        .map_err(|error| error.to_string())?;
    let excess = (total + bytes.len()).saturating_sub(MAX_ACCOUNT_CACHE_BYTES);
    if excess > 0 {
        // Evict the oldest downloaded files only. Unsynced attachments and
        // document drafts live in different tables and are never evicted.
        tx.execute("DELETE FROM cached_files WHERE rowid IN (SELECT rowid FROM (SELECT rowid,sum(length(bytes)) OVER (ORDER BY saved,rowid ROWS UNBOUNDED PRECEDING)-length(bytes) AS preceding_bytes FROM cached_files WHERE account=?) WHERE preceding_bytes<?)",params![scope.account,excess]).map_err(|error|error.to_string())?;
    }
    let extension = match mime {
        "image/png" => "png",
        "image/jpeg" => "jpg",
        "image/gif" => "gif",
        "image/webp" => "webp",
        _ => "mp4",
    };
    tx.execute(
        "INSERT INTO cached_files VALUES(?,?,?,?,?,?,?,?,?)",
        params![
            scope.account,
            scope.workspace,
            scope.resource,
            scope.kind,
            scope.attachment,
            mime,
            format!("{}.{}", scope.attachment, extension),
            bytes,
            now()
        ],
    )
    .map_err(|_| "Could not save this media. Check free disk space.".to_owned())?;
    tx.commit().map_err(|error| error.to_string())
}

#[tauri::command]
pub fn desktop_read_media(
    window: WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
    resource_id: String,
    kind: String,
    attachment_id: String,
) -> Result<Option<MediaData>, String> {
    let runtime = runtime(&window)?;
    let scope = Scope::new(
        &runtime,
        &expected_account_id,
        workspace_id,
        resource_id,
        kind,
        attachment_id,
    )?;
    cached(&runtime.store, &scope)
}

#[tauri::command]
pub async fn desktop_cache_media(
    window: WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
    resource_id: String,
    kind: String,
    attachment_id: String,
) -> Result<MediaData, String> {
    let runtime = runtime(&window)?;
    let scope = Scope::new(
        &runtime,
        &expected_account_id,
        workspace_id,
        resource_id,
        kind,
        attachment_id,
    )?;
    let _permit = DOWNLOADS
        .acquire()
        .await
        .map_err(|_| "The media downloader is unavailable")?;
    if runtime.store.account()? != scope.account {
        return Err("Your account changed before the media download".into());
    }
    runtime.store.authorize(&scope.account, &scope.workspace)?;
    let token = runtime.transport.token(&scope.account).await?;
    let response = runtime
        .transport
        .request(Method::GET, &scope.path(), &token)?
        .send()
        .await
        .map_err(|_| "Connect to download this media for offline use")?;
    if matches!(response.status().as_u16(), 401 | 403 | 404) {
        evict(&runtime.store, &scope)?;
        return Err("This media is no longer available to your account".into());
    }
    if !matches!(response.status().as_u16(), 302 | 303 | 307 | 308) {
        return Err("The server could not authorize this media download".into());
    }
    let url = blob_url(
        response
            .headers()
            .get(header::LOCATION)
            .and_then(|value| value.to_str().ok())
            .ok_or("Missing media download address")?,
    )?;
    let (mime, bytes) = fetch_blob(url).await?;
    if runtime.store.account()? != scope.account {
        return Err("Your account changed during the media download".into());
    }
    save(&runtime.store, &scope, &mime, &bytes)?;
    Ok(MediaData {
        mime_type: mime,
        data: STANDARD.encode(bytes),
    })
}

pub(super) async fn fetch_blob(url: Url) -> Result<(String, Vec<u8>), String> {
    // A separate client has no cookie jar, session headers, Referer, or redirect
    // following. An expiring storage URL is kept in this Rust stack frame only.
    let client = reqwest::Client::builder()
        .redirect(reqwest::redirect::Policy::none())
        .connect_timeout(Duration::from_secs(8))
        .timeout(Duration::from_secs(60))
        .build()
        .map_err(|_| "Could not start the media downloader")?;
    let mut response = client
        .get(url)
        .send()
        .await
        .map_err(|_| "The media download was interrupted")?;
    if !response.status().is_success() {
        return Err("The media download expired. Try again while connected.".into());
    }
    let mime = response
        .headers()
        .get(header::CONTENT_TYPE)
        .and_then(|value| value.to_str().ok())
        .unwrap_or_default()
        .split(';')
        .next()
        .unwrap_or_default()
        .trim()
        .to_ascii_lowercase();
    if response
        .content_length()
        .is_some_and(|length| length > MAX_MEDIA_BYTES as u64)
    {
        return Err("Only media up to 10 MB can be saved offline".into());
    }
    let mut bytes = Vec::new();
    while let Some(chunk) = response
        .chunk()
        .await
        .map_err(|_| "The media download was interrupted")?
    {
        if bytes.len() + chunk.len() > MAX_MEDIA_BYTES {
            return Err("Only media up to 10 MB can be saved offline".into());
        }
        bytes.extend_from_slice(&chunk);
    }
    Ok((mime, bytes))
}

#[tauri::command]
pub async fn desktop_export_attachment(
    window: WebviewWindow,
    expected_account_id: String,
    workspace_id: String,
    resource_id: String,
    kind: String,
    attachment_id: String,
) -> Result<bool, String> {
    export::export_attachment(
        window,
        expected_account_id,
        workspace_id,
        resource_id,
        kind,
        attachment_id,
    )
    .await
}

#[cfg(test)]
mod tests;
