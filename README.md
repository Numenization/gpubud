# gpubud

Tool for monitoring GPU prices and availability from certain retailers.

The current database information can be viewed on the web client frontend, and notifications can be set up with the Discord bot 
to send messages to a Discord channel whenever a GPU that is being monitored by that channel is updated.

Server backend in Go HTTP and sqlite, web frontend made with HTMX, and Discord bot made with discordgo.