use crate::docker::process;
use anyhow::{Context, Result};
use bollard::{
    container::{ListContainersOptions, StartContainerOptions, StopContainerOptions},
    image::ListImagesOptions,
    models::{ContainerSummary, ImageSummary},
    Docker,
};
use std::{collections::HashMap, path::PathBuf};
use tokio::runtime::Builder;

const PROJECT_KEY: &str = "ds_project";

pub trait DockerClient {
    fn list_containers(&self, project_name: &str) -> Result<Vec<ContainerSummary>>;
    fn list_images(&self, project_name: &str) -> Result<Vec<ImageSummary>>;
    fn build_image(&self, project_name: &str, dockerfile: &str) -> Result<()>;
    fn start_container(&self, name: &str) -> Result<()>;
    fn stop_container(&self, name: &str) -> Result<()>;
    fn run(&self, name: &str, image_name: &str, deattach: bool, args: Vec<&str>) -> Result<()>;
    fn exec(&self, contain_name: &str, cmd: &[&str]) -> Result<()>;
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

impl DockerClient for DockerClientImpl {
    fn list_containers(&self, project_name: &str) -> Result<Vec<ContainerSummary>> {
        let label = format!("{}={}", PROJECT_KEY, &project_name);
        let options = ListContainersOptions {
            all: true,
            filters: HashMap::from([("label", vec![label.as_ref()])]),
            ..Default::default()
        };
        let runtime = Builder::new_current_thread().enable_all().build()?;
        runtime
            .block_on(self.client.list_containers(Some(options)))
            .context("can not list containers")
    }

    fn list_images(&self, project_name: &str) -> Result<Vec<ImageSummary>> {
        let label = format!("{}={}", PROJECT_KEY, project_name);

        let options = ListImagesOptions {
            all: true,
            filters: HashMap::from([("label", vec![label.as_str()])]),
            ..Default::default()
        };
        let runtime = Builder::new_current_thread().enable_all().build()?;
        runtime
            .block_on(self.client.list_images(Some(options)))
            .context("can not list images")
    }

    fn build_image(&self, project_name: &str, dockerfile: &str) -> Result<()> {
        let options = BuildOptions {
            tag: &format!("{}:latest", project_name),
            path: PathBuf::from("."),
            dockerfile,
            labels: HashMap::from([(PROJECT_KEY, project_name)]),
            ..Default::default()
        };
        self.cli.build(&options)
    }

    fn start_container(&self, name: &str) -> Result<()> {
        let runtime = Builder::new_current_thread().enable_all().build()?;
        let options: StartContainerOptions<String> = Default::default();
        runtime
            .block_on(self.client.start_container(name, Some(options)))
            .context("can not start container")
    }

    fn stop_container(&self, name: &str) -> Result<()> {
        let runtime = Builder::new_current_thread().enable_all().build()?;
        let options: StopContainerOptions = Default::default();
        runtime
            .block_on(self.client.stop_container(name, Some(options)))
            .context("can not stop container")
    }

    fn run(&self, name: &str, image_name: &str, deattach: bool, args: Vec<&str>) -> Result<()> {
        let options = RunOptions {
            name,
            image: image_name,
            deattach,
            args,
            ..Default::default()
        };
        self.cli.run(&options)
    }

    fn exec(&self, contain_name: &str, cmd: &[&str]) -> Result<()> {
        let options = ExecOptions {
            container: contain_name,
            args: Vec::from(cmd),
            ..Default::default()
        };
        self.cli.exec(&options)
    }
}

#[derive(Debug)]
struct BuildOptions<'a> {
    cmd: &'static str,
    tag: &'a str,
    dockerfile: &'a str,
    labels: HashMap<&'a str, &'a str>,
    path: PathBuf,
}

impl<'a> BuildOptions<'_> {
    fn build(&'a self) -> Vec<String> {
        let mut args = vec![
            self.cmd.to_string(),
            "-t".to_string(),
            self.tag.to_string(),
            "-f".to_string(),
            self.dockerfile.to_string(),
        ];
        args.push("--label".to_string());
        for (key, value) in &self.labels {
            args.push(format!("{}={}", key, value));
        }
        args.push(self.path.display().to_string());
        args
    }
}

