# Golang Template

## Quick start

Prerequisite:

- [Golang](https://golang.org/doc/install) (>= 1.21.3)
- [NodeJs](https://nodejs.org/en/download/) (>= v20.6.1)
- [Make](https://www.gnu.org/software/make/) (>= 3.81)

Install nodemon to auto refresh the server when code changed:

```bash
npx install -g nodemon

# Verify installation version
nodemon --version
```

Clone this repo and install dependencies:

1. Clone this repo

    ```bash
    git clone https://github.com/khiemledev/golang_template.git
    ```

2. Change directory into `golang_template`

    ```bash
    cd golang_template
    ```

3. Install dependencies Go dependencies

    ```bash
    go mod download
    ```

4. Start dev server

    ```bash
    make dev_server
    ```

5. Now you can vist [http://127.0.0.1:8080](http://127.0.0.1:8080) to check the server is running
    or using cURL

    ```bash
    curl http://127.0.0.1:8080
    ```

