# Kafka UI Project

## Overview
This project provides a web-based UI for managing and monitoring Apache Kafka clusters. It consists of a Go-based backend API and a React-based frontend. The backend handles Kafka operations and authentication, while the frontend offers a modern, user-friendly interface for interacting with Kafka topics, messages, brokers, and consumers.

## Project Structure

```
├── backend/         # Go backend API for Kafka operations and authentication
│   ├── api/         # API route handlers (topics, messages, brokers, auth)
│   ├── kafka/       # Kafka client logic and interfaces (using Sarama)
│   ├── middleware/  # JWT authentication and bootstrap middleware
│   ├── models/      # Data models (e.g., Topic)
│   ├── utils/       # Utility functions (CSV user management, responses)
│   ├── src/         # Main entry point (main.go)
│   └── data/        # User data (users.csv)
├── frontend/        # React frontend (UI)
│   ├── src/         # Source code (components, contexts, main app)
│   │   ├── components/ # React UI components
│   │   ├── contexts/   # React Contexts for shared state (e.g., MessageFormContext)
│   │   ├── App.js      # Main app component
│   │   ├── api.js      # Axios instance for API requests
│   │   └── ...         # Other source files
│   ├── public/      # Static assets
│   ├── nginx.conf   # Nginx config for deployment (if used)
│   └── ...          # Config files, dependencies
├── Makefile         # Cross-platform build commands
└── README.md        # Project documentation
```

## Backend (Go)
- **API Endpoints:**
  - `/api/login` – JWT login
  - `/api/check-connection` – Kafka connection check
  - `/api/topics` – List topics
  - `/api/topics/:name/messages` – Get messages
  - `/api/topics/:name/partitions` – Partition info
  - `/api/produce` – Produce message
  - `/api/topics/:name/messages` (DELETE) – Delete all messages
  - `/api/topics/:name` (DELETE) – Delete topic
  - `/api/topics` (POST) – Create topic
  - `/api/consumers` – List consumers
  - `/api/brokers` – List brokers
  - `/api/change-password` – Change user password
- **Authentication:** JWT-based, user data stored in `backend/data/users.csv`.
- **Kafka Integration:** Uses [Sarama](https://github.com/IBM/sarama) for all Kafka operations.
- **Config:**
  - Server port via `PORT` env var (default: `8080`)
  - Kafka broker address is configured dynamically via `bootstrapServer` query parameter
  - CORS is configured to allow requests from `http://localhost:3000`
  - Protected routes require JWT authentication and bootstrap server configuration

## Frontend (React)
- **Main Features:**
  - Login/logout with JWT
  - Dynamic Kafka broker configuration
  - View, create, and delete topics
  - View and produce messages
  - View brokers and consumers
  - Change password
- **Tech Stack:** React, Material UI, Axios
- **Start:**
  - `npm install` in `frontend/`
  - `npm start` in `frontend/` (runs on [http://localhost:3000](http://localhost:3000))

## Prerequisites

### Required Software
- **Go:** 1.24.1 or later
- **Node.js:** 18.0.0 or later (LTS recommended)
- **npm:** 8.0.0 or later (comes with Node.js)
- **Git:** 2.30.0 or later (optional but recommended)

## Installation & Setup

### Windows, macOS, Linux

1. **Install Required Software**
   - Go: [golang.org/dl/](https://golang.org/dl/)
   - Node.js: [nodejs.org/](https://nodejs.org/)
2. **Clone the repository**
   ```sh
   git clone https://github.com/AnishChhetry/kafka-ui-sarama.git
   cd kafka-ui-sarama
   ```
3. **Start Backend**
   ```sh
   cd backend
   go mod tidy
   go run src/main.go
   ```
4. **Start Frontend** (in a new terminal)
   ```sh
   cd frontend
   npm install
   npm start
   ```

## Usage
- Access the UI at [http://localhost:3000](http://localhost:3000)
- The backend runs on [http://localhost:8080](http://localhost:8080) by default
- Configure the Kafka broker address in the UI before using topic/message features

## Default Credentials
- **Username:** admin
- **Password:** password

## Useful Makefile Commands

| Command                     | Description                                      |
|-----------------------------|--------------------------------------------------|
| make install-deps           | Install Go and Node.js dependencies              |
| make install-backend-deps   | Install backend (Go) dependencies                |
| make install-frontend-deps  | Install frontend (React) dependencies            |
| make start                  | Start both backend and frontend servers          |
| make stop                   | Stop all running backend and frontend processes  |
