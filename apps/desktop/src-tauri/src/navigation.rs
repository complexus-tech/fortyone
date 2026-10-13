use tauri::Url;

/// The local shell never needs to navigate to an API or identity provider.
/// Those requests either use fetch or open in the user's browser.
pub struct NavigationPolicy {
    app_url: Url,
}

impl NavigationPolicy {
    pub fn new(app_url: Url) -> Result<Self, &'static str> {
        let is_loopback = matches!(
            app_url.host_str(),
            Some("localhost" | "127.0.0.1" | "[::1]")
        );
        if !matches!(app_url.scheme(), "http" | "https")
            || !is_loopback
            || has_credentials(&app_url)
            || app_url.fragment().is_some()
        {
            return Err("FORTYONE_DESKTOP_URL must be a loopback HTTP(S) URL without credentials or a fragment");
        }

        Ok(Self { app_url })
    }

    pub fn app_url(&self) -> &Url {
        &self.app_url
    }

    pub fn allows_navigation(&self, url: &Url) -> bool {
        !has_credentials(url) && url.origin() == self.app_url.origin()
    }
}

pub fn can_open_externally(url: &Url) -> bool {
    !has_credentials(url) && matches!(url.scheme(), "https" | "http" | "mailto")
}

fn has_credentials(url: &Url) -> bool {
    !url.username().is_empty() || url.password().is_some()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn keeps_local_routes_and_fragments_inside_the_window() {
        let policy = NavigationPolicy::new(Url::parse("http://localhost:3000").unwrap()).unwrap();
        for target in [
            "http://localhost:3000/acme/maya",
            "http://localhost:3000/verify?callbackUrl=%2Facme%2Fmaya",
            "http://localhost:3000/acme/docs#heading",
        ] {
            assert!(policy.allows_navigation(&Url::parse(target).unwrap()));
        }
    }

    #[test]
    fn rejects_other_origins_and_embedded_credentials() {
        let policy = NavigationPolicy::new(Url::parse("http://localhost:3000").unwrap()).unwrap();
        for target in [
            "http://localhost:8000/auth/google",
            "https://localhost:3000",
            "http://localhost.example.com:3000",
            "http://localhost:3000@example.com",
            "http://user:password@localhost:3000",
            "file:///etc/passwd",
            "javascript:alert(1)",
        ] {
            assert!(!policy.allows_navigation(&Url::parse(target).unwrap()));
        }
    }

    #[test]
    fn validates_local_startup_targets() {
        for target in [
            "http://localhost:3000",
            "http://127.0.0.1:3001",
            "http://[::1]:3000",
        ] {
            assert!(NavigationPolicy::new(Url::parse(target).unwrap()).is_ok());
        }
        for target in [
            "https://cloud.fortyone.app",
            "http://localhost.example.com:3000",
            "http://user:password@localhost:3000",
            "http://localhost:3000/#fragment",
            "file:///tmp/index.html",
        ] {
            assert!(NavigationPolicy::new(Url::parse(target).unwrap()).is_err());
        }
    }

    #[test]
    fn only_opens_web_and_email_links_in_external_apps() {
        for target in [
            "https://example.com",
            "http://localhost:8000/auth/google",
            "mailto:hello@example.com",
        ] {
            assert!(can_open_externally(&Url::parse(target).unwrap()));
        }
        for target in [
            "file:///etc/passwd",
            "javascript:alert(1)",
            "data:text/html,hello",
            "fortyone://login",
            "https://user:password@example.com",
        ] {
            assert!(!can_open_externally(&Url::parse(target).unwrap()));
        }
    }
}
