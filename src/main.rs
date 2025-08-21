mod cli;
mod docker_service;
mod service_manager;
mod commands;

use clap::Parser;
use cli::Cli;
use colored::*;
use anyhow::{Context, Result};

use service_manager::ServiceManager;

#[tokio::main]
async fn main() -> Result<()> {
    let cli = Cli::parse();

    let manager = ServiceManager::new(&cli.path).await.context("Failed to initialize service manager")?;
    match cli.command {
        cli::Commands::Up { services, build } => {
            println!("{}", "🚀 Starting services...".green().bold());
             manager.start_services(services, build).await?;
        },
        cli::Commands::Down { services, volumes } => {
            println!("{}", "🛑 Stopping services...".red().bold());
            manager.stop_services(services, volumes).await?;
        },
        cli::Commands::Restart { services } => {
            println!("{}", "🔄 Restarting services...".yellow().bold());
            manager.restart_services(services).await?;
        }
        cli::Commands::List { verbose} => {
            println!("{}", "📋 Discovered services:".blue().bold());
            manager.list_services(verbose).await?;
        } ,
        cli::Commands::Status {services, watch} => {
            if watch {
                println!("{}", "👀 Watching service status... (Ctrl+C to exit)".cyan().bold());
                manager.watch_status(services).await?;
            } else {
                println!("{}", "📊 Service status:".cyan().bold());
                manager.show_status(services).await?;
            }
        } ,
    }
    Ok(())
}
