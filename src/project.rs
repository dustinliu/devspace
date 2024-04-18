use anyhow::{anyhow, Context, Result};
use jsonc_parser::parse_to_serde_value;
use serde::Deserialize;
use std::{
    io::{BufReader, Read},
    path::PathBuf,
};

const CONFIG_DIR: &str = ".devcontainer";
const CONFIG_FILE: &str = "devcontainer.json";

#[derive(Debug, PartialEq, Eq)]
pub enum ImageSource {
    Image(String),
    Dockerfile(String),
}

impl Default for ImageSource {
    fn default() -> Self {
        ImageSource::Image(String::new())
    }
}

#[derive(Debug, Deserialize, Default)]
#[serde(rename_all = "camelCase")]
pub struct Config {
    name: Option<String>,
    image: Option<String>,
    #[serde(rename = "dockerFile")]
    dockerfile: Option<String>,
    #[serde(skip)]
    pub image_source: ImageSource,
    pub post_create_command: Option<Vec<String>>,
}

impl Config {
    pub fn new(input: impl Read) -> Result<Self> {
        let mut reader = BufReader::new(input);
        let mut content = String::new();
        reader
            .read_to_string(&mut content)
            .context("config read failed")?;
        let value =
            parse_to_serde_value(&content, &Default::default()).context("config parse failed")?;

        let config: Config = match value {
            Some(value) => serde_json::from_value(value).context("config parse failed")?,
            None => return Err(anyhow!("invalid config")),
        };
        normalize_config(config)
    }
}

fn normalize_config(mut config: Config) -> Result<Config> {
    if let Some(image) = &config.image {
        config.image_source = ImageSource::Image(image.to_owned());
    } else if let Some(docker_file) = &config.dockerfile {
        config.image_source = ImageSource::Dockerfile(
            PathBuf::from(CONFIG_DIR)
                .join(docker_file)
                .display()
                .to_string(),
        );
    } else {
        return Err(anyhow!("invalid config, dockerfile or image not specified"));
    }

    Ok(config)
}

#[derive(Debug, Default)]
pub struct Project {
    pub root: PathBuf,
    pub config_dir: PathBuf,
    pub config_file: PathBuf,
    pub name: String,
    pub config: Config,
}

impl Project {
    fn new<P: Into<PathBuf>>(root: P, config: Config) -> Result<Self> {
        let root = root.into();
        let name = match &config.name {
            Some(name) => name.to_string(),
            None => get_project_name(&root)?,
        }
        .replace(' ', "_");
        let config_dir = root.join(CONFIG_DIR);
        let config_file = config_dir.join(CONFIG_FILE);

        Ok(Self {
            root,
            config_dir,
            config_file,
            name,
            config,
        })
    }
}

impl TryFrom<&PathBuf> for Project {
    type Error = anyhow::Error;

    fn try_from(root: &PathBuf) -> Result<Self> {
        let c = root.join(CONFIG_DIR).join(CONFIG_FILE);
        let f = std::fs::File::open(&c)
            .with_context(|| format!("failed to open config file {:?}", &c))?;
        let config = Config::new(f)?;
        Project::new(root, config)
    }
}

impl TryFrom<&str> for Project {
    type Error = anyhow::Error;

    fn try_from(root: &str) -> Result<Self> {
        Project::try_from(&PathBuf::from(root))
    }
}

impl AsRef<Project> for Project {
    fn as_ref(&self) -> &Project {
        self
    }
}

fn get_project_name(root: &PathBuf) -> Result<String> {
    let root = PathBuf::from(root);
    let name = root.file_name().ok_or_else(|| anyhow!("invalid path"))?;

    Ok(name
        .to_str()
        .ok_or_else(|| anyhow!("invalid path"))?
        .to_string())
}

#[cfg(test)]
pub mod tests {
    use super::*;
    use tempfile::TempDir;

