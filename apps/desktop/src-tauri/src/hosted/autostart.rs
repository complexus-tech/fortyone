//! Apply the login-launch default once, without overriding an existing preference
//! or re-registering after the user disables the login item in system settings.
use std::path::{Path, PathBuf};

use rusqlite::Connection;
use tauri::{AppHandle, Manager};
use tauri_plugin_autostart::{AutoLaunchManager, ManagerExt};

use super::{preferences, State};

trait Autostart {
    fn is_enabled(&self) -> Result<bool, String>;
    fn enable(&self) -> Result<(), String>;
}

impl Autostart for AutoLaunchManager {
    fn is_enabled(&self) -> Result<bool, String> {
        self.is_enabled().map_err(|error| error.to_string())
    }

    fn enable(&self) -> Result<(), String> {
        self.enable().map_err(|error| error.to_string())
    }
}

fn eligible_installation(
    production_release: bool,
    identifier: &str,
    product_name: Option<&str>,
    executable: &Path,
    installed_executables: &[PathBuf],
) -> bool {
    production_release
        && identifier == "app.fortyone.desktop"
        && product_name == Some("FortyOne")
        && installed_executables.iter().any(|path| path == executable)
}

fn installed_executables(app: &AppHandle) -> Result<Vec<PathBuf>, String> {
    #[cfg(target_os = "macos")]
    let paths = vec![
        PathBuf::from("/Applications/FortyOne.app/Contents/MacOS/fortyone-desktop"),
        app.path()
            .home_dir()
            .map_err(|error| error.to_string())?
            .join("Applications/FortyOne.app/Contents/MacOS/fortyone-desktop"),
    ];
    #[cfg(target_os = "windows")]
    let paths = {
        let mut roots = vec![app
            .path()
            .local_data_dir()
            .map_err(|error| error.to_string())?];
        roots.extend(
            ["ProgramFiles", "ProgramW6432", "ProgramFiles(x86)"]
                .into_iter()
                .filter_map(std::env::var_os)
                .map(PathBuf::from),
        );
        roots
            .into_iter()
            .map(|root| root.join("FortyOne/fortyone-desktop.exe"))
            .collect::<Vec<_>>()
    };
    // Automatic registration is limited to the supported installed packages.
    // Other locations/platforms retain the explicit preference control.
    #[cfg(not(any(target_os = "macos", target_os = "windows")))]
    let paths: Vec<PathBuf> = {
        let _ = app;
        Vec::new()
    };
    // Windows canonical paths use the verbatim prefix; normalize candidates in
    // the same way as current_exe. On macOS require the actual install location,
    // so a symlink in Applications cannot enroll an app on a mounted disk.
    #[cfg(target_os = "windows")]
    let paths = paths
        .into_iter()
        .filter_map(|path| path.canonicalize().ok())
        .collect();
    Ok(paths)
}

pub(super) fn initialize(app: &AppHandle) -> Result<(), String> {
    if !cfg!(all(feature = "production", not(debug_assertions))) {
        return Ok(());
    }
    let executable = std::env::current_exe()
        .and_then(|path| path.canonicalize())
        .map_err(|error| error.to_string())?;
    if !eligible_installation(
        true,
        &app.config().identifier,
        app.config().product_name.as_deref(),
        &executable,
        &installed_executables(app)?,
    ) {
        return Ok(());
    }

    let state = app.state::<std::sync::Arc<State>>();
    let db = state.db.lock().map_err(|error| error.to_string())?;
    initialize_default(&db, &*app.autolaunch())
}

