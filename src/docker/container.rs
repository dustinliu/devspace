use bollard::models::ContainerSummary;

use super::client::DockerClient;
use anyhow::{anyhow, Result};

// pub trait Container {
//     fn name(&self) -> &str;
//     fn is_existing(&self) -> bool;
//     fn is_running(&self) -> bool;
// }

pub struct Container {
    name: String,
    summary: Option<ContainerSummary>,
}

impl Container {
    pub fn name(&self) -> &str {
        &self.name
    }

    pub fn existing(&self) -> bool {
        self.summary.is_some()
    }

    pub fn running(&self) -> bool {
        self.summary
            .as_ref()
            .is_some_and(|s| s.state.as_ref().is_some_and(|s| s == "running"))
    }
}

pub trait ContainerFactory {
    fn get_container(&self, name: &str) -> Result<Container>;
}

pub fn new_factory() -> Result<Box<dyn ContainerFactory>> {
    Ok(Box::new(ContainerFactoryImpl {
        client: crate::docker::client::new_client()?,
    }))
}

struct ContainerFactoryImpl {
    client: Box<dyn DockerClient>,
}

impl ContainerFactory for ContainerFactoryImpl {
    fn get_container(&self, name: &str) -> Result<Container> {
        let containers = self.client.list_containers(name)?;

        match containers.len() {
            0 => Ok(Container {
                name: name.to_string(),
                summary: None,
            }),
            1 => Ok(Container {
                name: name.to_string(),
                summary: Some(containers[0].clone()),
            }),
            _ => Err(anyhow!("Multiple containers with the same name")),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::docker::client::MockDockerClient;
    use bollard::models::ContainerSummary;

    #[test]
    fn test_get_container_existing() {
        let mut client = MockDockerClient::new();
        client.expect_list_containers().returning(|_| {
            Ok(vec![ContainerSummary {
                names: Some(vec!["test".to_string()]),
                state: Some("running".to_string()),
                ..Default::default()
            }])
        });
        let factory = ContainerFactoryImpl {
            client: Box::new(client),
        };

        let container = factory.get_container("test").unwrap();
        assert!(container.existing());
        assert!(container.running());
        assert_eq!(container.name(), "test");
    }

    #[test]
    fn test_get_container_existing_not_running() {
        let mut client = MockDockerClient::new();
        client.expect_list_containers().returning(|_| {
            Ok(vec![ContainerSummary {
                names: Some(vec!["test".to_string()]),
                state: Some("Exited".to_string()),
                ..Default::default()
            }])
        });
        let factory = ContainerFactoryImpl {
            client: Box::new(client),
        };

        let container = factory.get_container("test").unwrap();
        assert!(container.existing());
        assert!(!container.running());
        assert_eq!(container.name(), "test");
    }

    #[test]
    fn test_get_container_not_existing() {
        let mut client = MockDockerClient::new();
        client.expect_list_containers().returning(|_| Ok(vec![]));
        let factory = ContainerFactoryImpl {
            client: Box::new(client),
        };

        let container = factory.get_container("test").unwrap();
        assert!(!container.existing());
        assert!(!container.running());
        assert_eq!(container.name(), "test");
    }
}
