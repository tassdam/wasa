# WASAText Messenger

WASAText is a web-based messaging application that provides both one-on-one and group chat functionalities along with features such as file attachments, message reactions, and message forwarding. The project is built on a Go backend (using SQLite as the datastore) and a Vue.js frontend.

## Features

- **Direct & Group Messaging:** Start private conversations or create groups.
- **File Attachments:** Send images and GIFs.
- **Message Reactions:** React to messages (e.g., like with a ❤️).
- **Forwarding & Replying:** Forward messages to other chats and reply with context.
- **Profile Management:** Update your username and profile photo.
- **User Search:** Find contacts by username.

## Technologies Used

### Backend
- **Go:** Core business logic and API server.
- **SQLite:** Embedded database.
- **[httprouter](https://github.com/julienschmidt/httprouter):** Lightweight routing.
- **[logrus](https://github.com/sirupsen/logrus):** Structured logging.
- **[uuid](https://github.com/gofrs/uuid):** Unique identifier generation.

### Frontend
- **Vue.js:** Reactive UI framework.
- **Vue Router:** SPA navigation.
- **Axios:** HTTP client for API calls.
- **Bootstrap & Custom CSS:** Responsive design 


## Setup & Running the Application on your local machine

### Backend

1. **Prerequisites:**
   - [Go](https://golang.org/dl/) (v1.16 or higher recommended)
   - SQLite

2. **Build and Run**

   Open a terminal in the project root and run:

   ```bash
   go run ./cmd/webapi/
   ```

   By default, the server listens on port `3000`. You can adjust settings (such as API host, database file, and timeouts) via command-line flags or by editing the configuration file (default location: `/conf/config.yml`).

### Frontend

1. **Prerequisites:**
   - [Node.js](https://nodejs.org/) (LTS version recommended)
   - npm or yarn

2. **Build:**

   Build dist with:

   ```bash
   yarn run build-prod
   ```

3. **Run:**


   ```bash
   yarn run preview
   ```

## Docker

The project includes Dockerfiles for both the backend and frontend. Use the following commands to build and run container images:

### Build Container Images

- **Backend:**

  ```bash
  docker build -t wasatext-backend:latest -f Dockerfile.backend .
  ```

- **Frontend:**

  ```bash
  docker build -t wasatext-frontend:latest -f Dockerfile.frontend .
  ```

### Run Container Images

- **Backend:**

  ```bash
  docker run -it --rm -p 3000:3000 wasatext-backend:latest
  ```

- **Frontend:**

  ```bash
  docker run -it --rm -p 8080:80 wasatext-frontend:latest
  ```



## Configuration

The backend configuration is read from command-line flags and an optional YAML file (default: `/conf/config.yml`). Use these settings to modify the API host, database location, read/write timeouts, and other parameters.
