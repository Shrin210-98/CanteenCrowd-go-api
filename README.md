# CanteenCrowd-go-api

This project is a Go-based API application that interacts with a PostgreSQL database. It is designed to be deployed on AWS EC2 and utilizes Nginx as a reverse proxy. The project also includes CI/CD integration with GitHub Actions for automated testing and deployment.

## Project Structure

```
go-api
├── cmd
│   └── api
│       └── main.go          # Entry point of the application
├── internal
│   ├── database
│   │   └── queries.go      # Database query functions
│   ├── handlers
│   │   └── handler.go      # HTTP handlers for processing requests
│   └── server
│       └── server.go       # HTTP server setup and middleware
├── .github
│   └── workflows
│       └── ci.yml          # CI/CD workflow for GitHub Actions
├── nginx
│   └── nginx.conf          # Nginx configuration for the application
├── .env                    # Environment variables for the application
├── Makefile                # Build instructions for the application
├── go.mod                  # Module definition and dependencies
├── go.sum                  # Checksums for module dependencies
└── README.md               # Project documentation
```

## Setup Instructions

1. **Clone the Repository**
   ```bash
   git clone <repository-url>
   cd go-api
   ```

2. **Install Dependencies**
   Ensure you have Go installed, then run:
   ```bash
   go mod tidy
   ```

3. **Configure Environment Variables**
   Create a `.env` file in the root directory and set the necessary environment variables for your application, including database connection details.

4. **Build the Application**
   Use the Makefile to build the application:
   ```bash
   make build
   ```

5. **Run the Application**
   You can run the application using:
   ```bash
   make run
   ```

6. **Testing**
   To run tests, use:
   ```bash
   make test
   ```

## Deployment

This application is designed to be deployed on AWS EC2. Ensure that you have configured your EC2 instance with the necessary security groups and IAM roles to allow access to RDS and other AWS services.

## CI/CD Integration

The project includes a GitHub Actions workflow defined in `.github/workflows/ci.yml` that automates the build, test, and deployment process. Make sure to configure your GitHub repository settings to enable Actions.

## Nginx Configuration

The Nginx configuration file located in `nginx/nginx.conf` sets up a reverse proxy to forward requests to the Go application. Ensure that Nginx is installed and configured on your server.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.