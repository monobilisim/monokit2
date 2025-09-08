# Development Guidelines

Please don't depend on CGO in this repository if its not necessary.
Use https://github.com/rs/zerolog for logging.
Use sqlite for agent storage but can use other database drivers for testing purposes.
For creating a new plugin please create a package under plugins directory and follow the existing plugin examples like osHealth.
