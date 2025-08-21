use crate::docker_service::{DockerService, ServiceStatus};
use anyhow::{Result, Context, anyhow};
use colored::*;
use std::collections::HashMap;
use std::path::PathBuf;
use tokio::time::{sleep, Duration};
use walkdir::WalkDir;

pub struct ServiceManager {
    cloudhub_path: PathBuf,
    services: HashMap<String, DockerService>,
    docker_compose_cmd: Vec<String>,
}

impl ServiceManager {
    pub async fn new(path: &PathBuf) -> Result<Self> {
        if !path.exists() {
            return Err(anyhow!("CloudHub path does not exist: {}", path.display()));
        }

        let docker_compose_cmd = Self::detect_docker_compose_command().await?;

        let mut manager = ServiceManager {
            cloudhub_path: path.clone(),
            services: HashMap::new(),
            docker_compose_cmd,
        };

        manager.discover_services().await?;
        Ok(manager)
    }

    async fn detect_docker_compose_command() -> Result<Vec<String>> {
        if let Ok(output) = tokio::process::Command::new("docker-compose")
            .arg("--version")
            .output()
            .await
        {
            if output.status.success() {
                println!("{}", "🔧 Using docker-compose (V1)".dimmed());
                return Ok(vec!["docker-compose".to_string()]);
            }
        }

        if let Ok(output) = tokio::process::Command::new("docker")
            .arg("compose")
            .arg("version")
            .output()
            .await
        {
            if output.status.success() {
                println!("{}", "🔧 Using docker compose (V2)".dimmed());
                return Ok(vec!["docker".to_string(), "compose".to_string()]);
            }
        }

        Err(anyhow!("Neither 'docker-compose' nor 'docker compose' command found. Please install Docker Compose."))
    }

    fn create_compose_command(&self) -> tokio::process::Command {
        let mut cmd = tokio::process::Command::new(&self.docker_compose_cmd[0]);
        
        for arg in &self.docker_compose_cmd[1..] {
            cmd.arg(arg);
        }
        
        cmd
    }

    async fn discover_services(&mut self) -> Result<()> {
        println!("{}", "🔍 Discovering services...".yellow());
        
        for entry in WalkDir::new(&self.cloudhub_path)
            .min_depth(1)
            .max_depth(2)
            .into_iter()
            .filter_map(|e| e.ok())
            .filter(|e| e.file_type().is_dir())
        {
            let path = entry.path().to_path_buf();
            
            if let Some(service) = DockerService::discover(&path).await? {
                println!("  {} Found service: {}", "✓".green(), service.name.cyan());
                self.services.insert(service.name.clone(), service);
            }
        }

        println!("{}", format!("📦 Discovered {} services", self.services.len()).green());
        Ok(())
    }

    pub async fn start_services(&self, service_names: Vec<String>, build: bool) -> Result<()> {
        let services_to_start = self.resolve_services(service_names)?;
        
        for service_name in services_to_start {
            let service = self.services.get(&service_name).unwrap();
            println!("🚀 Starting service: {}", service_name.cyan().bold());
            
            let mut cmd = self.create_compose_command();
            cmd.arg("-f")
               .arg(&service.compose_file)
               .current_dir(&service.path);

            if build || service.has_build_configs() {
                cmd.arg("up").arg("--build").arg("-d");
            } else {
                cmd.arg("up").arg("-d");
            }

            let output = cmd.output().await
                .context(format!("Failed to start service: {}", service_name))?;

            if output.status.success() {
                println!("  {} Service {} started successfully", "✅".green(), service_name.cyan());
            } else {
                let error = String::from_utf8_lossy(&output.stderr);
                println!("  {} Failed to start service {}: {}", "❌".red(), service_name.cyan(), error.red());
            }
        }

        Ok(())
    }

    pub async fn stop_services(&self, service_names: Vec<String>, remove_volumes: bool) -> Result<()> {
        let services_to_stop = self.resolve_services(service_names)?;
        
        for service_name in services_to_stop {
            let service = self.services.get(&service_name).unwrap();
            println!("🛑 Stopping service: {}", service_name.cyan().bold());
            
            let mut cmd = self.create_compose_command();
            cmd.arg("-f")
               .arg(&service.compose_file)
               .arg("down")
               .current_dir(&service.path);

            if remove_volumes {
                cmd.arg("-v");
            }

            let output = cmd.output().await
                .context(format!("Failed to stop service: {}", service_name))?;

            if output.status.success() {
                println!("  {} Service {} stopped successfully", "✅".green(), service_name.cyan());
            } else {
                let error = String::from_utf8_lossy(&output.stderr);
                println!("  {} Failed to stop service {}: {}", "❌".red(), service_name.cyan(), error.red());
            }
        }

        Ok(())
    }

