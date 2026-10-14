//! Remember one safe location at normal shutdown, without adding filesystem
//! work to route transitions or storing authentication and search parameters.

use std::{
    fs::{self, File, OpenOptions},
    io::{self, Read, Write},
    path::Path,
};

#[cfg(feature = "production")]
use tauri::Manager;
use tauri::Url;

use crate::navigation::NavigationPolicy;

const FILE_NAME: &str = "last-workspace-url";
const MAX_LOCATION_BYTES: usize = 8_192;

fn read_location(path: &Path, policy: &NavigationPolicy) -> io::Result<Option<Url>> {
    let file = match File::open(path) {
        Ok(file) => file,
        Err(error) if error.kind() == io::ErrorKind::NotFound => return Ok(None),
        Err(error) => return Err(error),
    };
    let mut value = String::new();
    file.take((MAX_LOCATION_BYTES + 1) as u64)
        .read_to_string(&mut value)?;
    if value.len() > MAX_LOCATION_BYTES {
        return Ok(None);
    }
    Ok(Url::parse(&value)
        .ok()
        .and_then(|url| policy.restorable_location(&url))
        .filter(|url| url.as_str() == value))
}

fn write_location(path: &Path, policy: &NavigationPolicy, url: &Url) -> io::Result<()> {
    let Some(location) = policy.restorable_location(url) else {
        return Ok(());
    };
    if location.as_str().len() > MAX_LOCATION_BYTES {
        return Ok(());
    }
    if read_location(path, policy).ok().flatten().as_ref() == Some(&location) {
        return Ok(());
    }
    if let Some(parent) = path.parent() {
        fs::create_dir_all(parent)?;
    }
    let temporary = path.with_extension(format!("{}.tmp", std::process::id()));
    let result = (|| {
        let mut options = OpenOptions::new();
        options.write(true).create_new(true);
        #[cfg(unix)]
        {
            use std::os::unix::fs::OpenOptionsExt;
            options.mode(0o600);
        }
        let mut file = options.open(&temporary)?;
        file.write_all(location.as_str().as_bytes())?;
        file.sync_all()?;
        drop(file);
        fs::rename(&temporary, path)
    })();
    if result.is_err() {
        let _ = fs::remove_file(&temporary);
    }
    result
}

fn clear_location(path: &Path) -> io::Result<()> {
    match fs::remove_file(path) {
        Err(error) if error.kind() == io::ErrorKind::NotFound => Ok(()),
        result => result,
    }
}

#[cfg(feature = "production")]
pub fn start_url(app: &tauri::AppHandle, policy: &NavigationPolicy) -> Url {
    match app.path().app_config_dir() {
        Ok(directory) => match read_location(&directory.join(FILE_NAME), policy) {
            Ok(Some(url)) => return url,
            Ok(None) => {}
            Err(error) => eprintln!("Could not restore the last workspace: {error}"),
        },
        Err(error) => eprintln!("Could not find the desktop settings directory: {error}"),
    }
    policy.app_url().clone()
}

#[cfg(feature = "production")]
pub fn save_window(window: &tauri::WebviewWindow) {
    let policy = window.state::<std::sync::Arc<NavigationPolicy>>();
    let result = (|| -> Result<(), Box<dyn std::error::Error>> {
        let url = window.url()?;
        let path = window.app_handle().path().app_config_dir()?.join(FILE_NAME);
        if policy.clears_restored_location(&url) {
            clear_location(&path)?;
        } else {
            write_location(&path, &policy, &url)?;
        }
        Ok(())
    })();
    if let Err(error) = result {
        eprintln!("Could not save the last workspace: {error}");
    }
}

