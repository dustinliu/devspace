use anyhow::{Context, Result};
use std::{
    ffi::OsStr,
    fmt::Debug,
    process::{Command, Output, Stdio},
};

pub fn pipe_cmd<C, T, S>(cmd: C, args: T) -> Result<()>
where
    C: AsRef<OsStr>,
    T: IntoIterator<Item = S> + Debug,
    S: AsRef<OsStr>,
{
    dbg!(&args);
    println!("press a key to continue...");
    std::io::stdin().read_line(&mut String::new())?;
    let mut child = Command::new(&cmd)
        .args(args)
        .stdout(Stdio::inherit())
        .stderr(Stdio::inherit())
        .stdin(Stdio::piped())
        .spawn()?;

    match child.wait() {
        Ok(status) => {
            if status.success() {
                Ok(())
            } else {
                Err(anyhow::anyhow!("Command failed with status: {}", status))
            }
        }
        Err(e) => Err(e).with_context(|| anyhow::anyhow!("Failed to execute {:?}", cmd.as_ref())),
    }
}

fn run_cmd(cmd: &str, args: &[&str]) -> std::io::Result<Output> {
    return Command::new(cmd).args(args).output();
}
