use super::credentials::{CredentialError, Credentials};
use reqwest::{header, Client, Method, Response};
use serde_json::Value;
use std::time::Duration;

pub struct Transport {
    pub client: Client,
    pub api: String,
    pub website: String,
    credentials: Credentials,
}

impl Transport {
    pub fn new(account: Option<String>) -> Result<Self, String> {
        let production = cfg!(feature = "production");
        let environment = if production {
            "production"
        } else {
            "development"
        };
        Ok(Self {
            client: Client::builder()
                .redirect(reqwest::redirect::Policy::none())
                .timeout(Duration::from_secs(30))
                .connect_timeout(Duration::from_secs(8))
                .build()
                .map_err(|e| e.to_string())?,
            api: if production {
                "https://api.fortyone.app"
            } else {
                "http://localhost:8000"
            }
            .into(),
            website: if production {
                "https://cloud.fortyone.app"
            } else {
                "http://localhost:3000"
            }
            .into(),
            credentials: Credentials::new(environment, account),
        })
    }

    pub async fn token(&self, account: &str) -> Result<String, String> {
        self.credentials
            .token(account)
            .await
            .map_err(CredentialError::message)
    }

    pub fn retry_credentials(&self, account: &str) -> Result<(), String> {
        self.credentials
            .retry_failed(account)
            .map_err(CredentialError::message)
    }

    pub async fn replace_session(&self, account: &str, token: String) -> Result<(), String> {
        self.credentials
            .replace(account, token)
            .await
            .map_err(CredentialError::message)
    }

    pub async fn remove_session(&self, account: &str) -> Result<(), String> {
        self.credentials.remove(account).await.map_err(|_| "Could not remove your secure session. Allow access to your system credential store, then try signing out again.".into())
    }

    pub fn request(
        &self,
        method: Method,
        path: &str,
        token: &str,
    ) -> Result<reqwest::RequestBuilder, String> {
        if !path.starts_with('/') || path.starts_with("//") || path.contains(['?', '#', '\\']) {
            return Err("Invalid API path".into());
        }
        if token.is_empty() || token.contains(['\r', '\n', ';']) {
            return Err("Invalid desktop session".into());
        }
        Ok(self
            .client
            .request(method, format!("{}{path}", self.api))
            .header(header::COOKIE, format!("fortyone_session={token}"))
            .header(header::ORIGIN, &self.website)
            .header(header::ACCEPT, "application/json"))
    }

    pub async fn get(&self, path: &str, token: &str) -> Result<Value, String> {
        decode(
            self.request(Method::GET, path, token)?
                .send()
                .await
                .map_err(|_| {
                    "Could not reach FortyOne. Your local work is available.".to_owned()
                })?,
        )
        .await
    }
}

pub async fn decode(mut response: Response) -> Result<Value, String> {
    let status = response.status();
    if response
        .content_length()
        .is_some_and(|len| len > 32 * 1024 * 1024)
    {
        return Err("The server response exceeds the desktop limit".into());
    }
    let mut bytes = Vec::new();
    while let Some(chunk) = response
        .chunk()
        .await
        .map_err(|_| "The response was interrupted. Retry to reconcile the change.".to_owned())?
    {
        if bytes.len() + chunk.len() > 32 * 1024 * 1024 {
            return Err("The server response exceeds the desktop limit".into());
        }
        bytes.extend_from_slice(&chunk);
    }
    if !status.is_success() {
        return Err(match status.as_u16() {
            401 => "Sign in again to synchronize. Your local changes are retained.".into(),
            403 => "Access changed. Your pending work has been retained for review.".into(),
            409 | 412 => "This item changed elsewhere. Review your changes before retrying.".into(),
            _ => format!(
                "FortyOne returned {}. Your pending work is retained.",
                status.as_u16()
            ),
        });
    }
    if bytes.is_empty() {
        return Ok(Value::Null);
    }
    let value: Value =
        serde_json::from_slice(&bytes).map_err(|_| "Invalid API response".to_owned())?;
    Ok(value.get("data").cloned().unwrap_or(value))
}
