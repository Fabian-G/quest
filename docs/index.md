# Getting Started

After installing quest the minimum configuration you'll need is:
```toml
# ~/.config/quest/config.toml
todo-file = "$HOME/path/to/your/todo.txt"
```
If, you are just getting started with todo.txt and don't have a todo.txt yet you might even skip configuration completely, because in this case the default location of `$HOME/.local/share/quest/todo.txt` is assumed.

With this configuration you can already do basic operations on your todo.txt, like adding tasks `quest add a new task` or
completing a task `quest complete 1`.
In this case *1* refers to the line number of the task you want to complete, but there are many more ways you can 
select tasks here (see [Selecting Tasks](qql.md)).
For a full list of available commands refer to `quest --help`.

However, one of the main features of Quest is that you can define views of your todo.txt inside your config file.
Suppose for example you want a view that displays your "hot" tasks. That is, all tasks that have a priority 
higher than C and are due within the next 3 days. For this you'd put this definition in your config:

```toml
[views.hot]
query = 'priority >= prioD && date(tag("due")) + 3d >= today'
sort = ['-priority']
```
Note, that the there is nothing special about the *due* tag here.
The only thing that makes it special is how it is used in the configuration.
With this configuration in place you can find all your hot tasks by simply typing `quest hot`. 
Not only that, but you can also edit all your hot tasks at once by typing `quest hot edit`.
This will open up your favourite editor (defined by the `EDITOR` environment variable) on a file, which
contains the subset of matching tasks.
There you can make any changes with full support for [Tag Expansion](tags.md), Syntax validation and [Recurrence](recurrence.md).  
After saving and closing this file your changed will be merged back into your main todo.txt.

It might take a while until you've setup all the views you need for a decent workflow.
If you are planning to employ a GTD-like workflow you might want to checkout this [example configuration](https://github.com/Fabian-G/quest/blob/main/examples/gtd/config.toml) which
should get you started quickly. 
