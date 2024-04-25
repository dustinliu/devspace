use anyhow::Result;
use serde::Deserialize;

const PREFIX: &str = "devspace";
const CONFIG: &str = "config.toml";

#[derive(Debug, Deserialize, Default)]
pub struct Config {
    pub dotfiles: Option<String>,
}

impl Config {
    pub fn new() -> Result<Self> {
        let xdg = xdg::BaseDirectories::with_prefix(PREFIX)?;
        if let Some(f) = xdg.find_config_file(CONFIG) {
            let toml = std::fs::read_to_string(f)?;
            return Self::from_str(&toml);
        }
        Ok(Default::default())
    }

    fn from_str(toml: &str) -> Result<Self> {
        Ok(toml::from_str(toml)?)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_config() {
        let toml = r#"
        dotfiles = "uuuuuuu"
        "#;
        let config = Config::from_str(toml).unwrap();
        assert_eq!(config.dotfiles, Some("uuuuuuu".to_string()));

        let config = Config::from_str("").unwrap();
        assert_eq!(config.dotfiles, None);
    }
}
