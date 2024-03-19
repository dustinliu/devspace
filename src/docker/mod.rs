mod client;
mod container;
mod image;
mod process;
mod state;

use self::{
    container::{Container, ContainerFactory},
    image::Imagefactory,
};
use crate::project::Project;
use anyhow::Result;

pub struct DockerFactory {
    container: Box<dyn ContainerFactory>,
    image: Box<dyn Imagefactory>,
}

impl DockerFactory {
    pub fn new() -> Result<Self> {
        Ok(Self {
            container: container::new_factory()?,
            image: image::new_factory()?,
        })
    }

    pub fn get_container(&self, project_name: &str) -> Result<Container> {
        self.container.get_container(project_name)
    }

    pub fn get_image(&self, project: &Project) -> Result<Box<dyn image::Image>> {
        self.image.get_image(project)
    }
}
