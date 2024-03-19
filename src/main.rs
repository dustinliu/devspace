mod cli;
mod docker;
mod project;

use anyhow::Result;

// TODO: check if the error message is useer freindly
fn main() -> Result<()> {
    if let Err(e) = cli::run() {
        eprintln!("error: {:#}", e);
        return Err(e);
    }
    Ok(())
}
