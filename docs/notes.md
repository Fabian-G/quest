# Note Taking

One major limitation of todo.txt is that todo items are always single line items.
And this is fine most of the time, because the todo item should serve as a simple
reminder on what to do and not how to do it.
Sometimes though, we do want to attach more information to a todo item 
(like progress notes, problems encountered...) and this is what the `notes` view command
enables us to do.

Basically the `notes` view command creates a separate markdown file with a unique id and
attaches it to a task by setting a tag to that id.

To start using notes you just have to tell quest which tag to use:

```toml
[notes]
tag = "n"
```

Now you can run `quest notes 42` to open the note file for task 42 in your editor.
The note will be created in `$HOME/.local/share/quest/notes` by default. 
To change this set the `dir` property in the `[notes]` section appropriately.

As always the command takes any type of selectors as described in [Selecting Tasks](selection.md).
However, `notes` enforces that your selection matches a single task.
If it matches multiple task you will be prompted to select a match.

## Id generation

By default `notes` generates 4 character long alphanumeric ids. 
Which should be a good compromise between length of id and amount of available 
ids.
If, for whatever reason, you want to change that default set the `id-length` property
in the `[notes]` section to whatever value you like.

In the unlikely event that you run out of ids you can try running `quest notes clean`, 
which will remove all notes that are not referenced from any item (todo.txt or done.txt). 



