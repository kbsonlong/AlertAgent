# Product Overview

## AlertAgent - 运维告警管理系统

AlertAgent is an operations alert management system built with Go (Gin) backend and React frontend, integrated with Ollama for intelligent alert analysis.

### Core Features

- **Alert Rule Management**: Configure alert rules with flexible trigger conditions
- **Alert Record Management**: Track alert events with status flow (new → acknowledged → resolved)
- **Notification Management**: Multi-channel notifications (email, SMS, webhook) with group management
- **Template Management**: Customizable notification templates with variable support
- **AI-Powered Analysis**: Ollama integration for intelligent alert analysis and solution recommendations

### Key Components

- **Backend**: RESTful API server handling alert processing, rule management, and AI analysis
- **Frontend**: React-based web interface for alert monitoring and management
- **Database**: MySQL for persistent storage with Redis for caching and queuing
- **AI Integration**: Ollama for local AI model inference and alert analysis
- **Queue System**: Redis-based task queue for asynchronous alert processing

### Target Users

Operations teams, DevOps engineers, and system administrators who need intelligent alert management and automated incident response capabilities.