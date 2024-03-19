use crate::project::{ImageSource, Project};
use anyhow::Result;
use bollard::models::ImageSummary;

use super::client::DockerClient;

pub trait Image {
    fn name(&self) -> &str;
    fn is_existing(&self) -> bool;
}

trait BuildableImage {
    fn build(&self, client: &impl DockerClient) -> Result<()>;
}

struct ForeignImage {
    name: String,
}

impl Image for ForeignImage {
    fn name(&self) -> &str {
        &self.name
    }

    fn is_existing(&self) -> bool {
        true
    }
}

struct DockrefileImage {
    project_name: String,
    dockerfile: String,
    summary: Option<ImageSummary>,
}

impl Image for DockrefileImage {
    fn name(&self) -> &str {
        &self.project_name
    }

    fn is_existing(&self) -> bool {
        self.summary.is_some()
    }
}

impl BuildableImage for DockrefileImage {
    fn build(&self, client: &impl DockerClient) -> Result<()> {
        client.build_image(&self.project_name, &self.dockerfile)
    }
}

pub trait Imagefactory {
    fn get_image(&self, project: &Project) -> Result<Box<dyn Image>>;
}

pub fn new_factory() -> Result<Box<dyn Imagefactory>> {
    Ok(Box::new(ImageFactoryImpl {
        client: crate::docker::client::new_client()?,
    }))
}

struct ImageFactoryImpl {
    client: Box<dyn DockerClient>,
}

impl ImageFactoryImpl {
    fn load_dockerfile_image(
        &self,
        project_name: &str,
        dockerfile: &str,
    ) -> Result<Box<dyn Image>> {
        let images = self.client.list_images(project_name)?;
        let name = format!("{}:latest", project_name);
        let summary = images.iter().find(|i| i.repo_tags.contains(&name));

        Ok(Box::new(DockrefileImage {
            project_name: project_name.to_string(),
            dockerfile: dockerfile.to_string(),
            summary: summary.cloned(),
        }))
    }
}

impl Imagefactory for ImageFactoryImpl {
    fn get_image(&self, project: &Project) -> Result<Box<dyn Image>> {
        match &project.config.image_source {
            ImageSource::Image(name) => Ok(Box::new(ForeignImage {
                name: name.to_owned(),
            })),
            ImageSource::Dockerfile(dockerfile) => {
                self.load_dockerfile_image(&project.name, dockerfile)
            }
        }
    }
}

mod buildable {
    use super::*;
    use crate::docker::client::DockerClient;
    use anyhow::Result;

    trait BuildableImage: Image {
        fn build(&self, client: &impl DockerClient) -> Result<()>;
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::{docker::client::MockDockerClient, project::tests::TmpProjectDir};

    #[test]
    fn test_foreign_image() {
        let json = r#"
        {
            "name": "aaa",
            "image": "testimage:latest"
        }"#;
        let root = TmpProjectDir::new(json).devcontainer_json(json);
        let project = Project::from(root.root).unwrap();

        let mock_client = MockDockerClient::new();
        let factory = ImageFactoryImpl {
            client: Box::new(mock_client),
        };

        let image = factory.get_image(&project).unwrap();
        assert!(image.is_existing());
        assert_eq!(image.name(), "testimage:latest");
    }

    #[test]
    fn teset_dockerfile_image_not_existing() {
        let json = r#"
        {
            "name": "aaa",
            "dockerFile": "Dockerfile.test"
        }"#;
        let root = TmpProjectDir::new(json).devcontainer_json(json);
        let project = Project::from(root.root).unwrap();

        let mut mock_client = MockDockerClient::new();
        mock_client.expect_list_images().returning(|_| Ok(vec![]));
        let factory = ImageFactoryImpl {
            client: Box::new(mock_client),
        };

        let image = factory.get_image(&project).unwrap();
        assert!(!image.is_existing());
        assert_eq!(image.name(), "aaa");
    }

    #[test]
    fn teset_dockerfile_image_existing() {
        let json = r#"
        {
            "name": "bbb",
            "dockerFile": "Dockerfile.test2"
        }"#;
        let root = TmpProjectDir::new("yyyy").devcontainer_json(json);
        let project = Project::from(root.root).unwrap();

        let mut mock_client = MockDockerClient::new();
        mock_client.expect_list_images().returning(|_| {
            Ok(vec![ImageSummary {
                repo_tags: vec!["bbb:latest".to_string()],
                ..Default::default()
            }])
        });
        let factory = ImageFactoryImpl {
            client: Box::new(mock_client),
        };

        let image = factory.get_image(&project).unwrap();
        assert!(image.is_existing());
        assert_eq!(image.name(), "bbb");
    }
}
