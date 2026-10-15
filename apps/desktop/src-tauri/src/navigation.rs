use tauri::Url;

// Infrastructure and non-workspace sites. Do not copy every backend restricted
// creation slug: existing workspaces such as `complexus` still use those names.
const SERVICE_SUBDOMAINS: &[&str] = &[
    "admin",
    "api",
    "app",
    "assets",
    "auth",
    "cdn",
    "collaboration",
    "dev",
    "docs",
    "internal",
    "ops",
    "pitch",
    "preview",
    "prod",
    "production",
    "qa",
    "staging",
    "static",
    "status",
    "test",
    "testing",
    "worker",
    "www",
];

/// The shell never needs to navigate to an API or identity provider.
/// Those requests either use fetch or open in the user's browser.
pub struct NavigationPolicy {
    app_url: Url,
    production: bool,
}

impl NavigationPolicy {
    #[cfg(any(not(feature = "production"), test))]
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

        Ok(Self {
            app_url,
            production: false,
        })
    }

    #[cfg(any(feature = "production", test))]
    pub fn production() -> Self {
        Self {
            // Release navigation and native permissions must not be redirected by
            // build-machine or launch-environment URL overrides.
            app_url: Url::parse("https://cloud.fortyone.app/")
                .expect("The production app URL is valid"),
            production: true,
        }
    }

    pub fn app_url(&self) -> &Url {
        &self.app_url
    }

    /// Restore only authenticated app pages, never authentication callbacks,
    /// public links, or URL-carried credentials and transient filter state.
    pub fn restorable_location(&self, url: &Url) -> Option<Url> {
        if !self.allows_navigation(url)
            || (self.production && url.host_str() == self.app_url.host_str())
        {
            return None;
        }
        let mut segments = url.path().split('/').skip(1);
        if !self.production {
            let workspace = segments.next()?;
            if workspace.is_empty()
                || workspace.len() > 63
                || workspace.starts_with('-')
                || workspace.ends_with('-')
                || !workspace
                    .bytes()
                    .all(|byte| byte.is_ascii_lowercase() || byte.is_ascii_digit() || byte == b'-')
                || matches!(
                    workspace,
                    "auth" | "onboarding" | "shared" | "verify" | "api"
                )
            {
                return None;
            }
        }
        let route = segments.next()?;
        if !matches!(
            route,
            "my-work"
                | "summary"
                | "strategy"
                | "analytics"
                | "roadmap"
                | "docs"
                | "maya"
                | "sprints"
                | "key-results"
                | "calendar"
                | "notifications"
                | "story"
                | "work"
                | "teams"
                | "profile"
                | "views"
                | "search"
                | "settings"
        ) {
            return None;
        }
        let mut location = url.clone();
        location.set_query(None);
        location.set_fragment(None);
        Some(location)
    }

    pub fn clears_restored_location(&self, url: &Url) -> bool {
        (self.production && url.as_str() == "https://fortyone.app/?signedOut=true")
            || (self.allows_navigation(url)
                && url.origin() == self.app_url.origin()
                && url.path() == "/")
    }

    pub fn allows_navigation(&self, url: &Url) -> bool {
        if has_credentials(url) {
            return false;
        }
        if !self.production {
            return url.origin() == self.app_url.origin();
        }
        if url.scheme() != "https" || url.port_or_known_default() != Some(443) {
            return false;
        }
        let Some(label) = url
            .host_str()
            .and_then(|host| host.strip_suffix(".fortyone.app"))
        else {
            return false;
        };
        !label.is_empty()
            && label.len() <= 63
            && !label.starts_with('-')
            && !label.ends_with('-')
            && label
                .bytes()
                .all(|byte| byte.is_ascii_lowercase() || byte.is_ascii_digit() || byte == b'-')
            && !SERVICE_SUBDOMAINS.contains(&label)
    }

    pub fn app_destination(&self, url: &Url) -> Option<Url> {
        if self.allows_navigation(url) {
            return Some(url.clone());
        }
        // The hosted logout flow returns to the marketing homepage. Only this
        // exact logout destination is replaced; other marketing links stay external.
        if self.production && url.as_str() == "https://fortyone.app/?signedOut=true" {
            let mut destination = self.app_url.clone();
            destination.set_query(Some("signedOut=true"));
            return Some(destination);
        }
        None
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
    fn restores_only_known_workspace_pages_without_transient_url_state() {
        let policy = NavigationPolicy::production();
        let target =
            Url::parse("https://acme.fortyone.app/work/ENG-42?token=secret#private").unwrap();
        assert_eq!(
            policy.restorable_location(&target).unwrap().as_str(),
            "https://acme.fortyone.app/work/ENG-42"
        );
        for target in [
            "https://cloud.fortyone.app/work/ENG-42",
            "https://api.fortyone.app/work/ENG-42",
            "https://acme.fortyone.app/auth-callback?code=secret",
            "https://acme.fortyone.app/verify/email/token",
            "https://acme.fortyone.app/shared/docs/token",
            "https://acme.fortyone.app/onboarding",
            "https://acme.fortyone.app/unknown",
            "https://acme.fortyone.app/",
            "https://user:password@acme.fortyone.app/my-work",
            "https://acme.fortyone.app:8443/my-work",
            "http://acme.fortyone.app/my-work",
        ] {
            assert!(policy
                .restorable_location(&Url::parse(target).unwrap())
                .is_none());
        }
        let local = NavigationPolicy::new(Url::parse("http://localhost:3000/").unwrap()).unwrap();
        assert_eq!(
            local
                .restorable_location(
                    &Url::parse("http://localhost:3000/acme/my-work?filter=mine#private").unwrap()
                )
                .unwrap()
                .as_str(),
            "http://localhost:3000/acme/my-work"
        );
        for target in [
            "http://localhost:3000/auth/desktop",
            "http://localhost:3000/auth/my-work",
            "http://localhost:3000/shared/docs/private-token",
            "http://localhost:3000/acme/verify/private-token",
            "http://localhost:4000/acme/my-work",
            "http://127.0.0.1:3000/acme/my-work",
            "http://localhost:3000/",
        ] {
            assert!(local
                .restorable_location(&Url::parse(target).unwrap())
                .is_none());
        }
        assert!(local.clears_restored_location(
            &Url::parse("http://localhost:3000/?signedOut=true").unwrap()
        ));
    }

    #[test]
    fn sign_out_and_returning_to_cloud_login_clear_restored_workspace() {
        let policy = NavigationPolicy::production();
        for target in [
            "https://fortyone.app/?signedOut=true",
            "https://cloud.fortyone.app/?signedOut=true",
            "https://cloud.fortyone.app/",
            "https://cloud.fortyone.app/?callbackUrl=%2Fmy-work",
        ] {
            assert!(policy.clears_restored_location(&Url::parse(target).unwrap()));
        }
        for target in [
            "https://cloud.fortyone.app/auth-callback",
            "https://acme.fortyone.app/my-work",
            "https://fortyone.app/",
            "https://attacker.example/?signedOut=true",
            "https://user@cloud.fortyone.app/",
        ] {
            assert!(!policy.clears_restored_location(&Url::parse(target).unwrap()));
        }
    }

    #[test]
    fn production_keeps_auth_and_workspace_transitions_in_the_app() {
        let policy = NavigationPolicy::production();
        assert_eq!(policy.app_url().as_str(), "https://cloud.fortyone.app/");
        for target in [
            "https://cloud.fortyone.app/acme/maya",
            "https://cloud.fortyone.app/auth-callback?callbackUrl=%2Facme",
            "https://cloud.fortyone.app/acme/docs#heading",
            "https://acme.fortyone.app/my-work",
            "https://other-team.fortyone.app/settings?tab=members",
            "https://complexus.fortyone.app/maya",
            "https://product.fortyone.app/roadmap",
            "https://cloud.fortyone.app/",
        ] {
            assert!(policy.allows_navigation(&Url::parse(target).unwrap()));
            assert_eq!(
                policy
                    .app_destination(&Url::parse(target).unwrap())
                    .unwrap()
                    .as_str(),
                target
            );
        }
    }

    #[test]
    fn production_rejects_services_nested_hosts_and_non_https_origins() {
        let policy = NavigationPolicy::production();
        for target in [
            "http://cloud.fortyone.app/acme",
            "https://cloud.fortyone.app:8443/acme",
            "https://cloud.fortyone.app.attacker.example/acme",
            "https://api.fortyone.app/",
            "https://app.fortyone.app/",
            "https://user:password@cloud.fortyone.app/",
            "http://localhost:3000/",
            "file:///tmp/index.html",
            "https://fortyone.app/",
            "https://one.two.fortyone.app/",
            "https://acme.fortyone.app./",
            "https://-acme.fortyone.app/",
            "https://acme-.fortyone.app/",
            "https://acme_team.fortyone.app/",
            "https://acme.fortyone.app@attacker.example/",
        ] {
            assert!(
                !policy.allows_navigation(&Url::parse(target).unwrap()),
                "{target}"
            );
        }
        for label in SERVICE_SUBDOMAINS {
            let target = Url::parse(&format!("https://{label}.fortyone.app/")).unwrap();
            assert!(!policy.allows_navigation(&target), "{label}");
        }
        let too_long = Url::parse(&format!("https://{}.fortyone.app/", "a".repeat(64))).unwrap();
        assert!(!policy.allows_navigation(&too_long));
    }

    #[test]
    fn only_the_exact_production_logout_returns_to_in_app_sign_in() {
        let policy = NavigationPolicy::production();
        assert_eq!(
            policy
                .app_destination(&Url::parse("https://fortyone.app/?signedOut=true").unwrap())
                .unwrap()
                .as_str(),
            "https://cloud.fortyone.app/?signedOut=true"
        );
        for target in [
            "https://fortyone.app/",
            "https://fortyone.app/contact",
            "https://fortyone.app/?signedOut=false",
            "https://fortyone.app/?signedOut=true&callbackUrl=https://example.com",
            "https://fortyone.app/?signedOut=true#fragment",
            "https://user@fortyone.app/?signedOut=true",
            "http://fortyone.app/?signedOut=true",
        ] {
            assert!(
                policy
                    .app_destination(&Url::parse(target).unwrap())
                    .is_none(),
                "{target}"
            );
        }
        let local = NavigationPolicy::new(Url::parse("http://localhost:3000/").unwrap()).unwrap();
        assert!(local
            .app_destination(&Url::parse("https://fortyone.app/?signedOut=true").unwrap())
            .is_none());
    }

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
