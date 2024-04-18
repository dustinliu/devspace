// use anyhow::Result;
// use serde::{Deserialize, Serialize};
// use std::{fs, path::PathBuf};
//
// #[derive(Debug, Deserialize, Serialize, Default)]
// struct Image {
//     name: String,
// }
//
// #[derive(Debug, Deserialize, Serialize, Default)]
// struct State {
//     #[serde(skip)]
//     root: PathBuf,
//
//     image: Image,
// }
//
// impl State {
//     fn new<P: Into<PathBuf>>(root: P) -> Self {
//         let root = root.into();
//         State {
//             root,
//             ..Default::default()
//         }
//     }
//
//     fn save() -> Result<()> {
//         Ok(())
//     }
//
//     fn load() -> Result<Self> {
//         Ok(State::default())
//     }
// }
//
// trait DiskStore {
//     fn save_state(&self, state: &State) -> Result<()>;
//     fn load_state(&self) -> Result<State>;
// }
//
// const STATE_FILE: &str = "state.json";
// const PREFIX: &str = "devspace";
//
// struct XDGStore {
//     dirs: xdg::BaseDirectories,
// }
//
// impl XDGStore {
//     fn new() -> Result<Self> {
//         Ok(XDGStore {
//             dirs: xdg::BaseDirectories::with_prefix(PREFIX)?,
//         })
//     }
// }
//
// impl DiskStore for XDGStore {
//     fn save_state(&self, state: &State) -> Result<()> {
//         let state_file = self.dirs.place_state_file(STATE_FILE)?;
//         let f = fs::File::create(state_file)?;
//         serde_json::to_writer(f, state)?;
//         Ok(())
//     }
//
//     fn load_state(&self) -> Result<State> {
//         let state_file = self.dirs.place_state_file(STATE_FILE)?;
//         let content = fs::read_to_string(state_file)?;
//         Ok(serde_json::from_str(&content)?)
//     }
// }
//
// #[cfg(test)]
// mod tests {
//     use std::io::Write;
//
//     use super::*;
//     use tempfile::NamedTempFile;
//
//     struct TestStore {
//         state_file: NamedTempFile,
//     }
//
//     impl TestStore {
//         fn new() -> Self {
//             TestStore {
//                 state_file: NamedTempFile::new().unwrap(),
//             }
//         }
//     }
//
//     impl DiskStore for TestStore {
//         fn save_state(&self, state: &State) -> Result<()> {
//             let json = serde_json::to_string(state)?;
//             self.state_file.as_file().write_all(json.as_bytes())?;
//             Ok(())
//         }
//
//         fn load_state(&self) -> Result<State> {
//             let state = serde_json::from_reader(self.state_file.reopen()?)?;
//             Ok(state)
//         }
//     }
// }
