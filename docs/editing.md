# Editing tasks

## Basic Commands

The most basic way of modifying your todo.txt is by using the commands
`add`, `complete`, `prioritize` and `remove`.
With the exception of `add` these commands will take any number of selectors 
(as described in [Selection](selection.md)).
By default these commands are designed with the assumption that your intention
is to operate on a single task. Therefore, if your selection matches multiple 
tasks you will be presented a selection prompt.
This makes it safe to use very broad selectors. 
For example, let's say you just finished doing the laundry.
Now, to mark the corresponding item as completed in Quest, you could run `quest`, figure out the 
line number of the task and the run `quest complete <line>`.
However, we know that the task probably contains the word laundry.
We can therefore just run:
```bash
~ ❯ quest complete laundry
Completed item #113: Do the laundry
```
This is safe, because if there were multiple items containing the word "laundry"
we would have been prompted to select one (or more).

## Managing Projects/Contexts/Tags

The `set` and `unset` commands can be used to add or remove projects/contexts/tags.
Just as the "Basic Commands" these take a list of selectors as arguments, but they
also need a list of projects, contexts or tags to set or unset.
These two lists will be separated by the `on` keyword.
Examples:

```bash
# Sets the do tag to the date of today on tasks 1-4
quest set do:today on 1,2,3,4
# Sets @foo +bar t:tomorrow on tasks 1,2,3,4,6,7, 
# which are also in the baz context
quest set @foo +bar t:tomorrow on 1-4,6-7 @baz
```

The tags/projects/contexts added with the `set` command will always be appended to the
description. If a tag already exists its value will be changed in place.
If a project or context already exists, it will be left untouched.

Note that when unsetting tags you specify only the tag name (not the value):

```bash
quest unset due on 4
```

## Editing Sublists

One powerful feature of Quest is the ability to edit sublists of 
your todo.txt.
Effectively, what this does is:

1. Find tasks that match your query.
2. Write them to a temporary file for you to edit.
3. Wait for you to make your changes.
4. Validate the results.
5. Merge the changes back into the original todo.txt

The edit command fully supports all advanced features of Quest, like recurrence,
tag expansions and validations.
You can also remove tasks by simply removing lines and add tasks by adding lines.
There are only 3 minor things to keep in mind when editing your todos this way:

1. Reordering the tasks in the temporary file will have no effect.
2. Quest will add a special `quest-object-id` tag to the todo items in the temporary file.
    This is used to associate the changed state with the previous states stored in memory to
    trigger recurrence and other change based behaviour correctly. 
    If you remove such a tag, this will be handled like a removal of the old task followed by an
    addition. 
3. The intention of this feature is to make quick edits and then merge them back. 
    You should not keep the temporary file open in the background for long stretches of time,
    because if the todo.txt is written to in the meantime Quest will not
    be able to merge your changes back into the changed file. Your changes will not be lost
    in this case, but you will have to merge them manually.

It is also recommended to use this feature with an editor that supports todo.txt. 
Like neovim with a todo.txt plugin for example.
