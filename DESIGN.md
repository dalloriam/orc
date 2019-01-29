# ORC Personal Orchestrator

## Project Objectives
ORC is a daemon used to control the machine on which it runs while helping (as
automatically as possible with daily tasks)

## Feature Set
* Automatic management of docker services and tasks 
    (move from dockerfunc file defining bash functions to a simple shell 
    function performing an HTTP call to ORC,falling back on shell functions 
    if necessary). This would greatly simplify the invocation of tasks 
    from non-shell environments (e.g. triggering latex builds from VS Code).

* Automated non-DNS-based tracking of self-hosted services (datahose, plex,
    emby, etc.). Using ORC instead of DNS allows for supporting dynamic IPs as
    well as detecting whether we're running locally or not (ex. for resolving
    the plex server or the network shares).

* Extensibility via plugins.
    Want to be able to define custom tasks without recompiling and updating the
    service itself. This means that, at its core, ORC 



## Architecture

### Management Interfaces
* REST
    Receives commands via API and execute.
* Shell

### Plugins
Each plugin loaded by the application can be controlled in one of two ways,
either via shell or via network. When the plugin is shell-controlled, it must
specify its action < -- > call syntax mapping in the manifest. When the plugin is network-controlled,
it must similarly define its action < -- > request mapping in the manifest,
along with its service initialization command.
