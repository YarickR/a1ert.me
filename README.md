## Directed Acyclic Graph Processor

What does it do ? It processes events, fetching them using various plugins, according to rules specified in channels . 

### Modules 

Currently include: 
 - core
 - redis
 - xmpp 

Since Go has no means to dynamically load libaries at runtime, every module has to be built in 

### Plugins

Plugins implement fetching events from various external sources, processing them, possibly using external tools, and sending them to various external destinations. Each plugin is using one of a built-in modules, and has an id, one or more hooks, and plugin-specific global configuration.

### Channels

Each channel has a list of plugins attached to it. Each plugin could have per-channel config section, which is merged with global plugin config , resulting config defines how this plugin will process message entering the channel 



### Rules

Property of a core plugin. Set of conditions 

### Templates

Used by processing and output plugins, determining 