    pub async fn restart_services(&self, service_names: Vec<String>) -> Result<()> {
        let services_to_restart = self.resolve_services(service_names)?;
        
        for service_name in services_to_restart {
            let service = self.services.get(&service_name).unwrap();
            println!("🔄 Restarting service: {}", service_name.cyan().bold());
            
            let output = self.create_compose_command()
                .arg("-f")
                .arg(&service.compose_file)
                .arg("restart")
                .current_dir(&service.path)
                .output()
                .await
                .context(format!("Failed to restart service: {}", service_name))?;

            if output.status.success() {
                println!("  {} Service {} restarted successfully", "✅".green(), service_name.cyan());
            } else {
                let error = String::from_utf8_lossy(&output.stderr);
                println!("  {} Failed to restart service {}: {}", "❌".red(), service_name.cyan(), error.red());
            }
        }

        Ok(())
    }

    pub async fn list_services(&self, verbose: bool) -> Result<()> {
        if self.services.is_empty() {
            println!("No services found in {}", self.cloudhub_path.display());
            return Ok(());
        }

        for (name, service) in &self.services {
            if verbose {
                println!("📦 {}", name.cyan().bold());
                println!("   Path: {}", service.path.display().to_string().dimmed());
                println!("   Compose: {}", service.compose_file.display().to_string().dimmed());
                println!("   Services: {}", service.get_service_names().join(", ").yellow());
                
                if service.config.networks.is_some() {
                    let networks: Vec<String> = service.config.networks.as_ref().unwrap().keys().cloned().collect();
                    println!("   Networks: {}", networks.join(", ").blue());
                }
                
                if service.config.secrets.is_some() {
                    let secrets: Vec<String> = service.config.secrets.as_ref().unwrap().keys().cloned().collect();
                    println!("   Secrets: {}", secrets.join(", ").purple());
                }
                println!();
            } else {
                println!("📦 {}", name.cyan());
            }
        }

        Ok(())
    }

    pub async fn show_status(&self, service_names: Vec<String>) -> Result<()> {
        let services_to_check = self.resolve_services(service_names)?;
        
        for service_name in services_to_check {
            let mut service = self.services.get(&service_name)
                .ok_or_else(|| anyhow!("Service not found: {}", service_name))?
                .clone();
            
            service.update_status().await?;
            println!("📊 {} - {}", service_name.cyan().bold(), service.status);
        }

        Ok(())
    }

    pub async fn watch_status(&self, service_names: Vec<String>) -> Result<()> {
        let services_to_watch = self.resolve_services(service_names)?;
        
        loop {
            print!("\x1B[2J\x1B[1;1H");
            println!("{}", "📊 Service Status (refreshing every 5s)".cyan().bold());
            println!("{}", "=".repeat(50).dimmed());
            
            for service_name in &services_to_watch {
                let mut service = self.services.get(service_name).unwrap().clone();
                service.update_status().await?;
                println!("{} - {}", service_name.cyan().bold(), service.status);
            }
            
            println!("\n{}", "Press Ctrl+C to exit".dimmed());
            sleep(Duration::from_secs(5)).await;
        }
    }

    pub async fn show_logs(&self, service_name: &str, follow: bool, tail: u32) -> Result<()> {
        let service = self.services.get(service_name)
            .ok_or_else(|| anyhow!("Service not found: {}", service_name))?;

        let mut cmd = self.create_compose_command();
        cmd.arg("-f")
           .arg(&service.compose_file)
           .arg("logs")
           .arg("--tail")
           .arg(tail.to_string())
           .current_dir(&service.path);

        if follow {
            cmd.arg("-f");
        }

        let mut child = cmd.spawn()
            .context(format!("Failed to show logs for service: {}", service_name))?;

        let status = child.wait().await
            .context("Failed to wait for logs command")?;

        if !status.success() {
            return Err(anyhow!("Logs command failed with status: {}", status));
        }

        Ok(())
    }

    fn resolve_services(&self, service_names: Vec<String>) -> Result<Vec<String>> {
        if service_names.is_empty() {
            Ok(self.services.keys().cloned().collect())
        } else {
            let mut resolved = Vec::new();
            for name in service_names {
                if self.services.contains_key(&name) {
                    resolved.push(name);
                } else {
                    return Err(anyhow!("Service '{}' not found. Available services: {}", 
                        name, 
                        self.services.keys().cloned().collect::<Vec<_>>().join(", ")
                    ));
                }
            }
            Ok(resolved)
        }
    }

    pub async fn refresh_services(&mut self) -> Result<()> {
        self.services.clear();
        self.discover_services().await
    }
}