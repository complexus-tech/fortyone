fn main() {
    println!("cargo:rerun-if-env-changed=FORTYONE_DESKTOP_URL");
    if std::env::var_os("CARGO_FEATURE_PRODUCTION").is_none() {
        let url = std::env::var("FORTYONE_DESKTOP_URL")
            .unwrap_or_else(|_| "http://localhost:3000/".to_owned());
        println!("cargo:rustc-env=FORTYONE_DESKTOP_DEFAULT_URL={url}");
    }
    tauri_build::try_build(tauri_build::Attributes::new().app_manifest(
        tauri_build::AppManifest::new().commands(&[
            "sync_desktop_chrome",
            "desktop_navigation_state",
            "desktop_navigate",
            "desktop_window_action",
        ]),
    ))
    .expect("Could not build desktop permissions");
}
