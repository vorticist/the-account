# the-account

## Local Env

### Prerequisites
 - [Go](https://go.dev/doc/install)
 - [docker](https://docs.docker.com/engine/install/)

### Run
- Run the [menu_analyzer](https://github.com/saianfordx/menu_image_analyzer) project separately on localhost:8000
- Setup local mongodb using docker
  - ```shell
    docker run -e MONGO_INITDB_ROOT_USERNAME=admin -e MONGO_INITDB_ROOT_PASSWORD=secretpassword -p 27017:27017 --name some-mongo -d mongo:latest
    ``` 
  - Make note of the username and password used here
  

- Clone the repository and cd into the directory
  - ```shell
      git clone git@github.com:vorticist/the-account.git
      cd the-account
    ```
- Setup the local environment variables
  - ```shell
    export MONGODB_CONN_STRING="mongodb://admin:secretpassword@localhost:27017"
    export MENU_ANALYZER_URL="http://localhost:8000/analyze-menu"
    export OPENAI_API_KEY="sk-..."
    ```
- Run the project
  - ```shell
    go run .
    ```
- The server should be running on [localhost:9090/admin](http://localhost:9090/admin)