#[cfg(feature = "production")]
pub fn clear_after_logout(app: &tauri::AppHandle, policy: &NavigationPolicy, url: &Url) {
    if !policy.clears_restored_location(url) {
        return;
    }
    let result = (|| -> Result<(), Box<dyn std::error::Error>> {
        clear_location(&app.path().app_config_dir()?.join(FILE_NAME))?;
        Ok(())
    })();
    if let Err(error) = result {
        eprintln!("Could not clear the last workspace after sign-out: {error}");
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::sync::atomic::{AtomicUsize, Ordering};

    static NEXT_DIRECTORY: AtomicUsize = AtomicUsize::new(0);

    struct TestDirectory(std::path::PathBuf);

    impl TestDirectory {
        fn new() -> Self {
            let path = std::env::temp_dir().join(format!(
                "fortyone-session-test-{}-{}",
                std::process::id(),
                NEXT_DIRECTORY.fetch_add(1, Ordering::Relaxed)
            ));
            fs::create_dir_all(&path).unwrap();
            Self(path)
        }

        fn location(&self) -> std::path::PathBuf {
            self.0.join(FILE_NAME)
        }
    }

    impl Drop for TestDirectory {
        fn drop(&mut self) {
            let _ = fs::remove_dir_all(&self.0);
        }
    }

    #[test]
    fn persists_valid_pages_without_query_or_fragment_and_replaces_previous_location() {
        let directory = TestDirectory::new();
        let path = directory.location();
        let policy = NavigationPolicy::production();
        assert_eq!(read_location(&path, &policy).unwrap(), None);
        for target in [
            "https://acme.fortyone.app/work/ENG-42?token=secret#private",
            "https://other-team.fortyone.app/docs/document-id?view=edit",
        ] {
            let url = Url::parse(target).unwrap();
            write_location(&path, &policy, &url).unwrap();
            let expected = policy.restorable_location(&url).unwrap();
            assert_eq!(fs::read_to_string(&path).unwrap(), expected.as_str());
            assert_eq!(read_location(&path, &policy).unwrap(), Some(expected));
        }
        clear_location(&path).unwrap();
        clear_location(&path).unwrap();
        assert_eq!(read_location(&path, &policy).unwrap(), None);
    }

    #[test]
    fn rejects_tampered_oversized_and_non_workspace_state() {
        let directory = TestDirectory::new();
        let path = directory.location();
        let policy = NavigationPolicy::production();
        for value in [
            "not a URL".to_owned(),
            "https://api.fortyone.app/work/ENG-42".to_owned(),
            "https://cloud.fortyone.app/auth-callback?code=secret".to_owned(),
            "https://acme.fortyone.app/verify/email/token".to_owned(),
            "https://acme.fortyone.app/shared/docs/secret".to_owned(),
            "https://attacker.example/work/ENG-42".to_owned(),
            "https://acme.fortyone.app/work/ENG-42?token=secret".to_owned(),
            "https://acme.fortyone.app/work/ENG-42#private".to_owned(),
            "https://acme.fortyone.app/my-work\n".to_owned(),
            format!(
                "https://acme.fortyone.app/work/{}",
                "x".repeat(MAX_LOCATION_BYTES)
            ),
        ] {
            fs::write(&path, value).unwrap();
            assert_eq!(read_location(&path, &policy).unwrap(), None);
        }
    }

    #[test]
    fn transient_auth_pages_do_not_replace_a_saved_workspace() {
        let directory = TestDirectory::new();
        let path = directory.location();
        let policy = NavigationPolicy::production();
        let workspace = Url::parse("https://acme.fortyone.app/my-work").unwrap();
        write_location(&path, &policy, &workspace).unwrap();
        write_location(
            &path,
            &policy,
            &Url::parse("https://cloud.fortyone.app/auth-callback").unwrap(),
        )
        .unwrap();
        assert_eq!(read_location(&path, &policy).unwrap(), Some(workspace));
    }

    #[test]
    fn failed_replacement_preserves_the_previous_location() {
        let directory = TestDirectory::new();
        let path = directory.location();
        let policy = NavigationPolicy::production();
        let previous = Url::parse("https://acme.fortyone.app/my-work").unwrap();
        write_location(&path, &policy, &previous).unwrap();
        // An unavailable temporary destination must not truncate the saved URL.
        fs::create_dir(path.with_extension(format!("{}.tmp", std::process::id()))).unwrap();
        let next = Url::parse("https://acme.fortyone.app/work/ENG-42").unwrap();
        assert!(write_location(&path, &policy, &next).is_err());
        assert_eq!(read_location(&path, &policy).unwrap(), Some(previous));
    }
}
