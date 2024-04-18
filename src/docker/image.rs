use crate::{docker::client::DockerClient, project::ImageSource};
use anyhow::{anyhow, Result};
use bollard::models::ImageSummary;

pub trait Image {
    fn name(&self) -> &str;
    fn existing(&self) -> bool;
    fn build(&mut self, client: &dyn DockerClient) -> Result<()>;
}

pub fn new_image(
    project_name: &str,
    source: &ImageSource,
    client: &dyn DockerClient,
) -> Result<Box<dyn Image>> {
    match source {
        ImageSource::Image(name) => Ok(Box::new(ForeignImage {
            name: name.to_owned(),
        })),
        ImageSource::Dockerfile(dockerfile) => {
            let images = client.list_images(project_name)?;
            let name = format!("{}:latest", project_name);
            let summary = images.iter().find(|i| i.repo_tags.contains(&name));

            Ok(Box::new(DockrefileImage {
                project_name: project_name.to_string(),
                dockerfile: dockerfile.to_string(),
                summary: summary.cloned(),
            }))
        }
    }
}

struct ForeignImage {
    name: String,
}

impl Image for ForeignImage {
    fn name(&self) -> &str {
        &self.name
    }

    fn existing(&self) -> bool {
        true
    }

    fn build(&mut self, _: &dyn DockerClient) -> Result<()> {
        Ok(())
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

    fn existing(&self) -> bool {
        self.summary.is_some()
    }

    fn build(&mut self, client: &dyn DockerClient) -> Result<()> {
        client.build_image(&self.project_name, &self.dockerfile)?;
        let summaries = client.list_images(&self.project_name)?;
        let tag = format!("{}:latest", self.project_name);
        self.summary = summaries.into_iter().find(|i| i.repo_tags.contains(&tag));

        if self.summary.is_none() {
            return Err(anyhow!("Image not found after build"));
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::{
        docker::client::tests::MockDockerClient,
        project::{tests::TmpProjectDir, Project},
    };

    #[test]
    fn test_foreign_image() {
        let json = r#"
        {
            "name": "aaa",
            "image": "testimage:latest"
        }"#;
        let tmp_project_dir = TmpProjectDir::new(json).devcontainer_json(json);
        let project = Project::try_from(&tmp_project_dir.root).unwrap();
        let client = MockDockerClient::new();

        let image: Box<dyn Image> =
            new_image(&project.name, &project.config.image_source, &client).unwrap();
        assert!(image.existing());
        assert_eq!(image.name(), "testimage:latest");
    }

    #[test]
    fn test_dockerfile_image_not_existing() {
        let json = r#"
        {
            "name": "aaa",
            "dockerFile": "Dockerfile.test"
        }"#;
        let tmp_project_dir = TmpProjectDir::new(json).devcontainer_json(json);
        let project = Project::try_from(&tmp_project_dir.root).unwrap();

        let mut mock_client = MockDockerClient::new();
        mock_client.expect_list_images().returning(|_| Ok(vec![]));

        let image: Box<dyn Image> =
            new_image(&project.name, &project.config.image_source, &mock_client).unwrap();
        assert!(!image.existing());
        assert_eq!(image.name(), "aaa");
    }

    #[test]
    fn test_dockerfile_image_existing() {
        let json = r#"
        {
            "name": "bbb",
            "dockerFile": "Dockerfile.test2"
        }"#;
        let tmp_project_dir = TmpProjectDir::new("yyyy").devcontainer_json(json);
        let project = Project::try_from(&tmp_project_dir.root).unwrap();

        let mut mock_client = MockDockerClient::new();
        mock_client.expect_list_images().returning(|_| {
            Ok(vec![ImageSummary {
                repo_tags: vec!["bbb:latest".to_string()],
                ..Default::default()
            }])
        });

        let image: Box<dyn Image> =
            new_image(&project.name, &project.config.image_source, &mock_client).unwrap();
        assert!(image.existing());
        assert_eq!(image.name(), "bbb");
    }
}
