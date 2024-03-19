use crate::docker::process;
use anyhow::{Context, Result};
use bollard::{
    container::ListContainersOptions,
    image::ListImagesOptions,
    models::{ContainerSummary, ImageSummary},
    Docker,
};
use std::{collections::HashMap, path::PathBuf};
use tokio::runtime::Builder;

#[cfg(test)]
use mockall::{automock, predicate::*};

const PROJECT_KEY: &str = "ds_project";

#[cfg_attr(test, automock)]
pub trait DockerClient {
    fn list_containers(&self, project_name: &str) -> Result<Vec<ContainerSummary>>;
    fn list_images(&self, project_name: &str) -> Result<Vec<ImageSummary>>;
    fn build_image(&self, project_name: &str, dockerfile: &str) -> Result<()>;
}

pub fn new_client() -> Result<Box<dyn DockerClient>> {
    let client = Docker::connect_with_local_defaults()?;
    let cli = DockerCli::new()?;
    Ok(Box::new(DockerClientImpl { client, cli }))
}

pub struct DockerClientImpl {
    client: Docker,
    cli: DockerCli,
}

impl DockerClientImpl {}

impl DockerClient for DockerClientImpl {
    fn list_containers(&self, project_name: &str) -> Result<Vec<ContainerSummary>> {
        let options = ListContainersOptions::<String> {
            all: true,
            filters: HashMap::from([(
                "label".to_owned(),
                vec![format!(
                    "{}={}",
                    PROJECT_KEY.to_owned(),
                    project_name.to_owned()
                )],
            )]),
            ..Default::default()
        };
        let runtime = Builder::new_current_thread().enable_all().build()?;
        runtime
            .block_on(self.client.list_containers(Some(options)))
            .context("can not list containers")
    }

    fn list_images(&self, project_name: &str) -> Result<Vec<ImageSummary>> {
        let options = ListImagesOptions::<String> {
            all: true,
            filters: HashMap::from([(
                "label".to_owned(),
                vec![format!(
                    "{}={}",
                    PROJECT_KEY.to_owned(),
                    project_name.to_owned()
                )],
            )]),
            ..Default::default()
        };
        let runtime = Builder::new_current_thread().enable_all().build()?;
        runtime
            .block_on(self.client.list_images(Some(options)))
            .context("can not list images")
    }

    fn build_image(&self, project_name: &str, dockerfile: &str) -> Result<()> {
        let options = BuildOptions {
            tag: format!("{}:latest", project_name),
            path: PathBuf::from("."),
            dockerfile: dockerfile.to_owned(),
            labels: HashMap::from([(PROJECT_KEY.to_owned(), project_name.to_owned())]),
            ..Default::default()
        };
        self.cli.build(&options)
    }
}

#[derive(Debug)]
struct BuildOptions {
    cmd: String,
    tag: String,
    dockerfile: String,
    labels: HashMap<String, String>,
    path: PathBuf,
}

impl BuildOptions {
    fn build(&self) -> Vec<String> {
        let dockfile = self.path.join(&self.dockerfile);
        let mut args = vec![
            self.cmd.to_string(),
            "-t".to_string(),
            self.tag.to_string(),
            "-f".to_string(),
            dockfile.display().to_string(),
        ];
        args.push("--label".to_string());
        for (key, value) in &self.labels {
            args.push(format!("{}={}", key, value));
        }
        args.push(self.path.display().to_string());
        args
    }
}

impl Default for BuildOptions {
    fn default() -> Self {
        BuildOptions {
            cmd: "build".to_string(),
            tag: Default::default(),
            dockerfile: "Dockerfile".to_owned(),
            labels: Default::default(),
            path: PathBuf::from("."),
        }
    }
}

#[derive(Debug)]
struct RunOptions {
    cmd: String,
    name: String,
    deattch: bool,
    image: String,
    args: Vec<String>,
}

impl RunOptions {
    fn build(&self) -> Vec<String> {
        let mut args = vec![
            self.cmd.to_owned(),
            "--name".to_owned(),
            self.name.to_owned(),
        ];
        if self.deattch {
            args.push("-d".to_owned())
        }

        args.push(self.image.to_owned());
        args.extend(self.args.to_owned());

        args
    }
}

impl Default for RunOptions {
    fn default() -> Self {
        RunOptions {
            cmd: "run".to_owned(),
            name: Default::default(),
            deattch: false,
            image: Default::default(),
            args: Default::default(),
        }
    }
}

struct DockerCli {
    command: PathBuf,
}

impl DockerCli {
    fn new() -> Result<Self> {
        let command = which::which("docker").context("can not find docker executable")?;
        Ok(DockerCli { command })
    }

    fn build(&self, options: &BuildOptions) -> Result<()> {
        let args = options.build();
        process::pipe_cmd(&self.command, args)
    }

    fn run(&self, image: &str, options: &RunOptions) -> Result<()> {
        let args = options.build();
        process::pipe_cmd(&self.command, args)
    }
}
