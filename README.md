# ShareWay Backend

## Prerequisites

- Go >= 1.21.3 ([Installation Guide](https://go.dev/doc/install))
- Node.js >= v20.6.1 ([Download](https://nodejs.org/en/download/))
- GNU Make >= 3.81 ([Documentation](https://www.gnu.org/software/make/))

## Development Setup

1. Install nodemon globally for auto-refresh functionality:
	```bash
	   # Install nodemon globally
	   npm install -g nodemon
	   
	   # Verify installation
	   nodemon --version

	```

2.  Clone and setup the repository:

     ```bash
	    # Clone the repository
	    git clone https://github.com/crazyads69/shareway-be.git
    
		# Navigate to project directory
		cd shareway-be
    
		# Install Go dependencies
		go mod download
	  ```
    
4.  Configure the environment:
    
    -   Create `app.env` file in the project root
    -   Add required environment variables

5.  Start the development server:
    ```bash
    
	    make dev_server
    
    ```
    
6.  Verify the server is running:
    
    -   Visit [http://127.0.0.1:8080](http://127.0.0.1:8080) in your browser
    -   Or use cURL:
	 ```bash
        
        curl http://127.0.0.1:8080
        
	```
        

## Additional Information

-   The server will automatically refresh when code changes are detected thanks to nodemon
-   Default server port is `8080`
-   Make sure all prerequisites are properly installed and accessible from your `PATH`