fn initialize_default(db: &Connection, autostart: &impl Autostart) -> Result<(), String> {
    // Older versions stored the real OS value alongside other preferences. A
    // saved false is treated as a choice, even if another setting caused the save.
    if preferences::read_stored(db)?.is_some() {
        return Ok(());
    }
    // Claim the one-time attempt before touching the OS. Even a failed or
    // interrupted registration must not keep prompting or undo a later OS choice.
    let claimed = db
        .execute(
            "INSERT INTO hosted_preferences(key,value) VALUES('autostart-default-v1','attempted') ON CONFLICT(key) DO NOTHING",
            [],
        )
        .map_err(|error| error.to_string())?;
    if claimed == 0 || autostart.is_enabled()? {
        return Ok(());
    }
    autostart.enable()?;
    if !autostart.is_enabled()? {
        return Err("The operating system did not enable launch at login".into());
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use std::cell::Cell;

    use super::*;

    #[derive(Default)]
    struct FakeAutostart {
        enabled: Cell<bool>,
        calls: Cell<usize>,
        enable_error: bool,
        read_error: bool,
        ignores_enable: bool,
    }

    impl Autostart for FakeAutostart {
        fn is_enabled(&self) -> Result<bool, String> {
            if self.read_error {
                return Err("Could not read the login item".into());
            }
            Ok(self.enabled.get())
        }

        fn enable(&self) -> Result<(), String> {
            self.calls.set(self.calls.get() + 1);
            if self.enable_error {
                return Err("Registration denied".into());
            }
            self.enabled.set(!self.ignores_enable);
            Ok(())
        }
    }

    fn database() -> Connection {
        let db = Connection::open_in_memory().unwrap();
        db.execute_batch(
            "CREATE TABLE hosted_preferences(key TEXT PRIMARY KEY,value TEXT NOT NULL);",
        )
        .unwrap();
        db
    }

    #[test]
    fn fresh_install_registers_once_and_preserves_later_system_disable() {
        let db = database();
        let autostart = FakeAutostart::default();
        initialize_default(&db, &autostart).unwrap();
        assert!(autostart.enabled.get());
        autostart.enabled.set(false);
        initialize_default(&db, &autostart).unwrap();
        assert!(!autostart.enabled.get());
        assert_eq!(autostart.calls.get(), 1);
    }

    #[test]
    fn saved_preferences_are_preserved_even_when_os_state_differs() {
        for saved in [true, false] {
            let db = database();
            db.execute(
                "INSERT INTO hosted_preferences VALUES('preferences',?)",
                [format!(r#"{{"autostart":{saved}}}"#)],
            )
            .unwrap();
            let autostart = FakeAutostart {
                enabled: Cell::new(!saved),
                ..Default::default()
            };
            initialize_default(&db, &autostart).unwrap();
            assert_eq!(autostart.enabled.get(), !saved);
            assert_eq!(autostart.calls.get(), 0);
            assert_eq!(preferences::read(&db).unwrap().autostart, saved);
        }
    }

    #[test]
    fn an_existing_login_item_is_not_registered_again() {
        let db = database();
        let autostart = FakeAutostart {
            enabled: Cell::new(true),
            ..Default::default()
        };
        initialize_default(&db, &autostart).unwrap();
        autostart.enabled.set(false);
        initialize_default(&db, &autostart).unwrap();
        assert_eq!(autostart.calls.get(), 0);
        assert!(!autostart.enabled.get());
    }

    #[test]
    fn registration_failures_are_reported_without_repeated_attempts() {
        for autostart in [
            FakeAutostart {
                enable_error: true,
                ..Default::default()
            },
            FakeAutostart {
                read_error: true,
                ..Default::default()
            },
            FakeAutostart {
                ignores_enable: true,
                ..Default::default()
            },
        ] {
            let db = database();
            assert!(initialize_default(&db, &autostart).is_err());
            let calls = autostart.calls.get();
            initialize_default(&db, &autostart).unwrap();
            assert_eq!(autostart.calls.get(), calls);
            assert!(!autostart.enabled.get());
        }
    }

    #[test]
    fn cannot_register_when_the_attempt_cannot_be_saved() {
        let db = database();
        db.execute_batch("PRAGMA query_only=ON;").unwrap();
        let autostart = FakeAutostart::default();
        assert!(initialize_default(&db, &autostart).is_err());
        assert_eq!(autostart.calls.get(), 0);
    }

    #[test]
    fn only_the_installed_production_app_receives_the_default() {
        let installed = PathBuf::from("/Applications/FortyOne.app/Contents/MacOS/fortyone-desktop");
        for (release, identifier, name, executable, expected) in [
            (
                true,
                "app.fortyone.desktop",
                "FortyOne",
                installed.as_path(),
                true,
            ),
            (
                false,
                "app.fortyone.desktop",
                "FortyOne",
                installed.as_path(),
                false,
            ),
            (
                true,
                "app.fortyone.desktop.dev",
                "FortyOne",
                installed.as_path(),
                false,
            ),
            (
                true,
                "app.fortyone.desktop",
                "FortyOne Main QA",
                installed.as_path(),
                false,
            ),
            (
                true,
                "app.fortyone.desktop",
                "FortyOne",
                Path::new("/Volumes/FortyOne/FortyOne.app/Contents/MacOS/fortyone-desktop"),
                false,
            ),
            (
                true,
                "app.fortyone.desktop",
                "FortyOne",
                Path::new("/tmp/target/release/fortyone-desktop"),
                false,
            ),
        ] {
            assert_eq!(
                eligible_installation(
                    release,
                    identifier,
                    Some(name),
                    executable,
                    std::slice::from_ref(&installed)
                ),
                expected
            );
        }
    }
}