impl Default for BuildOptions<'_> {
    fn default() -> Self {
        BuildOptions {
            cmd: "build",
            tag: Default::default(),
            dockerfile: ".devcontainer/Dockerfile",
            labels: Default::default(),
            path: PathBuf::from("."),
        }
    }
}

#[derive(Debug)]
struct RunOptions<'a> {
    cmd: &'static str,
    pub name: &'a str,
    pub deattach: bool,
    pub image: &'a str,
    pub volume: HashMap<&'a str, &'a str>,
    pub args: Vec<&'a str>,
}

impl RunOptions<'_> {
    fn build(&self) -> Vec<String> {
        let mut args = vec![
            self.cmd.to_string(),
            "--name".to_string(),
            self.name.to_string(),
        ];
        if self.deattach {
            args.push("-d".to_string())
        }

        for (key, value) in &self.volume {
            args.push("-v".to_string());
            args.push(format!("{}:{}", key, value));
        }

        args.push(self.image.to_string());
        args.extend(self.args.iter().map(|s| s.to_string()));
        args
    }
}

impl Default for RunOptions<'_> {
    fn default() -> Self {
        RunOptions {
            cmd: "run",
            name: Default::default(),
            deattach: false,
            image: Default::default(),
            volume: Default::default(),
            args: Default::default(),
        }
    }
}

pub struct ExecOptions<'a> {
    cmd: &'static str,
    pub container: &'a str,
    pub args: Vec<&'a str>,
}

impl ExecOptions<'_> {
    fn build(&self) -> Vec<String> {
        let mut args = vec![
            self.cmd.to_owned(),
            "-it".to_owned(),
            self.container.to_owned(),
        ];
        args.extend(self.args.iter().map(|s| s.to_string()));
        args
    }
}

impl Default for ExecOptions<'_> {
    fn default() -> Self {
        ExecOptions {
            cmd: "exec",
            container: Default::default(),
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

    fn run(&self, options: &RunOptions) -> Result<()> {
        process::pipe_cmd(&self.command, options.build())
    }

    fn exec(&self, options: &ExecOptions) -> Result<()> {
        process::pipe_cmd(&self.command, options.build())
    }
}

#[cfg(test)]
pub mod tests {
    use super::*;
    use mockall::mock;

    #[ignore]
    #[test]
    fn test_run() {
        let client = DockerClientImpl {
            client: Docker::connect_with_local_defaults().unwrap(),
            cli: DockerCli::new().unwrap(),
        };
        client
            .run("dev_space_test", "alpine", false, vec!["ls", "-l"])
            .unwrap();
    }

    #[ignore]
    #[test]
    fn test_list_container() {
        let client = DockerClientImpl {
            client: Docker::connect_with_local_defaults().unwrap(),
            cli: DockerCli::new().unwrap(),
        };
        let containers = client.list_containers("dev_space_test").unwrap();
        assert_eq!(containers.len(), 1);
        assert_eq!(containers[0].names.as_ref().unwrap()[0], "/dev_space_test");
    }

    mock! {
        pub DockerClient {}

        impl DockerClient for DockerClient {
            fn list_containers(&self, project_name: &str) -> Result<Vec<ContainerSummary>>;
            fn list_images(&self, project_name: &str) -> Result<Vec<ImageSummary>>;
            fn build_image(&self, project_name: &str, dockerfile: &str) -> Result<()>;
            fn start_container(&self, name: &str) -> Result<()>;
            fn stop_container(&self, name: &str) -> Result<()>;
            fn run<'a>(&self, name: &str, image_name: &str, deattach: bool, args: Vec<&'a str>) -> Result<()>;
            fn exec<'a>(&self, contain_name: &str, cmd: &[&'a str]) -> Result<()>;
        }
    }
}
