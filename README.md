[![Gitter](https://img.shields.io/gitter/room/nwjs/nw.js.svg)](https://gitter.im/gnode-gin/Lobby)
[![Docker Automated build](https://img.shields.io/docker/automated/gnode/gogs.svg)](https://hub.docker.com/r/gnode/gogs/builds/)
[![Build Status](https://travis-ci.org/G-Node/gogs.svg?branch=master)](https://travis-ci.org/G-Node/gogs)

# About Gin Gogs

gin-gogs is the web interface for the **GIN** (**G**-Node **IN**frastructure) services.

## What is gin?
Management of scientific data, including consistent organization, annotation and storage of data, is a challenging task. Accessing and managing data from multiple workplaces while keeping it in sync, backed up, and easily accessible from within or outside the lab is even more demanding.

To minimize the time and effort scientists have to spend on these tasks, we develop the GIN (G-Node Infrastructure) services, a free data management system designed for comprehensive and reproducible management of scientific data.

## Why should I use GIN?
### Manage your data from anywhere
* Upload your data on a repository based structure: you can create as many repositories as you like.
* Access your data from anywhere: once the data is at the main repository service you can securely access your data from anywhere you like.
* Synchronize your data: you can download complete or partial repositories on any workplace you like, work on them locally and upload the changes back to the main repository.

### Version your data
* When changing your files and uploading them to the server, the history is automatically kept, you can always go back to a previous version.


### Share your data
* Make your data public: if you want to make your data accessible to the world, just make your repository publicly available. The data will be accessible but only you will be able to change it.
* Share your data with collaborators: you can also share repositories with other users of the GIN service making it easy to jointly work on a project.
* Make your data citable: through the gin DOI service you can obtain registered identifiers for your public datasets.


### Choose how you want to use our service
* Register with the GIN services and use the provided infrastructure.
* Set up and host your own in-house instance - our software is open source, you can use it for free.

### Enhanced search of your repositories **in development**
By indexing the repository contents it's easy to find the files you are looking for. When using the [NIX](http://www.g-node.org/nix) data format for scientific data and metadata, even the contents of these files will be indexed and searchable, making it easy for you to identify the data you are looking for.

## Acknowledgments
GIN is based on [Gogs](https://github.com/gogits/gogs)

## License

This project is under the MIT License. See the [LICENSE](https://github.com/G-Node/gogs/blob/master/LICENSE) file for the full license text.
