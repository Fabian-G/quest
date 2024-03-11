# Views

In Quest there are two types of commands: global commands (e.g. open and init) and view commands (e.g. set and edit).
View commands run in the context of a view and global commands do not.
The purpose of a view is to make it more convenient to work with subsets of your todo.txt
by defining a query, sort order and other options that apply to all view commands that 
run in the context of the view.

For example you might want to consider items with the @inbox context as inbox items.

In that case to list all the inbox items you can run:

```bash
quest -q '!done && @inbox'
```

Note that by default the output will also show a contexts column, which is irrelevant to 
us, because we already know that the items have the @inbox context.
We can therefore specify the projection manually:

```bash
quest -q '!done && @inbox' -p "line,creation,description"
```

Slowly this is becoming infeasible to write everytime to list your inbox.
By defining a view we can make it more convenient:

```toml
[views.inbox]
query = '!done && @inbox'
projection = ["line","creation","description"]
```

With this in your configuration we can now just write `quest inbox` to show the inbox.
Not only that, but we can also use every view command on the subset defined by this view.
A few examples:

```bash
# Opens your inbox items in your editor. 
# This is very convenient to review inbox items 
quest inbox edit 
# Removes the inbox context from all items in your inbox
quest inbox unset @inbox 
# Prioritizes all inbox items as A
quest inbox prio A
```

As you might have guest the general command structure is `quest [view] [view-command]`.
If the view is omitted the view specified by the `[default-view]` section applies.

Although it does not run any queries the `add` command is also a view command.
This is because you can define an `add-prefix` and `add-suffix` in a view.
This is useful, because when running `quest inbox add foo` you would expect 
"foo" to actually show up in your inbox, but it does not, because it does not 
have the appropriate contexts.
With `add-suffix` your can fix this (within limitations) by defining a string
that is appended to every item added through the view.
By setting this to `@inbox` you create the illusion of actually adding to the inbox.

To read about all the available view options checkout the [config reference](configuration.md).
