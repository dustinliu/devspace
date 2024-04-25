mod command;

use anyhow::Result;
use clap::Parser;
use clap::Subcommand;

#[derive(Parser)]
#[command(version, about, long_about = None)]
#[command(propagate_version = true)]
pub struct Cli {
    #[command(subcommand)]
    pub cmds: Commands,

    /// project root, default "."
    #[arg(long, global = true, default_value = ".")]
    root: String,
}

#[derive(Subcommand)]
pub enum Commands {
    Shell {
        #[arg(from_global)]
        root: String,

        /// stop container after shell exits
        #[arg(short, long)]
        stop: bool,
    },
}

pub fn run() -> Result<()> {
    let root_cmd = Cli::parse();

    match &root_cmd.cmds {
        Commands::Shell { root, stop } => command::shell(root, stop),
    }
}
