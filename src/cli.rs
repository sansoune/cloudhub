use std::path::PathBuf;

use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(name = "cloudhub")]
#[command(about = "CloudHub CLI - Manage your containers", long_about = None)]
pub struct Cli {
    #[arg(short, long, default_value = "/opt/cloudhub")]
    pub path: PathBuf,

    #[command(subcommand)]
    pub command: Commands,
}

#[derive(Subcommand)]
pub enum Commands {
    Up {
        services: Vec<String>,
        #[arg(short, long)]
        build: bool,
    },
    Down {
        services: Vec<String>,
        #[arg(short, long)]
        volumes: bool,
    },
    Restart {
        services: Vec<String>,
    },
    List {
        #[arg(short, long)]
        verbose: bool,
    },
    Status {
        services: Vec<String>,
        #[arg(short, long)]
        watch: bool,
    },
}