    #[test]
    fn test_config() {
        let json = r#"
        {
            "name": "test",
            "dockerFile": "Dockerfile",
            "postCreateCommand": ["echo", "hello"]
        }"#;

        let config = Config::new(json.as_bytes()).unwrap();
        assert_eq!(config.name.unwrap(), "test");
        assert_eq!(
            config.image_source,
            ImageSource::Dockerfile(
                PathBuf::from(CONFIG_DIR)
                    .join("Dockerfile")
                    .display()
                    .to_string()
            )
        );
        assert_eq!(config.post_create_command.unwrap(), ["echo", "hello"]);
    }

    #[test]
    fn test_config_none() {
        let json = r#"
        {
            //test
            "image": "test image",
            "postCreateCommand": ["echo", "hello"],
        }"#;

        let config = Config::new(json.as_bytes()).unwrap();
        assert_eq!(
            config.image_source,
            ImageSource::Image("test image".to_string())
        );
        assert_eq!(config.name, None);
    }

    #[test]
    #[should_panic]
    fn test_config_invalid_format() {
        let json = r#"
        {
        f
            "dockerFile": "Dockerfile",
            "postCreateCommand": ["echo", "hello"],
        }"#;

        Config::new(json.as_bytes()).unwrap();
    }

    #[test]
    fn test_validate_config() {
        let json = r#"
        {
            "image": "test",
            "dockerFile": "Dockerfile",
            "postCreateCommand": ["echo", "hello"],
        }"#;
        let config = normalize_config(Config::new(json.as_bytes()).unwrap()).unwrap();
        assert_eq!(config.image_source, ImageSource::Image("test".to_string()));
    }

    #[test]
    #[should_panic]
    fn test_invalidate_config_without_dockfile_image() {
        let json = r#"
        {
            "postCreateCommand": ["echo", "hello"],
        }"#;
        normalize_config(Config::new(json.as_bytes()).unwrap()).unwrap();
    }

    #[test]
    fn test_project_name_from_path() {
        let json = r#"
        {
            "postCreateCommand": ["echo", "hello"],
            "image": "test"
        }"#;
        let tmp_project = TmpProjectDir::new("xxxx xxx").devcontainer_json(json.as_bytes());
        let project = Project::try_from(&tmp_project.root).unwrap();
        assert_eq!(project.name, "xxxx_xxx");
    }

    #[test]
    fn test_project_name_from_config() {
        let json = r#"
        {
            "name": "taskcommander dev",
            "postCreateCommand": ["echo", "hello"],
            "image": "test"
        }"#;

        let tmp_project = TmpProjectDir::new("xxxx").devcontainer_json(json.as_bytes());
        let project = Project::try_from(&tmp_project.root).unwrap();
        assert_eq!(project.name, "taskcommander_dev");
    }

    pub struct TmpProjectDir {
        pub name: String,
        _tmpdir: TempDir,
        pub root: PathBuf,
    }

    impl TmpProjectDir {
        pub fn new(name: &str) -> Self {
            let tmpdir = TempDir::new().unwrap();
            let root = tmpdir.path().join("devspace_test").join(name);
            std::fs::create_dir_all(root.join(CONFIG_DIR)).unwrap();

            Self {
                name: name.to_string(),
                _tmpdir: tmpdir,
                root,
            }
        }

        // pub fn dockerfile(self, name: &str, mut content: impl std::io::Read) -> Self {
        //     let mut buffer = Vec::new();
        //     content.read_to_end(&mut buffer).unwrap();
        //     let mut f = std::fs::File::create(self.root.join(CONFIG_DIR).join(name)).unwrap();
        //     f.write_all(&buffer).unwrap();
        //     self
        // }

        pub fn devcontainer_json(self, content: impl AsRef<[u8]>) -> Self {
            std::fs::write(self.root.join(CONFIG_DIR).join(CONFIG_FILE), content).unwrap();
            self
        }
    }
}
