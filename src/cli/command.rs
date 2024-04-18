use crate::{docker::Container, project::Project};
use anyhow::Result;

pub fn shell(root: &str) -> Result<()> {
    let project = Project::try_from(root)?;
    let container: Container = Container::try_from(&project)?;

    if !container.existing() {
        container.setup()?;
    } else if !container.running() {
        container.start()?;
    }

    container.exec(&["/bin/zsh"])?;

    Ok(())
}

#[cfg(test)]
mod tests {}
