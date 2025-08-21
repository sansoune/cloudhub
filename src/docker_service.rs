use serde::{Deserialize, Serialize};
use std::path::PathBuf;
use std::collections::HashMap;
use anyhow::{Result, Context};
use anyhow::anyhow;
use tokio::fs;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DockerService {
    pub name: String,
    pub path: PathBuf,
    pub compose_file: PathBuf,
    pub config: DockerComposeConfig,
    pub status: ServiceStatus,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DockerComposeConfig {
    pub services: HashMap<String, ServiceConfig>,
    pub networks: Option<HashMap<String, NetworkConfig>>,
    pub volumes: Option<HashMap<String, VolumeConfig>>,
    pub secrets: Option<HashMap<String, SecretConfig>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ServiceConfig {
    pub image: Option<String>,
    pub build: Option<BuildConfig>,
    pub ports: Option<Vec<String>>,
    pub environment: Option<HashMap<String, String>>,
    pub volumes: Option<Vec<String>>,
    pub networks: Option<Vec<String>>,
    pub depends_on: Option<Vec<String>>,
    pub restart: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum BuildConfig {
    Simple(String),
    Complex {
        context: String,
        dockerfile: Option<String>,
        args: Option<HashMap<String, String>>,
    },
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NetworkConfig {
    pub driver: Option<String>,
    pub external: Option<bool>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VolumeConfig {
    pub driver: Option<String>,
    pub external: Option<bool>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SecretConfig {
    pub file: Option<String>,
    pub external: Option<bool>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum ServiceStatus {
    Unknown,
    Running,
    Stopped,
    Error(String),
    Starting,
    Stopping,
}

impl DockerService {
    pub async fn discover(path: &PathBuf) -> Result<Option<Self>> {
        let compose_files = ["docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"];
        
        for filename in &compose_files {
            let compose_path = path.join(filename);
            if compose_path.exists() {
                let content = fs::read_to_string(&compose_path).await
                    .context(format!("Failed to read compose file: {}", compose_path.display()))?;
                
                let config: DockerComposeConfig = serde_yaml::from_str(&content)
                    .context(format!("Failed to parse compose file: {}", compose_path.display()))?;
                
                let service_name = path.file_name()
                    .and_then(|n| n.to_str())
                    .unwrap_or("unknown")
                    .to_string();
                
                return Ok(Some(DockerService {
                    name: service_name,
                    path: path.clone(),
                    compose_file: compose_path,
                    config,
                    status: ServiceStatus::Unknown,
                }));
            }
        }
        
        Ok(None)
    }

    pub async fn update_status(&mut self) -> Result<()> {
        let docker_compose_cmd = Self::detect_docker_compose_command().await?;
        let mut cmd = tokio::process::Command::new(&docker_compose_cmd[0]);
        
        for arg in &docker_compose_cmd[1..] {
            cmd.arg(arg);
        }
        
        let output = cmd
            .arg("-f")
            .arg(&self.compose_file)
            .arg("ps")
            .arg("-q")
            .current_dir(&self.path)
            .output()
            .await
            .context("Failed to execute docker compose ps")?;

        if !output.status.success() {
            self.status = ServiceStatus::Error(
                String::from_utf8_lossy(&output.stderr).to_string()
            );
            return Ok(());
        }

        let container_ids = String::from_utf8_lossy(&output.stdout);
        if container_ids.trim().is_empty() {
            self.status = ServiceStatus::Stopped;
            return Ok(());
        }

        for container_id in container_ids.lines() {
            if container_id.trim().is_empty() {
                continue;
            }

            let inspect_output = tokio::process::Command::new("docker")
                .arg("inspect")
                .arg("--format")
                .arg("{{.State.Status}}")
                .arg(container_id.trim())
                .output()
                .await
                .context("Failed to inspect container")?;

            if inspect_output.status.success() {
                let status = String::from_utf8_lossy(&inspect_output.stdout).trim().to_string();
                match status.as_str() {
                    "running" => self.status = ServiceStatus::Running,
                    "exited" => self.status = ServiceStatus::Stopped,
                    "restarting" => self.status = ServiceStatus::Starting,
                    _ => self.status = ServiceStatus::Unknown,
                }
            }
        }

        Ok(())
    }

    async fn detect_docker_compose_command() -> Result<Vec<String>> {
        // Try docker-compose first (V1)
        if let Ok(output) = tokio::process::Command::new("docker-compose")
            .arg("--version")
            .output()
            .await
        {
            if output.status.success() {
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
                return Ok(vec!["docker".to_string(), "compose".to_string()]);
            }
        }

        Err(anyhow!("Neither 'docker-compose' nor 'docker compose' command found"))
    }

    pub fn get_service_names(&self) -> Vec<String> {
        self.config.services.keys().cloned().collect()
    }

    pub fn has_build_configs(&self) -> bool {
        self.config.services.values().any(|s| s.build.is_some())
    }
}

impl std::fmt::Display for ServiceStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ServiceStatus::Unknown => write!(f, "❓ Unknown"),
            ServiceStatus::Running => write!(f, "✅ Running"),
            ServiceStatus::Stopped => write!(f, "⭕ Stopped"),
            ServiceStatus::Error(msg) => write!(f, "❌ Error: {}", msg),
            ServiceStatus::Starting => write!(f, "🟡 Starting"),
            ServiceStatus::Stopping => write!(f, "🟠 Stopping"),
        }
    }
}