use bollard::models::ContainerSummary;

use super::client::DockerClient;
use crate::{
    docker::{client::new_client, image::new_image, Image},
    project::{Config, Project},
};
use anyhow::{anyhow, Result};
use std::fmt;

pub struct Container<'a> {
    name: &'a str,
    config: &'a Config,
    summary: Option<ContainerSummary>,
    client: Box<dyn DockerClient>,
}

impl<'a> Container<'a> {
    fn new(
        name: &'a str,
        config: &'a Config,
        client: Box<dyn DockerClient>,
    ) -> Result<Container<'a>> {
        let containers = client.list_containers(name)?;

        match containers.len() {
            0 => Ok(Container {
                name,
                config,
                summary: None,
                client,
            }),
            1 => Ok(Container {
                name,
                config,
                summary: Some(containers[0].clone()),
                client,
            }),
            _ => Err(anyhow!("Multiple containers with the same name")),
        }
    }

    pub fn existing(&self) -> bool {
        self.summary.is_some()
    }

    pub fn running(&self) -> bool {
        self.summary
            .as_ref()
            .is_some_and(|s| s.state.as_ref().is_some_and(|s| s == "running"))
    }

    pub fn setup(&self) -> Result<()> {
        let mut image: Box<dyn Image> =
            new_image(self.name, &self.config.image_source, self.client.as_ref())?;

        if !image.existing() {
            image.build(self.client.as_ref())?;
        }

        self.client
            .run(self.name, image.name(), true, vec!["sleep", "infinity"])?;

        if let Some(command) = &self.config.post_create_command {
            let c = command.iter().map(|s| s.as_str()).collect::<Vec<_>>();
            self.client.exec(self.name, &c)?;
        }

        Ok(())
    }

    pub fn exec(&self, cmd: &[&str]) -> Result<()> {
        self.client.exec(self.name, cmd)
    }

    pub fn start(&self) -> Result<()> {
        self.client.start_container(self.name)
    }
}

impl<'a> TryFrom<&'a Project> for Container<'a> {
    type Error = anyhow::Error;

    fn try_from(p: &'a Project) -> Result<Self> {
        Container::new(&p.name, &p.config, new_client()?)
    }
}

impl fmt::Debug for Container<'_> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("Container")
            .field("name", &self.name)
            .field("config", &self.config)
            .field("summary", &self.summary)
            .finish()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::docker::client::tests::MockDockerClient;
    use bollard::models::ContainerSummary;

    #[test]
    fn test_get_container_existing() {
        let mut client = MockDockerClient::new();
        client.expect_list_containers().returning(|_| {
            Ok(vec![ContainerSummary {
                names: Some(vec!["aaa".to_string()]),
                state: Some("running".to_string()),
                ..Default::default()
            }])
        });

        let project = Project {
            name: "aaa".to_string(),
            ..Default::default()
        };

        let container = Container::new(&project.name, &project.config, Box::new(client)).unwrap();
        assert!(container.existing());
        assert!(container.running());
        assert_eq!(container.name, "aaa");
    }

    #[test]
    fn test_get_container_existing_not_running() {
        let mut client = MockDockerClient::new();
        client.expect_list_containers().returning(|_| {
            Ok(vec![ContainerSummary {
                names: Some(vec!["bbb".to_string()]),
                state: Some("Exited".to_string()),
                ..Default::default()
            }])
        });

        let project = Project {
            name: "bbb".to_string(),
            ..Default::default()
        };

        let container = Container::new(&project.name, &project.config, Box::new(client)).unwrap();
        assert!(container.existing());
        assert!(!container.running());
        assert_eq!(container.name, "bbb");
    }

    #[test]
    fn test_get_container_existing_multiple() {
        let mut client = MockDockerClient::new();
        client.expect_list_containers().returning(|_| {
            Ok(vec![
                ContainerSummary {
                    names: Some(vec!["bbb".to_string()]),
                    state: Some("Exited".to_string()),
                    ..Default::default()
                },
                ContainerSummary {
                    names: Some(vec!["zzz".to_string()]),
                    state: Some("Exited".to_string()),
                    ..Default::default()
                },
            ])
        });

        let project = Project {
            name: "bbb".to_string(),
            ..Default::default()
        };

        let err = Container::new(&project.name, &project.config, Box::new(client)).unwrap_err();
        assert_eq!(
            format!("{:?}", err),
            "Multiple containers with the same name"
        );
    }

    #[test]
    fn test_get_container_not_existing() {
        let mut client = MockDockerClient::new();
        client.expect_list_containers().returning(|_| Ok(vec![]));

        let project = Project {
            name: "bbb".to_string(),
            ..Default::default()
        };

        let container = Container::new("ccc", &project.config, Box::new(client)).unwrap();
        assert!(!container.existing());
        assert!(!container.running());
        assert_eq!(container.name, "ccc");
    }
}
