use crate::{docker::DockerFactory, project::Project};
use anyhow::Result;

pub fn shell(root: &str) -> Result<()> {
    let project = Project::from(root)?;
    let factory = DockerFactory::new()?;
    let container = factory.get_container(project.name.as_str())?;

    if !container.existing() {}

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::project::tests::TmpProjectDir;
}
