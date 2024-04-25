use crate::{docker::Container, project::Project};
use anyhow::Result;

pub fn shell(root: &str, stop: &bool) -> Result<()> {
    let project = Project::try_from(root)?;
    let container: Container = Container::try_from(&project)?;

    if !container.existing() {
        println!("container does not exist, creating...");
        container.setup()?;
    } else if !container.running() {
        println!("container is not running, starting...");
        container.start()?;
    }

    container.exec(&["/bin/zsh"])?;

    if *stop {
        container.stop()?;
    }

    Ok(())
}

#[cfg(test)]
mod tests {}
